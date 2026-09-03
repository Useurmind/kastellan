# Protocol Buffer API Reference

## Package
`kastellan.agent.v1alpha1`

## Service Definition

```protobuf
service AgentProtocol {
  rpc Connect(stream ProtocolMessage) returns (stream ProtocolMessage);
}
```

## Common Types

### ProtocolVersion
Represents a semantic version number.

| Field | Type | Description |
|-------|------|-------------|
| major | uint32 | Major version number |
| minor | uint32 | Minor version number |

### ResourceIdentity
Uniquely identifies a Kubernetes resource.

| Field | Type | Description |
|-------|------|-------------|
| api_version | string | API version (e.g., "kastellan.io/v1alpha1") |
| kind | string | Resource kind (e.g., "PodmanPlay") |
| namespace | string | Resource namespace |
| name | string | Resource name |
| uid | string | Resource UID (unique identifier) |
| generation | int64 | Resource generation number |

### ErrorDetail
Provides detailed error information.

| Field | Type | Description |
|-------|------|-------------|
| code | string | Error code (machine-readable) |
| message | string | Human-readable error message |
| retryable | bool | Whether the operation can be retried |

### RuntimeInformation
Information about the runtime environment.

| Field | Type | Description |
|-------|------|-------------|
| name | string | Runtime name (e.g., "podman") |
| version | string | Runtime version |
| capabilities | repeated string | List of supported capabilities |

### RuntimeHealth
Runtime health status.

| Field | Type | Description |
|-------|------|-------------|
| available | bool | Whether the runtime is available |
| message | string | Human-readable status message |

## Agent Messages

### AgentHello
Initial connection message sent by the agent.

| Field | Type | Description |
|-------|------|-------------|
| agent_id | string | Unique agent identifier |
| host_name | string | Host name |
| agent_version | string | Agent version |
| supported_protocols | repeated ProtocolVersion | Supported protocol versions |
| runtime | RuntimeInformation | Runtime environment info |
| last_session_id | string | Previous session ID (for reconnects) |
| last_received_revision | uint64 | Last received revision number |
| last_applied_revision | uint64 | Last applied revision number |

### Heartbeat
Periodic status update from the agent.

| Field | Type | Description |
|-------|------|-------------|
| session_id | string | Current session ID |
| runtime | RuntimeHealth | Runtime health status |
| workloads | WorkloadStatus | Workload status summary |
| host_name | string | Host name |

### HostInventory
Host capabilities and inventory.

| Field | Type | Description |
|-------|------|-------------|
| session_id | string | Current session ID |
| host | HostInfo | Host system information |
| podman | PodmanInfo | Podman runtime information |
| capabilities | repeated string | List of supported capabilities |

### ReconcileResult
Workload reconciliation results.

| Field | Type | Description |
|-------|------|-------------|
| session_id | string | Current session ID |
| host | string | Host name |
| revision | uint64 | Applied revision number |
| workloads | repeated WorkloadResult | List of workload results |

### WorkloadStatus
Status of a single workload.

| Field | Type | Description |
|-------|------|-------------|
| session_id | string | Current session ID |
| identity | ResourceIdentity | Workload identity |
| phase | string | Current phase (Pending, Applying, Ready, Failed, etc.) |
| runtime | RuntimeInformation | Runtime information |

## Operator Messages

### OperatorHello
Server's response to AgentHello.

| Field | Type | Description |
|-------|------|-------------|
| session_id | string | Assigned session ID |
| selected_protocol | ProtocolVersion | Negotiated protocol version |
| heartbeat_interval_seconds | uint32 | Heartbeat interval |
| state_report_interval_seconds | uint32 | State report interval |
| server_time_unix | int64 | Server timestamp (Unix epoch) |

### DesiredState
Complete workload assignment.

| Field | Type | Description |
|-------|------|-------------|
| host | string | Target host name |
| revision | uint64 | Revision number |
| podman_plays | repeated PodmanPlayAssignment | List of workload assignments |

### DesiredStateUpdate
Incremental updates.

| Field | Type | Description |
|-------|------|-------------|
| session_id | string | Current session ID |
| host | string | Target host name |
| revision | uint64 | Revision number |
| additions | repeated PodmanPlayAssignment | New workloads to add |
| deletions | repeated string | UIDs of workloads to delete |

## Proto Message Wrapper

### ProtocolMessage
Wrapper for all protocol messages using oneof pattern.

| Field | Type | Description |
|-------|------|-------------|
| payload | oneof | One of the protocol messages |

Supported payloads include all agent and operator messages, plus common types.

## Go Package Usage

### Import
```go
import (
    agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)
```

### Create Protocol Version
```go
version := &agentv1alpha1.ProtocolVersion{
    Major: 1,
    Minor: 0,
}
```

### Create Resource Identity
```go
identity := &agentv1alpha1.ResourceIdentity{
    ApiVersion: "kastellan.io/v1alpha1",
    Kind:       "PodmanPlay",
    Namespace:  "default",
    Name:       "my-workload",
    Uid:        "123e4567-e89b-12d3-a456-426614174000",
    Generation: 1,
}
```

### Convert to/from Go Types
```go
// From existing messages
containerProto := container.ToProto()
playProto := play.ToProto()

// To existing messages
container := messages.ContainerInfoFromProto(protoContainer)
play := messages.PodmanPlayFromProto(protoPlay)
```

## gRPC Service Registration

### Server Side
```go
import (
    "google.golang.org/grpc"
    agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)

// Create gRPC server
grpcServer := grpc.NewServer()

// Register the service
agentv1alpha1.RegisterAgentProtocolServer(grpcServer, myService)
```

### Client Side
```go
import (
    "google.golang.org/grpc"
    agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)

// Create connection
conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
if err != nil {
    return err
}
defer conn.Close()

// Create client
client := agentv1alpha1.NewAgentProtocolClient(conn)
```

## Build Integration

### Generate Proto Files
```bash
make proto-generate
```

### Format Proto Files
```bash
make proto-fmt
```

### Lint Proto Files (requires protoc-gen-lint)
```bash
make proto-lint
```

### Run Proto Tests
```bash
go test ./pkg/agentprotocol/messages/proto/test/... -v
make test-proto
```

## Serialization Examples

### Marshaling
```go
proto := &agentv1alpha1.AgentHello{
    AgentId:      "agent-123",
    HostName:     "host-1",
    AgentVersion: "0.1.0",
}

data, err := proto.Marshal()
if err != nil {
    return err
}
// data contains the protobuf-encoded bytes
```

### Unmarshaling
```go
proto := &agentv1alpha1.AgentHello{}
err := proto.Unmarshal(data)
if err != nil {
    return err
}
```

### JSON Marshaling (via protojson)
```go
import "google.golang.org/protobuf/encoding/protojson"

jsonBytes, err := protojson.Marshal(proto)
if err != nil {
    return err
}
// jsonBytes contains JSON-encoded data
```

## Type Mappings

| Proto Type | Go Type |
|------------|---------|
| string | string |
| uint32 | uint32 |
| uint64 | uint64 |
| int32 | int32 |
| int64 | int64 |
| bool | bool |
| repeated T | []T or []*T |
| enum | int32 (generated type with constants) |

## Notes

1. All generated types implement `proto.Message` interface
2. Use `ProtoReflect()` for reflection-based operations
3. Use `Marshal()` and `Unmarshal()` for serialization
4. Use `Reset()` to clear message state
5. Use `String()` for debugging output
6. All fields are optional by default in proto3
7. Zero values are serialized as empty (not explicit nil)
