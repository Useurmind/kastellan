# Kastellan Agent Design Draft

## Purpose

The Kastellan Agent runs on every external Linux host managed by Kastellan.

It connects to the Kastellan Operator and reconciles workloads assigned to its host. The agent uses the local Podman installation to deploy resources defined through `PodmanPlay`.

The external host does not require direct access to the Kubernetes API. Instead, the agent establishes an outbound connection to the operator and receives only the desired state relevant to that host.

---

## High-Level Architecture

```text
+----------------------------+
| Kubernetes Cluster         |
|                            |
|  +----------------------+  |
|  | Kastellan Operator   |  |
|  |                      |  |
|  | - Watches resources  |  |
|  | - Resolves groups    |  |
|  | - Assigns workloads  |  |
|  | - Maintains status   |  |
|  +----------+-----------+  |
+-------------|--------------+
              ^
              |
              | Long-lived outbound connection
              | gRPC over mTLS
              |
+-------------|--------------+
| External Linux Host        |
|             |              |
|  +----------+-----------+  |
|  | Kastellan Agent      |  |
|  |                      |  |
|  | - Receives state     |  |
|  | - Runs Podman        |  |
|  | - Observes workloads |  |
|  | - Reports status     |  |
|  +----------+-----------+  |
|             |              |
|             v              |
|       Podman Runtime       |
+----------------------------+
```

---

# Responsibilities

## Operator Responsibilities

The Kastellan Operator is responsible for:

- Watching `ExternalHost`, `ExternalHostGroup` and `PodmanPlay` resources.
- Determining which workloads are assigned to each host.
- Validating resource references.
- Generating a desired-state revision for each host.
- Delivering desired state to connected agents.
- Receiving host and workload observations.
- Updating Kubernetes resource status.
- Coordinating workload deletion through finalizers.
- Authenticating and authorizing agents.

The operator remains the only component that accesses the Kubernetes API.

## Agent Responsibilities

The Kastellan Agent is responsible for:

- Enrolling the external host.
- Establishing an outbound connection to the operator.
- Authenticating the operator.
- Authenticating itself to the operator.
- Maintaining the control connection.
- Receiving the complete desired state for its host.
- Reconciling assigned `PodmanPlay` resources.
- Validating workload definitions before applying them.
- Invoking Podman.
- Detecting local workload state and configuration drift.
- Reporting host capabilities and workload status.
- Removing workloads that are no longer assigned.
- Continuing to run existing workloads while disconnected.

## Agent Non-Responsibilities

The agent is not responsible for:

- Connecting directly to the Kubernetes API.
- Selecting workloads or host groups.
- Scheduling workloads to other hosts.
- Interpreting Kubernetes RBAC.
- Accepting inbound management connections.
- Providing unrestricted shell access.
- Executing arbitrary commands received from users.
- Managing unrelated Podman containers.
- Managing the Linux operating system.

---

# Agent Deployment

The agent runs as a Podman container on the managed host.

A simplified deployment could look like:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: kastellan-agent
spec:
  hostNetwork: true

  containers:
    - name: agent
      image: ghcr.io/example/kastellan-agent:v0.1.0

      env:
        - name: KASTELLAN_OPERATOR_ADDRESS
          value: kastellan.example.internal:443

        - name: KASTELLAN_HOST_NAME
          value: lb01

      volumeMounts:
        - name: podman-socket
          mountPath: /run/podman/podman.sock

        - name: agent-state
          mountPath: /var/lib/kastellan

        - name: agent-certs
          mountPath: /etc/kastellan/certs
          readOnly: true

  volumes:
    - name: podman-socket
      hostPath:
        path: /run/podman/podman.sock
        type: Socket

    - name: agent-state
      hostPath:
        path: /var/lib/kastellan
        type: DirectoryOrCreate

    - name: agent-certs
      hostPath:
        path: /etc/kastellan/certs
        type: Directory
```

The exact Podman socket path depends on whether the agent operates Podman in rootful or rootless mode.

The initial implementation should choose one supported mode explicitly rather than attempting to support both immediately.

---

# Connection Model

## Connection Direction

The agent always initiates the connection:

```text
Kastellan Agent ---> Kastellan Operator
```

The operator never initiates a connection to the external host.

This allows hosts to operate behind:

- Firewalls
- NAT
- Customer network boundaries
- Outbound-only security zones
- Dynamic IP addresses

Only outbound access to the operator endpoint is required.

## Transport

The initial communication protocol should use:

```text
Bidirectional gRPC streaming over mutual TLS
```

A single long-lived connection is used for:

- Agent registration
- Heartbeats
- Desired-state delivery
- Workload status
- Host inventory
- Reconciliation results

The default endpoint should use TCP port `443` wherever possible.

## Conceptual Protocol

```protobuf
service AgentService {
  rpc Connect(stream AgentMessage) returns (stream OperatorMessage);
}
```

The stream is bidirectional:

```text
Agent                                  Operator
  |                                       |
  |--------- Connect -------------------->|
  |--------- AgentHello ----------------->|
  |<-------- OperatorHello ---------------|
  |<-------- DesiredState ----------------|
  |--------- WorkloadStatus ------------->|
  |<-------- DesiredStateUpdate ----------|
  |--------- Heartbeat ------------------>|
  |--------- HostStatus ----------------->|
```

---

# Agent Identity and Enrollment

## Host Preparation

Before starting an agent, an administrator creates an `ExternalHost` resource.

```yaml
apiVersion: kastellan.io/v1alpha1
kind: ExternalHost
metadata:
  name: lb01
spec:
  enabled: true
```

The host does not become active until an agent has enrolled successfully.

## Enrollment Process

The initial enrollment flow is:

1. Administrator creates an `ExternalHost`.
2. Administrator generates a short-lived enrollment token for that host.
3. The token is transferred securely to the external host.
4. The agent starts with the enrollment token.
5. The agent connects to the operator over TLS.
6. The operator validates the token.
7. The operator binds the agent to the `ExternalHost`.
8. The operator issues an agent identity and client certificate.
9. The agent stores its identity locally.
10. The one-time enrollment token is invalidated.
11. The agent reconnects using mutual TLS.

```text
ExternalHost: lb01
       |
       v
One-time enrollment token
       |
       v
Agent connects and enrolls
       |
       v
Client certificate issued
       |
       v
Token invalidated
```

## Agent Identity

Each certificate must identify exactly one `ExternalHost`.

The identity should include:

- Kubernetes cluster identity
- External host identity
- Agent identity
- Certificate expiration
- Allowed usage

An agent must not be able to select its identity simply by sending another host name during connection setup.

The operator derives the authorized host from the authenticated certificate.

## Certificate Rotation

Agent certificates should be short-lived and rotated automatically.

The agent should request rotation before expiration through the existing authenticated connection.

If the certificate expires before rotation succeeds, administrative recovery or a new enrollment token is required.

---

# Connection Lifecycle

## Initial Connection

When establishing a connection, the agent sends an initial message:

```yaml
type: AgentHello

agent:
  id: agent-2fd317
  version: 0.1.0

host:
  name: lb01
  hostname: lb01.example.internal

protocolVersions:
  - v1alpha1

runtime:
  name: podman
  version: 5.6.0

capabilities:
  - play-kube
  - replace
  - configmap
  - secret
  - host-path
```

The operator responds with:

```yaml
type: OperatorHello

session:
  id: session-844d12

protocolVersion: v1alpha1

configuration:
  heartbeatInterval: 30s
  stateReportInterval: 60s
```

## Heartbeats

The agent sends periodic heartbeats:

```yaml
type: Heartbeat

sessionID: session-844d12
timestamp: "2026-08-31T10:30:00Z"

runtime:
  available: true

workloads:
  assigned: 3
  ready: 3
  failed: 0
```

The operator uses heartbeats to update `ExternalHost.status`.

Example:

```yaml
status:
  connected: true
  lastSeen: "2026-08-31T10:30:00Z"

  conditions:
    - type: Connected
      status: "True"
      reason: AgentHeartbeatReceived

    - type: Ready
      status: "True"
      reason: PodmanAvailable
```

## Connection Loss

If the connection is interrupted:

- Existing workloads continue running.
- The agent does not remove workloads solely because the operator is unreachable.
- The agent periodically attempts to reconnect.
- Reconnection uses exponential backoff with jitter.
- The operator marks the host as disconnected after a configurable timeout.
- No new desired state is applied until synchronization succeeds.

Recommended initial defaults:

```yaml
heartbeatInterval: 30s
offlineAfter: 2m
reconnect:
  initialDelay: 1s
  maximumDelay: 1m
```

---

# Desired-State Model

The operator sends the complete workload assignment for a host.

```yaml
type: DesiredState

host: lb01
revision: 42

podmanPlays:
  - uid: 0f831f62-342a-4add-b05c-c968ec71b679
    namespace: infrastructure
    name: haproxy
    generation: 4
    manifest: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: haproxy
      spec:
        containers:
          - name: haproxy
            image: docker.io/library/haproxy:lts

  - uid: 6dd6acc8-06a4-45cd-8a52-4b147521137c
    namespace: infrastructure
    name: node-exporter
    generation: 2
    manifest: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: node-exporter
      spec:
        containers:
          - name: node-exporter
            image: quay.io/prometheus/node-exporter:latest
```

The complete desired-state snapshot is authoritative for resources managed by Kastellan on that host.

A workload that exists locally but is not present in the desired-state snapshot is removed only if it is clearly marked as managed by Kastellan.

Unmanaged Podman pods and containers must never be modified.

---

# Revision Model

Every desired-state assignment has a monotonically increasing revision per host.

```text
Host lb01:
  Revision 40
  Revision 41
  Revision 42
```

The agent records:

- Last received revision
- Last successfully applied revision
- Per-workload generation
- Per-workload manifest digest

Each desired-state message includes a unique revision.

The agent must handle receiving the same revision more than once. Applying the same desired state repeatedly must not create duplicate workloads.

## Full Snapshots

The MVP should use full desired-state snapshots rather than incremental commands.

Advantages:

- Simpler recovery
- Easier reconciliation
- No dependency on missed messages
- Less protocol complexity
- Easier debugging

Incremental updates can be introduced later as an optimization.

---

# Agent Reconciliation

For each desired-state revision, the agent performs the following steps:

1. Validate the host identity.
2. Validate the desired-state revision.
3. Parse each `PodmanPlay` manifest.
4. Verify that only supported Kubernetes resource types are used.
5. Apply local security policy.
6. Calculate a digest of each normalized manifest.
7. Inspect existing Podman resources.
8. Compare desired and observed state.
9. Create, replace or remove managed workloads.
10. Observe workload health.
11. Store the successfully applied revision.
12. Report the result to the operator.

```text
Desired State
      |
      v
Manifest Validation
      |
      v
Security Validation
      |
      v
Compare with Local State
      |
      +--> No change
      |
      +--> Create workload
      |
      +--> Replace workload
      |
      +--> Remove workload
      |
      v
Observe State
      |
      v
Report Result
```

---

# PodmanPlay Lifecycle

## Create

For a new workload, the agent:

1. Writes the manifest into the workload state directory.
2. Validates the manifest.
3. Runs `podman kube play`.
4. Inspects the created Podman resources.
5. Records their runtime identifiers.
6. Reports the observed state.

Conceptually:

```bash
podman kube play workload.yaml
```

## Update

If the manifest digest or generation changes, the initial update strategy is replacement.

Conceptually:

```bash
podman kube play --replace workload.yaml
```

The agent reports the workload as `Updating` until the replacement has completed successfully.

## Delete

When a previously managed workload is absent from the desired-state snapshot, the agent removes it.

Conceptually:

```bash
podman kube down workload.yaml
```

Deletion applies only to resources owned by the corresponding `PodmanPlay` UID.

## Drift

If somebody manually deletes a managed Podman pod, the agent recreates it during the next reconciliation.

If somebody changes a managed workload outside Kastellan, the agent restores the declared state where that drift can be detected reliably.

The initial drift policy is:

```text
Reconcile
```

---

# Local Ownership

Every deployed workload must be identifiable as belonging to Kastellan.

Recommended annotations and labels:

```yaml
metadata:
  labels:
    kastellan.io/managed: "true"
    kastellan.io/namespace: infrastructure
    kastellan.io/workload: haproxy
    kastellan.io/workload-uid: 0f831f62-342a-4add-b05c-c968ec71b679
    kastellan.io/generation: "4"
```

The Kubernetes object UID is the primary workload identity.

A resource deleted and recreated with the same namespace and name receives a new UID and must be treated as a new workload.

The agent must not rely solely on the workload name.

---

# Local Agent State

The agent stores its persistent state under:

```text
/var/lib/kastellan/
```

Suggested structure:

```text
/var/lib/kastellan/
├── identity/
│   ├── agent-id
│   ├── certificate.pem
│   ├── private-key.pem
│   └── ca-bundle.pem
│
├── state/
│   ├── last-received-revision
│   └── last-applied-revision
│
└── workloads/
    └── <workload-uid>/
        ├── manifest.yaml
        ├── manifest.sha256
        ├── metadata.json
        └── status.json
```

This state is used for recovery and reconciliation but is not the authoritative desired state.

The agent should be able to reconstruct most local state by inspecting Podman resources and their Kastellan ownership labels.

---

# Status Reporting

After reconciliation, the agent sends a result for each workload.

```yaml
type: ReconciliationResult

host: lb01
revision: 42

workloads:
  - uid: 0f831f62-342a-4add-b05c-c968ec71b679
    namespace: infrastructure
    name: haproxy
    generation: 4
    phase: Ready
    manifestDigest: sha256:93f3d2...
    runtime:
      podID: 12d769d4...
      containers:
        - name: haproxy
          id: fdd471ee...
          state: running

  - uid: 6dd6acc8-06a4-45cd-8a52-4b147521137c
    namespace: infrastructure
    name: node-exporter
    generation: 2
    phase: Failed
    reason: ImagePullFailed
    message: Unable to pull the requested image
```

The operator translates this into `PodmanPlay.status`.

```yaml
status:
  observedGeneration: 4
  phase: Ready

  hosts:
    - name: lb01
      phase: Ready
      appliedGeneration: 4
      lastTransitionTime: "2026-08-31T10:35:00Z"

    - name: lb02
      phase: Ready
      appliedGeneration: 4
      lastTransitionTime: "2026-08-31T10:35:02Z"

  conditions:
    - type: Available
      status: "True"
      reason: WorkloadReadyOnAllHosts
```

---

# Error Handling

Errors should be classified as either transient or permanent.

## Transient Errors

Examples:

- Registry temporarily unavailable
- Podman temporarily unavailable
- Network timeout
- Image pull timeout
- Operator connection interrupted

The agent retries transient errors using exponential backoff.

## Permanent Errors

Examples:

- Invalid Kubernetes YAML
- Unsupported resource type
- Forbidden host mount
- Port already allocated
- Invalid image reference
- Unsupported Podman capability

Permanent errors are reported immediately and are not retried continuously until:

- The workload generation changes.
- The relevant host capability changes.
- An administrator requests reconciliation.

This prevents invalid workloads from causing endless high-frequency retry loops.

---

# Security Boundaries

The agent controls Podman and must therefore be treated as a privileged host component.

The initial implementation should enforce these rules:

- The agent accepts desired state only from the authenticated operator.
- The operator derives host identity from the client certificate.
- The agent validates the operator certificate and expected identity.
- Arbitrary command execution is not supported.
- Only supported Kubernetes resource types are accepted.
- Dangerous host mounts are rejected by default.
- Privileged containers are rejected by default.
- Mounting the Podman socket into managed workloads is rejected.
- Host namespace access is rejected unless explicitly permitted.
- Manifests are size-limited.
- Status and error messages must not contain secret values.
- All agent actions are logged with workload UID and revision.
- Existing unmanaged Podman resources are never modified.

---

# Initial Protocol Messages

The first protocol version requires only a small set of messages.

## Agent to Operator

```text
AgentHello
EnrollmentRequest
Heartbeat
HostInventory
ReconciliationResult
WorkloadStatus
CertificateRotationRequest
```

## Operator to Agent

```text
OperatorHello
EnrollmentResponse
DesiredState
ReconcileRequest
CertificateRotationResponse
ConnectionClose
```

The agent should not initially support generic messages such as:

```text
ExecuteCommand
RunShell
CopyArbitraryFile
CallPodmanAPI
```

Those would unnecessarily expand the security boundary.

---

# MVP Behavior

The first agent implementation should support:

- One outbound gRPC connection.
- mTLS authentication.
- One agent identity per `ExternalHost`.
- Periodic heartbeats.
- Podman capability reporting.
- Complete desired-state snapshots.
- One or more `PodmanPlay` workloads per host.
- `Pod` manifests.
- Workload creation using `podman kube play`.
- Workload updates using `podman kube play --replace`.
- Workload removal using `podman kube down`.
- Local ownership labels.
- Per-workload status reporting.
- Reconnection with exponential backoff.
- Reconciliation after agent restart.
- Existing workloads continuing during connection loss.

---

# Open Questions

The following decisions still need to be made:

1. Should the agent use the Podman REST API or invoke the Podman CLI?
2. Should the initial implementation support rootful or rootless Podman?
3. How is the initial enrollment token generated and delivered?
4. Which certificate authority issues agent certificates?
5. How long should agent certificates remain valid?
6. What exact subset of Kubernetes Pod fields is supported?
7. Should ConfigMaps and Secrets be embedded in the desired-state snapshot?
8. How should workloads be handled when their host group changes?
9. How long should the operator wait for an offline host during deletion?
10. Should the agent periodically reconcile even without receiving a new revision?

---

# Recommended Initial Decisions

For the first proof of concept:

- Use bidirectional gRPC streaming over mTLS.
- Use complete desired-state snapshots.
- Use one-time enrollment tokens.
- Bind one certificate to one `ExternalHost`.
- Use rootful Podman initially.
- Invoke the Podman CLI initially.
- Support only Kubernetes `Pod` manifests.
- Reconcile periodically every 60 seconds.
- Leave existing workloads running while disconnected.
- Reject privileged workloads and dangerous host mounts.
- Do not provide remote shell or arbitrary command execution.