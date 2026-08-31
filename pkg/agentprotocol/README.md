# Agent Protocol Package

This package provides the agent-side connection protocol for communicating with the Kastellan Operator.

## Overview

The agent protocol package implements the communication protocol between the Kastellan Agent and the Kastellan Operator. It provides:

- Protocol message structures (JSON-based)
- gRPC client for bidirectional streaming
- Connection management with reconnection logic
- Authentication (mTLS with enrollment tokens)
- Heartbeat mechanism
- Status reporting

## Package Structure

```
pkg/agentprotocol/
├── messages/          # Protocol message definitions
│   ├── types.go       # Message type definitions
│   ├── json.go        # JSON serialization
│   └── messages_test.go
├── client/            # Client implementation
│   ├── client.go      # Main client
│   ├── connection.go  # Connection management
│   ├── reconnect.go   # Reconnection logic
│   ├── heartbeat.go   # Heartbeat handling
│   ├── status.go      # Status reporting
│   └── client_test.go
├── errors.go          # Error types
├── const.go           # Constants
└── README.md          # This file
```

## Usage

### Creating a Client

```go
import (
    "context"
    "time"
    
    "github.com/useurmind/kastellan/pkg/agentprotocol/client"
    "github.com/useurmind/kastellan/pkg/agentprotocol/messages"
)

// Create a new client
c := client.New(
    "kastellan-operator:443",  // server address
    "agent-2fd317",            // agent ID
    "0.1.0",                   // agent version
    "lb01",                    // host name
    "lb01.example.internal",   // host hostname
)

// Configure the client
c = c.WithCertificates(
    "/etc/kastellan/certs/certificate.pem",
    "/etc/kastellan/certs/private-key.pem",
    "/etc/kastellan/certs/ca-bundle.pem",
)

// Connect to the operator
ctx := context.Background()
if err := c.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer c.Close()

// Wait for connection
select {
case <-c.GetConnectedChannel():
    log.Println("Connected to operator")
case <-ctx.Done():
    log.Fatal("Connection timeout")
}
```

### Sending Messages

```go
// Send a heartbeat
heartbeat := messages.Heartbeat{
    Type:      messages.MessageTypeHeartbeat,
    SessionID: "session-id",
    Timestamp: time.Now(),
    // ... other fields
}

// Send through the message channel
c.GetMessageChannel() <- heartbeat
```

### Reconnection

The client automatically handles reconnection with exponential backoff:

```go
// The client will automatically reconnect when the connection is lost
// The reconnection delay increases exponentially up to the maximum delay

// You can configure the reconnection settings
c = c.WithReconnectDelay(1 * time.Second)
c = c.WithMaxReconnectDelay(1 * time.Minute)
```

### Heartbeat

The client automatically sends heartbeats at the configured interval:

```go
// Configure heartbeat interval
c = c.WithHeartbeatInterval(30 * time.Second)

// The heartbeat is sent automatically
// You can also manually trigger a heartbeat
heartbeat := client.CreateHeartbeatMessage("session-id", "lb01")
c.GetMessageChannel() <- heartbeat
```

### Status Reporting

The client automatically reports workload status:

```go
// Update workload state
state := &client.WorkloadState{
    UID:         "workload-uid",
    Namespace:   "default",
    Name:        "workload-name",
    Generation:  1,
    Phase:       "Running",
    LastUpdate:  time.Now(),
}
statusReporter.UpdateWorkloadState(state)

// Status is reported automatically at the configured interval
```

## Protocol Messages

### Agent to Operator

- `AgentHello` - Initial connection message
- `EnrollmentRequest` - Enrollment request with token
- `Heartbeat` - Periodic status updates
- `HostInventory` - Host capabilities
- `ReconciliationResult` - Workload status reporting
- `WorkloadStatus` - Single workload status
- `CertificateRotationRequest` - Certificate rotation request

### Operator to Agent

- `OperatorHello` - Server response to AgentHello
- `EnrollmentResponse` - Enrollment response with certificates
- `DesiredState` - Workload assignments
- `DesiredStateUpdate` - Incremental updates

## Connection Lifecycle

1. **Enrollment**: Agent connects with enrollment token
2. **Authentication**: Operator validates token and issues certificates
3. **Connection**: Agent reconnects with mTLS
4. **Heartbeat**: Periodic status updates
5. **Reconciliation**: Workload assignments and status reporting
6. **Reconnection**: Automatic reconnection on connection loss

## Error Handling

The package provides error types for different error conditions:

```go
import (
    "github.com/useurmind/kastellan/pkg/agentprotocol"
    "github.com/useurmind/kastellan/pkg/agentprotocol/messages"
)

// Check for protocol errors
if agentprotocol.IsProtocolError(err) {
    protocolErr := err.(*agentprotocol.ProtocolError)
    log.Printf("Protocol error: %s", protocolErr.Code)
}

// Check for connection errors
if connErr, ok := err.(*agentprotocol.ConnectionError); ok {
    if connErr.IsRetryable() {
        // Retry the connection
    }
}
```

## Configuration

### Default Values

```go
const (
    DefaultHeartbeatInterval    = 30 * time.Second
    DefaultStateReportInterval  = 60 * time.Second
    DefaultOfflineAfter         = 2 * time.Minute
    DefaultReconnectInitialDelay = time.Second
    DefaultReconnectMaxDelay    = time.Minute
)
```

### Custom Configuration

```go
// Create client with custom configuration
c := client.New(serverAddress, agentID, agentVersion, hostName, hostHostname)
c = c.WithReconnectDelay(2 * time.Second)
c = c.WithMaxReconnectDelay(2 * time.Minute)
c = c.WithHeartbeatInterval(60 * time.Second)
```

## Testing

Run the tests:

```bash
go test ./pkg/agentprotocol/... -v
```

## Documentation

Generate documentation:

```bash
go doc -all github.com/useurmind/kastellan/pkg/agentprotocol
```

## License

Apache 2.0
