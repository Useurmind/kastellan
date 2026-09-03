## Protocol Buffer Serialization Plan
### Phase 1: Proto Definition Structure
Create `api/proto/kastellan/agent/v1alpha1/agent.proto` with the following structure:
**Key Design Decisions:**
1. Package: `kastellan.agent.v1alpha1`
2. Go package: `github.com/useurmind/kastellan/api/proto/kastellan/agent/v1alpha1`
3. Single bidirectional stream service: `Connect`
4. Oneof payload pattern for message typing
### Phase 2: Mapped Messages (from spec to current implementation)

| Spec Message | Current Go Type | Status | Notes |
|---|---|---|---|
| AgentHello | AgentHello | ✅ Exists | Needs field mapping |
| OperatorHello | OperatorHello | ✅ Exists | Needs field mapping |
| DesiredState | DesiredState | ✅ Exists | Needs field mapping |
| DesiredStateUpdate | DesiredStateUpdate | ✅ Exists | Consider for MVP |
| Heartbeat | Heartbeat | ✅ Exists | Needs field mapping |
| HostInventory | HostInventory | ✅ Exists | Needs field mapping |
| ReconcileResult | ReconciliationResult | ✅ Exists | Name mismatch |
| ReconcileRequest | ReconcileRequest | ✅ Exists | Name mismatch |
| WorkloadStatus | WorkloadStatus | ⚠️ Partial | Needs spec alignment |
| ProtocolVersion | — | ❌ Missing | Common type needed |
| ResourceIdentity | — | ❌ Missing | Common type needed |
| ErrorDetail | ProtocolError | ⚠️ Partial | Name/field mapping |
| RuntimeInformation | Runtime | ⚠️ Partial | Need capabilities |
| RuntimeHealth | Runtime | ⚠️ Partial | Need available/message |
| WorkloadResult | WorkloadResult | ✅ Exists | Needs spec alignment |
| RuntimeResource | ContainerInfo | ⚠️ Partial | Need type/id/state |
| LocalWorkload | — | ❌ Missing | Inventory requirement |
| PodmanPlayAssignment | PodmanPlay | ⚠️ Partial | Need update strategy |
### Phase 3: Implementation Steps
**Step 1: Create proto directory and definition**
- Create `api/proto/kastellan/agent/v1alpha1/` directory
- Define `agent.proto` with all message types
- Include service definition for bidirectional stream
**Step 2: Generate Go code**
- Add protobuf generation to Makefile
- Run `make generate` to produce Go types
- Generate `zz_generated.*.go` files
**Step 3: Update Go types (backward compatibility)**
- Keep existing Go structs in `pkg/agentprotocol/messages/`
- Add proto conversion methods
- Implement proto wrappers in existing package
- Maintain JSON compatibility for existing consumers
**Step 4: Add proto-specific types**
- Create `api/proto/kastellan/agent/v1alpha1/` package
- Add proto message structs (auto-generated)
- Add conversion functions between proto and Go types
**Step 5: Test suite updates**
- Add proto-specific unit tests
- Update existing tests to use proto types
- Test serialization/deserialization round-trips
- Add benchmark tests for proto vs JSON
### Phase 4: Required Mappings
**Common Types (New):**
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
**Agent Messages:**
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
**Operator Messages:**
message OperatorHello {
  string session_id = 1;
  ProtocolVersion selected_protocol = 2;
  uint32 heartbeat_interval_seconds = 3;
  uint32 state_report_interval_seconds = 4;
  int64 server_time_unix = 5;
}
### Phase 5: Migration Strategy
**Option A: Parallel Implementation (Recommended)**
- Keep existing JSON-based messages intact
- Add proto-based messages alongside
- Gradually migrate consumers
- Maintain dual support during transition
**Option B: Full Replacement**
- Replace existing types with proto-generated ones
- Breaking change for existing consumers
- Requires coordinated update of all consumers
### Phase 6: Build & CI Integration
**Makefile updates:**
- Add `make proto-generate` target
- Add `make proto-lint` target
- Update `make generate` to include proto generation
- Update `make lint` to include proto linting
**CI checks:**
- Proto definition validation
- Generated code compatibility
- Benchmark regression detection
### Phase 7: Documentation
- Update `specs/communication.md` with proto details
- Add migration guide
- Document proto versioning strategy
- Add API compatibility guarantees
### Estimated Files to Create/Modify
**Create:**
- `api/proto/kastellan/agent/v1alpha1/agent.proto`
- `api/proto/kastellan/agent/v1alpha1/agent_grpc.pb.go` (generated)
- `api/proto/kastellan/agent/v1alpha1/agent.pb.go` (generated)
**Modify:**
- `pkg/agentprotocol/messages/common.go` (add proto conversions)
- `pkg/agentprotocol/messages/agent.go` (add proto conversions)
- `pkg/agentprotocol/messages/operator.go` (add proto conversions)
- `Makefile` (add proto targets)
- `go.mod` (add protobuf dependencies)
### Dependencies to Add
require (
  google.golang.org/protobuf v1.35.0
  google.golang.org/grpc v1.68.0
  google.golang.org/genproto/googleapis/rpc v0.0.0-20240515...
)
Would you like me to proceed with this plan? I can start with creating the proto definition file and the migration strategy.