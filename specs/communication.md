# Kastellan Agent/Operator Communication Specification

## 1. Purpose

This document specifies the communication between the Kastellan Operator and agents running on managed external Linux hosts.

The communication layer is responsible for:

- Authenticating agents
- Associating an agent with an `ExternalHost`
- Reporting host capabilities
- Delivering assigned `PodmanPlay` resources
- Reporting reconciliation results
- Detecting disconnected agents
- Recovering after reconnects
- Preventing duplicate or stale workload application

The agent initiates all connections. External hosts do not require inbound network access or direct access to the Kubernetes API.

---

## 2. Transport

The initial implementation uses:

```text
Bidirectional gRPC streaming over mutual TLS
```

A single long-lived stream carries messages in both directions:

```text
Agent                                      Operator
  |                                           |
  |------------ TLS connection ------------->|
  |------------ AgentHello ----------------->|
  |<----------- OperatorHello ---------------|
  |<----------- DesiredState ----------------|
  |------------ ReconcileResult ------------>|
  |------------ Heartbeat ------------------>|
  |<----------- DesiredState ----------------|
```

The production connection uses mTLS. Unit tests may use an in-memory gRPC listener with generated test certificates or a separate authentication test suite.

The Go gRPC `bufconn` package provides an in-memory full-duplex `net.Conn` and listener suitable for testing gRPC clients and servers without opening a real network port. 【1-337d47】【2-9df083】

---

## 3. Connection Lifecycle

The connection has the following states:

```text
Disconnected
     |
     v
Connecting
     |
     v
Authenticating
     |
     v
Synchronizing
     |
     v
Ready
     |
     +-----------------+
     |                 |
     v                 v
Disconnected        Closing
```

### 3.1 Disconnected

The agent has no connection to the operator.

Existing workloads continue running. The agent must not remove workloads merely because the operator is unavailable.

### 3.2 Connecting

The agent attempts to establish a TLS connection.

Failed attempts use exponential backoff with jitter.

Suggested defaults:

```yaml
initialDelay: 1s
maximumDelay: 60s
multiplier: 2
jitter: 0.2
```

### 3.3 Authenticating

The operator validates the agent's client certificate.

The certificate identity determines the associated `ExternalHost`. The host name transmitted in `AgentHello` is informational and must not override the authenticated identity.

### 3.4 Synchronizing

After the handshake, the operator sends a complete desired-state snapshot.

The agent compares the snapshot with its local state and applies all required changes.

### 3.5 Ready

The agent has successfully processed the latest desired-state revision.

The connection remains active for:

- Heartbeats
- Desired-state updates
- Status updates
- Reconciliation requests
- Graceful shutdown messages

---

## 4. Protocol Service

The initial gRPC service exposes one bidirectional stream.

```protobuf
syntax = "proto3";

package kastellan.agent.v1alpha1;

option go_package = "github.com/example/kastellan/api/proto/agent/v1alpha1";

service AgentService {
  rpc Connect(stream AgentMessage) returns (stream OperatorMessage);
}
```

---

## 5. Common Types

```protobuf
message ProtocolVersion {
  uint32 major = 1;
  uint32 minor = 2;
}

message ResourceIdentity {
  string api_version = 1;
  string kind = 2;
  string namespace = 3;
  string name = 4;
  string uid = 5;
  int64 generation = 6;
}

message ErrorDetail {
  string code = 1;
  string message = 2;
  bool retryable = 3;
}
```

The Kubernetes resource UID is the authoritative workload identity.

The tuple consisting only of namespace and name is insufficient because a deleted and recreated resource may reuse the same name while representing a different object.

---

## 6. Agent Messages

```protobuf
message AgentMessage {
  oneof payload {
    AgentHello hello = 1;
    Heartbeat heartbeat = 2;
    HostInventory inventory = 3;
    ReconcileResult reconcile_result = 4;
    AgentShutdown shutdown = 5;
  }
}
```

### 6.1 AgentHello

`AgentHello` must be the first application-level message sent over a new stream.

```protobuf
message AgentHello {
  string agent_id = 1;
  string host_name = 2;
  string agent_version = 3;

  repeated ProtocolVersion supported_protocols = 4;

  RuntimeInformation runtime = 5;

  string last_session_id = 6;
  uint64 last_received_revision = 7;
  uint64 last_applied_revision = 8;
}

message RuntimeInformation {
  string name = 1;
  string version = 2;
  repeated string capabilities = 3;
}
```

Example:

```yaml
agentID: agent-2fd317
hostName: lb01
agentVersion: 0.1.0

supportedProtocols:
  - major: 1
    minor: 0

runtime:
  name: podman
  version: 5.6.0
  capabilities:
    - play-kube
    - replace
    - configmap
    - secret
    - host-path

lastAppliedRevision: 41
```

### Validation

The operator must reject the message if:

- It is not the first message.
- The agent identity does not match the authenticated host.
- No compatible protocol version exists.
- The agent version is prohibited.
- Podman is unavailable or unsupported.
- Another active session exists and cannot be replaced safely.

---

### 6.2 Heartbeat

```protobuf
message Heartbeat {
  string session_id = 1;
  int64 timestamp_unix = 2;
  uint64 last_received_revision = 3;
  uint64 last_applied_revision = 4;

  RuntimeHealth runtime = 5;

  uint32 assigned_workloads = 6;
  uint32 ready_workloads = 7;
  uint32 failed_workloads = 8;
}

message RuntimeHealth {
  bool available = 1;
  string message = 2;
}
```

The heartbeat confirms connection liveness. It does not replace detailed workload status.

Recommended interval:

```text
30 seconds
```

The operator marks the host disconnected if no valid heartbeat or stream activity is observed for:

```text
2 minutes
```

---

### 6.3 HostInventory

```protobuf
message HostInventory {
  string session_id = 1;

  string hostname = 2;
  string operating_system = 3;
  string architecture = 4;

  RuntimeInformation runtime = 5;

  uint64 total_memory_bytes = 6;
  uint32 logical_cpus = 7;

  repeated LocalWorkload workloads = 8;
}

message LocalWorkload {
  ResourceIdentity resource = 1;
  string manifest_digest = 2;
  string runtime_id = 3;
  WorkloadPhase phase = 4;
}
```

Inventory is sent:

- After the handshake
- After an agent restart
- On request from the operator
- Periodically at a low frequency
- When the agent detects unexpected local drift

---

### 6.4 ReconcileResult

```protobuf
message ReconcileResult {
  string session_id = 1;
  uint64 revision = 2;

  repeated WorkloadResult workloads = 3;
}

message WorkloadResult {
  ResourceIdentity resource = 1;
  string manifest_digest = 2;
  WorkloadPhase phase = 3;

  repeated RuntimeResource runtime_resources = 4;

  ErrorDetail error = 5;
}

message RuntimeResource {
  string type = 1;
  string name = 2;
  string id = 3;
  string state = 4;
}

enum WorkloadPhase {
  WORKLOAD_PHASE_UNSPECIFIED = 0;
  WORKLOAD_PHASE_PENDING = 1;
  WORKLOAD_PHASE_APPLYING = 2;
  WORKLOAD_PHASE_READY = 3;
  WORKLOAD_PHASE_FAILED = 4;
  WORKLOAD_PHASE_DELETING = 5;
  WORKLOAD_PHASE_ABSENT = 6;
}
```

A result is sent when:

- A desired-state revision is fully processed.
- A workload changes state.
- A retry succeeds.
- A permanent error occurs.
- Drift is detected and reconciled.

---

## 7. Operator Messages

```protobuf
message OperatorMessage {
  oneof payload {
    OperatorHello hello = 1;
    DesiredState desired_state = 2;
    ReconcileRequest reconcile_request = 3;
    OperatorShutdown shutdown = 4;
  }
}
```

### 7.1 OperatorHello

```protobuf
message OperatorHello {
  string session_id = 1;
  ProtocolVersion selected_protocol = 2;

  uint32 heartbeat_interval_seconds = 3;
  uint32 state_report_interval_seconds = 4;

  int64 server_time_unix = 5;
}
```

The `session_id` is unique for each accepted connection.

The agent must include this ID in all subsequent messages. Messages with an incorrect or expired session ID are ignored or rejected.

---

### 7.2 DesiredState

The MVP uses complete, authoritative snapshots.

```protobuf
message DesiredState {
  string session_id = 1;
  uint64 revision = 2;

  repeated PodmanPlayAssignment podman_plays = 3;
}

message PodmanPlayAssignment {
  ResourceIdentity resource = 1;

  bytes manifest = 2;
  string manifest_digest = 3;

  UpdateStrategy update_strategy = 4;
}

enum UpdateStrategy {
  UPDATE_STRATEGY_UNSPECIFIED = 0;
  UPDATE_STRATEGY_REPLACE = 1;
}
```

Example:

```yaml
sessionID: session-844d12
revision: 42

podmanPlays:
  - resource:
      apiVersion: kastellan.io/v1alpha1
      kind: PodmanPlay
      namespace: infrastructure
      name: haproxy
      uid: 0f831f62-342a-4add-b05c-c968ec71b679
      generation: 4

    manifestDigest: sha256:93f3d2...

    manifest: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: haproxy
      spec:
        containers:
          - name: haproxy
            image: docker.io/library/haproxy:lts
```

### Desired-State Rules

1. Revisions are monotonically increasing per host.
2. Revisions are scoped to an `ExternalHost`.
3. A revision represents the complete set of Kastellan-managed workloads assigned to that host.
4. Receiving the same revision multiple times must be safe.
5. A lower revision than the last applied revision must not be applied.
6. A workload missing from a newer snapshot must be removed if it is owned by Kastellan.
7. Unmanaged Podman resources must never be removed.
8. The operator may resend the current revision after reconnecting.
9. The agent acknowledges a revision through `ReconcileResult`.
10. A revision is considered applied only after all workload operations have reached a terminal result for that attempt.

A failed workload does not block status reporting for successfully reconciled workloads in the same revision.

---

### 7.3 ReconcileRequest

```protobuf
message ReconcileRequest {
  string session_id = 1;

  enum Scope {
    SCOPE_UNSPECIFIED = 0;
    SCOPE_ALL = 1;
    SCOPE_WORKLOAD = 2;
    SCOPE_INVENTORY = 3;
  }

  Scope scope = 2;
  ResourceIdentity resource = 3;
}
```

This message requests another local inspection. It does not carry an imperative Podman command.

---

## 8. Handshake Sequence

```text
Agent                                        Operator
  |                                             |
  |---------- Establish mTLS ----------------->|
  |                                             |
  |---------- AgentHello --------------------->|
  |                                             |
  |                  Validate certificate       |
  |                  Resolve ExternalHost       |
  |                  Negotiate protocol         |
  |                  Create session             |
  |                                             |
  |<--------- OperatorHello -------------------|
  |                                             |
  |---------- HostInventory ------------------>|
  |                                             |
  |<--------- DesiredState revision 42 --------|
  |                                             |
  |          Validate full snapshot             |
  |          Reconcile local workloads          |
  |                                             |
  |---------- ReconcileResult revision 42 ---->|
  |                                             |
  |<-------- Optional ReconcileRequest --------|
  |                                             |
  |<---------> Periodic heartbeats <---------->|
```

The operator must not send desired state before completing the handshake.

The agent must not process desired-state messages whose session ID differs from the negotiated session.

---

## 9. Reconnection Sequence

```text
Agent                                         Operator
  |                                              |
  |             Connection lost                  |
  |                                              |
  |       Existing workloads continue            |
  |                                              |
  |---------- New mTLS connection -------------->|
  |---------- AgentHello ----------------------->|
  |           lastAppliedRevision: 42             |
  |                                              |
  |<--------- OperatorHello --------------------|
  |<--------- DesiredState revision 44 ---------|
  |                                              |
  |          Reconcile revision 44                |
  |                                              |
  |---------- ReconcileResult revision 44 ------>|
```

The agent must not assume that revisions missed while disconnected will be replayed individually.

The operator sends the latest complete snapshot.

---

## 10. Operator Status Updates

The operator updates `ExternalHost.status` from connection and heartbeat information.

```yaml
status:
  agentID: agent-2fd317
  agentVersion: 0.1.0
  lastSeen: "2026-08-31T11:00:00Z"

  runtime:
    name: podman
    version: 5.6.0

  conditions:
    - type: Connected
      status: "True"
      reason: AgentConnected

    - type: Ready
      status: "True"
      reason: RuntimeAvailable
```

The operator updates `PodmanPlay.status` from reconciliation results.

```yaml
status:
  observedGeneration: 4

  hosts:
    - name: lb01
      phase: Ready
      appliedGeneration: 4
      revision: 42
      manifestDigest: sha256:93f3d2...

  conditions:
    - type: Available
      status: "True"
      reason: ReadyOnAllTargetHosts
```

The status model should use Kubernetes-style conditions and `observedGeneration`, allowing clients to distinguish status belonging to the current resource generation from stale status. 【3-a0cfa5】【4-87abe4】

---

# Test Specification

## 11. Test Strategy

Communication should be tested at three levels:

```text
Unit tests
    ↓
In-process protocol tests
    ↓
Integration tests with Kubernetes API and fake Podman
```

The tests should not require Podman for most protocol and controller scenarios.

Use a runtime interface so the agent can run against a fake implementation:

```go
type Runtime interface {
    Inspect(
        ctx context.Context,
    ) ([]ObservedWorkload, error)

    Apply(
        ctx context.Context,
        workload DesiredWorkload,
    ) (ObservedWorkload, error)

    Remove(
        ctx context.Context,
        identity ResourceIdentity,
    ) error

    Capabilities(
        ctx context.Context,
    ) (RuntimeCapabilities, error)
}
```

A real Podman adapter implements the interface for production.

A fake adapter records invocations and exposes deterministic results in tests.

---

# 12. Unit Tests

## 12.1 Handshake Acceptance

### Purpose

Verify that a valid agent establishes a session and receives desired state.

### Given

- An authenticated identity associated with `ExternalHost/lb01`
- Agent supports protocol `1.0`
- Operator supports protocol `1.0`
- One `PodmanPlay` is assigned to `lb01`

### When

- Agent opens the stream
- Agent sends `AgentHello`

### Then

- Operator returns `OperatorHello`
- A non-empty session ID is assigned
- Protocol `1.0` is selected
- Operator sends a `DesiredState`
- Desired state contains the assigned `PodmanPlay`

---

## 12.2 Reject Host Identity Mismatch

### Given

- Authenticated certificate identity belongs to `lb01`
- `AgentHello.host_name` is `lb02`

### When

The agent opens a connection.

### Then

- Connection is rejected
- No desired state is sent
- `lb02` status remains unchanged
- A security-relevant log entry is created

---

## 12.3 Reject Incompatible Protocol

### Given

- Operator supports protocol `1.0`
- Agent supports only protocol `2.0`

### Then

- Operator closes the stream with `FailedPrecondition`
- No session is registered
- No desired state is sent

---

## 12.4 Reject Message Before AgentHello

### Given

A new stream without a completed handshake.

### When

The first message is `Heartbeat`.

### Then

- Stream is closed with `FailedPrecondition`
- No agent session is created

---

## 12.5 Idempotent Revision

### Given

- Revision `42` contains one workload
- Fake runtime reports no existing workload

### When

The agent receives revision `42` twice.

### Then

- `Apply` is invoked exactly once
- No duplicate workload is created
- Two acknowledgements may be sent
- Both acknowledgements report revision `42`

---

## 12.6 Reject Stale Revision

### Given

The agent has successfully applied revision `42`.

### When

It receives revision `41`.

### Then

- No runtime operation is invoked
- Last applied revision remains `42`
- Agent reports a stale-revision error or ignores the message with a warning

---

## 12.7 Replace Changed Workload

### Given

- Workload UID remains unchanged
- Existing manifest digest is `sha256:old`
- Desired manifest digest is `sha256:new`

### When

A new revision is received.

### Then

- Fake runtime receives one `Apply`
- Update strategy is `Replace`
- Reconciliation result contains the new digest
- Workload phase becomes `Ready`

---

## 12.8 Delete Removed Workload

### Given

- Revision `42` assigns workloads A and B
- Revision `43` assigns only workload A
- Workload B is marked as managed by Kastellan

### When

Revision `43` is applied.

### Then

- Fake runtime removes workload B
- Workload A is not unnecessarily recreated
- Result for B reports `Absent`
- Last applied revision becomes `43`

---

## 12.9 Preserve Unmanaged Workloads

### Given

Local inventory contains:

- One Kastellan-managed workload
- One manually created Podman workload

### When

The desired state is empty.

### Then

- Managed workload is removed
- Manual workload is untouched

---

## 12.10 Continue During Disconnect

### Given

A workload is running.

### When

The gRPC connection is closed.

### Then

- Agent does not invoke `Remove`
- Existing workload remains running
- Agent enters reconnect mode

---

# 13. In-Memory Communication Test

This verifies the complete stream exchange without opening a network port.

## Test Name

```text
TestAgentConnectsReceivesStateAndReportsReady
```

## Test Components

```text
In-memory gRPC listener
        |
        +-- Real gRPC operator server
        |
        +-- Real agent connection loop
        |
        +-- Fake Podman runtime
        |
        +-- Fake assignment source
```

The `bufconn.Listener` creates a buffered in-memory full-duplex connection between the real gRPC client and server, making it suitable for testing the generated protocol and stream behavior. 【1-337d47】【2-9df083】

## Test Flow

1. Start an in-memory gRPC server.
2. Register the real operator `AgentService`.
3. Configure authenticated host identity as `lb01`.
4. Configure one desired `PodmanPlay`.
5. Start the real agent connection loop.
6. Agent sends `AgentHello`.
7. Server sends `OperatorHello`.
8. Server sends desired-state revision `1`.
9. Agent calls fake runtime `Apply`.
10. Fake runtime returns `Ready`.
11. Agent sends `ReconcileResult`.
12. Test verifies the result.

## Example Go Test Skeleton

```go
func TestAgentConnectsReceivesStateAndReportsReady(t *testing.T) {
    t.Parallel()

    ctx, cancel := context.WithTimeout(
        context.Background(),
        5*time.Second,
    )
    defer cancel()

    listener := bufconn.Listen(1024 * 1024)

    assignments := NewFakeAssignmentSource()
    assignments.Set("lb01", DesiredState{
        Revision: 1,
        Workloads: []DesiredWorkload{
            {
                Identity: ResourceIdentity{
                    APIVersion: "kastellan.io/v1alpha1",
                    Kind:       "PodmanPlay",
                    Namespace:  "infrastructure",
                    Name:       "haproxy",
                    UID:        "test-uid",
                    Generation: 1,
                },
                Manifest: []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: haproxy
spec:
  containers:
    - name: haproxy
      image: docker.io/library/haproxy:lts
`),
                ManifestDigest: "sha256:test",
            },
        },
    })

    results := make(chan ReconcileResult, 1)

    gateway := NewGateway(GatewayOptions{
        Assignments: assignments,
        IdentityResolver: StaticIdentityResolver{
            HostName: "lb01",
        },
        ResultHandler: func(
            ctx context.Context,
            result ReconcileResult,
        ) error {
            results <- result
            return nil
        },
    })

    grpcServer := grpc.NewServer()
    agentv1alpha1.RegisterAgentServiceServer(grpcServer, gateway)

    go func() {
        _ = grpcServer.Serve(listener)
    }()

    t.Cleanup(func() {
        grpcServer.Stop()
        _ = listener.Close()
    })

    connection, err := grpc.NewClient(
        "passthrough:///bufnet",
        grpc.WithContextDialer(
            func(
                ctx context.Context,
                _ string,
            ) (net.Conn, error) {
                return listener.DialContext(ctx)
            },
        ),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    require.NoError(t, err)
    t.Cleanup(func() {
        _ = connection.Close()
    })

    runtime := NewFakeRuntime()
    runtime.ApplyResult = ObservedWorkload{
        Identity: ResourceIdentity{
            UID: "test-uid",
        },
        Phase: WorkloadPhaseReady,
        RuntimeResources: []RuntimeResource{
            {
                Type:  "pod",
                Name:  "haproxy",
                ID:    "pod-123",
                State: "running",
            },
        },
    }

    kastellanAgent := NewAgent(AgentOptions{
        HostName: "lb01",
        AgentID:  "agent-test",
        Runtime:  runtime,
        Client:   agentv1alpha1.NewAgentServiceClient(connection),
    })

    agentErrors := make(chan error, 1)

    go func() {
        agentErrors <- kastellanAgent.Run(ctx)
    }()

    select {
    case result := <-results:
        require.Equal(t, uint64(1), result.Revision)
        require.Len(t, result.Workloads, 1)
        assert.Equal(
            t,
            WorkloadPhaseReady,
            result.Workloads[0].Phase,
        )
        assert.Equal(t, "test-uid", result.Workloads[0].Identity.UID)

        calls := runtime.ApplyCalls()
        require.Len(t, calls, 1)
        assert.Equal(t, "haproxy", calls[0].Identity.Name)

    case err := <-agentErrors:
        require.NoError(t, err)
        t.Fatal("agent stopped before reporting reconciliation result")

    case <-ctx.Done():
        t.Fatal("timed out waiting for reconciliation result")
    }
}
```

The in-memory test may use insecure transport because it tests message flow rather than certificate validation. mTLS should have a dedicated transport test.

---

# 14. mTLS Authentication Test

## Test Name

```text
TestGatewayRejectsAgentCertificateForDifferentHost
```

## Setup

Generate a test CA and certificates during the test:

```text
Test CA
 ├─ Operator server certificate
 ├─ Agent certificate for lb01
 └─ Agent certificate for lb02
```

Certificate identity example:

```text
URI SAN:
spiffe://kastellan.test/cluster/test/hosts/lb01
```

## Flow

1. Start the gRPC gateway on `127.0.0.1` with the test server certificate.
2. Require and verify client certificates.
3. Connect using the `lb01` certificate.
4. Send `AgentHello` for `lb02`.
5. Verify rejection.
6. Reconnect using the `lb01` certificate.
7. Send `AgentHello` for `lb01`.
8. Verify successful handshake.

## Assertions

```text
Mismatched identity:
  gRPC code = PermissionDenied
  session created = false
  desired state sent = false

Matching identity:
  OperatorHello received = true
  authenticated host = lb01
```

---

# 15. Kubernetes Integration Test

The integration test verifies the complete path:

```text
Kubernetes resources
        |
        v
Operator reconciliation
        |
        v
Agent gRPC stream
        |
        v
Fake Podman runtime
        |
        v
PodmanPlay status
```

Use `controller-runtime/pkg/envtest` to start a local API server and etcd, install the Kastellan CRDs, and run the actual controllers without requiring a full Kubernetes cluster. 【5-20eb46】【6-f26bbd】

## Test Name

```text
TestPodmanPlayIsDeliveredToAllHostsAndStatusBecomesReady
```

## Test Setup

Create:

```yaml
apiVersion: kastellan.io/v1alpha1
kind: ExternalHost
metadata:
  name: lb01
  labels:
    role: load-balancer
spec:
  enabled: true
---
apiVersion: kastellan.io/v1alpha1
kind: ExternalHostGroup
metadata:
  name: production-lb
spec:
  selector:
    matchLabels:
      role: load-balancer
---
apiVersion: kastellan.io/v1alpha1
kind: PodmanPlay
metadata:
  name: haproxy
  namespace: infrastructure
spec:
  hostGroupRef:
    name: production-lb
  deploymentPolicy:
    mode: All
  workload: |
    apiVersion: v1
    kind: Pod
    metadata:
      name: haproxy
    spec:
      containers:
        - name: haproxy
          image: docker.io/library/haproxy:lts
```

## Test Flow

1. Start `envtest`.
2. Install Kastellan CRDs.
3. Start the real controller manager.
4. Start the real gRPC gateway.
5. Create `ExternalHost/lb01`.
6. Start an agent for `lb01` with a fake runtime.
7. Wait until `ExternalHost/lb01` becomes `Connected=True`.
8. Create `ExternalHostGroup/production-lb`.
9. Create `PodmanPlay/infrastructure/haproxy`.
10. Wait until the fake runtime receives `Apply`.
11. Fake runtime reports `Ready`.
12. Wait until `PodmanPlay.status` reports host `lb01` as ready.
13. Delete the `PodmanPlay`.
14. Wait until fake runtime receives `Remove`.
15. Verify the resource is deleted after finalizer cleanup.

## Assertions

```text
ExternalHost:
  Connected = True
  Ready = True

Fake runtime:
  Apply calls = 1
  Applied workload UID matches PodmanPlay UID

PodmanPlay:
  observedGeneration matches metadata.generation
  lb01 phase = Ready
  Available condition = True

After deletion:
  Remove calls = 1
  PodmanPlay no longer exists
```

## Go Test Skeleton

```go
func TestPodmanPlayIsDeliveredAndBecomesReady(t *testing.T) {
    ctx, cancel := context.WithTimeout(
        context.Background(),
        30*time.Second,
    )
    defer cancel()

    testEnvironment := &envtest.Environment{
        CRDDirectoryPaths: []string{
            filepath.Join("..", "..", "config", "crd", "bases"),
        },
        ErrorIfCRDPathMissing: true,
    }

    restConfig, err := testEnvironment.Start()
    require.NoError(t, err)

    t.Cleanup(func() {
        require.NoError(t, testEnvironment.Stop())
    })

    scheme := runtime.NewScheme()
    require.NoError(t, clientgoscheme.AddToScheme(scheme))
    require.NoError(t, kastellanv1alpha1.AddToScheme(scheme))

    manager, err := ctrl.NewManager(
        restConfig,
        ctrl.Options{
            Scheme: scheme,
        },
    )
    require.NoError(t, err)

    gateway := NewGateway(/* dependencies */)

    require.NoError(
        t,
        SetupControllers(manager, gateway),
    )

    go func() {
        _ = manager.Start(ctx)
    }()

    kubernetesClient := manager.GetClient()

    host := &kastellanv1alpha1.ExternalHost{
        ObjectMeta: metav1.ObjectMeta{
            Name: "lb01",
            Labels: map[string]string{
                "role": "load-balancer",
            },
        },
        Spec: kastellanv1alpha1.ExternalHostSpec{
            Enabled: true,
        },
    }

    require.NoError(
        t,
        kubernetesClient.Create(ctx, host),
    )

    fakeRuntime := NewFakeRuntime()

    testAgent := NewAgent(AgentOptions{
        HostName: "lb01",
        AgentID:  "agent-lb01",
        Runtime:  fakeRuntime,
        Client:   newTestGatewayClient(t, gateway),
    })

    go func() {
        _ = testAgent.Run(ctx)
    }()

    require.EventuallyWithT(
        t,
        func(collect *assert.CollectT) {
            current := &kastellanv1alpha1.ExternalHost{}

            err := kubernetesClient.Get(
                ctx,
                client.ObjectKey{Name: "lb01"},
                current,
            )
            assert.NoError(collect, err)
            assertCondition(
                collect,
                current.Status.Conditions,
                "Connected",
                metav1.ConditionTrue,
            )
        },
        10*time.Second,
        100*time.Millisecond,
    )

    group := newProductionLBGroup()
    require.NoError(
        t,
        kubernetesClient.Create(ctx, group),
    )

    play := newHAProxyPodmanPlay()
    require.NoError(
        t,
        kubernetesClient.Create(ctx, play),
    )

    require.EventuallyWithT(
        t,
        func(collect *assert.CollectT) {
            calls := fakeRuntime.ApplyCalls()

            if !assert.Len(collect, calls, 1) {
                return
            }

            assert.Equal(
                collect,
                "haproxy",
                calls[0].Identity.Name,
            )
        },
        10*time.Second,
        100*time.Millisecond,
    )

    require.EventuallyWithT(
        t,
        func(collect *assert.CollectT) {
            current := &kastellanv1alpha1.PodmanPlay{}

            err := kubernetesClient.Get(
                ctx,
                client.ObjectKey{
                    Namespace: "infrastructure",
                    Name:      "haproxy",
                },
                current,
            )
            assert.NoError(collect, err)

            assert.Equal(
                collect,
                current.Generation,
                current.Status.ObservedGeneration,
            )

            assertHostReady(
                collect,
                current.Status.Hosts,
                "lb01",
            )
        },
        10*time.Second,
        100*time.Millisecond,
    )
}
```

---

# 16. Reconnection Integration Test

## Test Name

```text
TestAgentReconnectsAndReceivesLatestSnapshot
```

## Flow

1. Agent connects and applies revision `1`.
2. Stop the gRPC server.
3. Verify running workloads are not removed.
4. Create or update a `PodmanPlay`.
5. The operator generates revision `2`.
6. Restart the gRPC server.
7. Agent reconnects.
8. Agent reports `lastAppliedRevision=1`.
9. Operator sends the complete revision `2` snapshot.
10. Agent applies the changed workload.
11. Status becomes ready at revision `2`.

## Assertions

```text
During disconnect:
  Remove calls = 0

After reconnect:
  Apply calls = 2
  Last applied revision = 2
  No duplicate workload created
```

---

# 17. Minimum Acceptance Test

The implementation is considered functionally complete for the first MVP when this scenario passes:

```text
Given:
  One ExternalHost
  One ExternalHostGroup selecting that host
  One PodmanPlay assigned to the group

When:
  The agent connects to the operator

Then:
  The operator authenticates the agent
  The operator sends the complete desired state
  The agent applies the workload exactly once
  The agent reports the workload as ready
  The operator updates PodmanPlay status

When:
  The connection is interrupted

Then:
  The running workload remains untouched

When:
  The agent reconnects

Then:
  The latest desired-state snapshot is synchronized
  No duplicate workload is created

When:
  The PodmanPlay is deleted

Then:
  The agent removes the managed workload
  The operator removes the finalizer
  The Kubernetes resource is deleted
```

---

# 18. Recommended Test Layout

```text
internal/
├── agent/
│   ├── agent.go
│   ├── reconcile.go
│   ├── reconcile_test.go
│   └── connection_test.go
│
├── gateway/
│   ├── server.go
│   ├── session.go
│   ├── handshake_test.go
│   └── stream_test.go
│
├── runtime/
│   ├── runtime.go
│   ├── fake/
│   │   └── runtime.go
│   └── podman/
│       └── runtime.go
│
└── integration/
    ├── podmanplay_delivery_test.go
    ├── reconnect_test.go
    ├── deletion_test.go
    └── mtls_test.go

proto/
└── agent/
    └── v1alpha1/
        └── agent.proto
```

---

# 19. Recommended Implementation Order

1. Define the Protocol Buffer contract.
2. Implement the fake runtime.
3. Implement the gateway handshake.
4. Implement the agent connection loop.
5. Add the in-memory gRPC test.
6. Implement desired-state snapshots.
7. Implement agent-side reconciliation.
8. Add idempotency and stale-revision tests.
9. Connect the gateway to the Kubernetes controllers.
10. Add the `envtest` integration test.
11. Add mTLS identity validation.
12. Add reconnect and deletion tests.