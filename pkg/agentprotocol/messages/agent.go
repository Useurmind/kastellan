// Package messages defines the protocol message structures for the Kastellan Agent Protocol.
package messages

import (
	"time"

	agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)

// AgentHello is the initial connection message sent by the agent.
type AgentHello struct {
	Type MessageType `json:"type"`

	// Agent identification
	Agent struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"agent"`

	// Host identification
	Host struct {
		Name      string `json:"name"`
		Hostname  string `json:"hostname"`
		IPAddress string `json:"ipAddress,omitempty"`
	} `json:"host"`

	// Protocol versions supported by the agent
	ProtocolVersions []string `json:"protocolVersions"`

	// Runtime information
	Runtime struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"runtime"`

	// Agent capabilities
	Capabilities []string `json:"capabilities"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`
}

// EnrollmentRequest is sent by the agent during initial enrollment.
type EnrollmentRequest struct {
	Type MessageType `json:"type"`

	// Enrollment token (one-time use)
	Token string `json:"token"`

	// Agent information
	Agent struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"agent"`

	// Host information
	Host struct {
		Name      string `json:"name"`
		Hostname  string `json:"hostname"`
		IPAddress string `json:"ipAddress,omitempty"`
	} `json:"host"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`
}

// Heartbeat is a periodic status update from the agent.
type Heartbeat struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`

	// Runtime status
	Runtime struct {
		Available bool   `json:"available"`
		Error     string `json:"error,omitempty"`
	} `json:"runtime"`

	// Workload status
	Workloads struct {
		Assigned int `json:"assigned"`
		Ready    int `json:"ready"`
		Failed   int `json:"failed"`
		Updating int `json:"updating,omitempty"`
		Unknown  int `json:"unknown,omitempty"`
	} `json:"workloads"`

	// Host information
	Host struct {
		Name string `json:"name"`
	} `json:"host"`
}

// HostInventory is sent by the agent to report host capabilities.
type HostInventory struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`

	// Host information
	Host struct {
		Name     string `json:"name"`
		Hostname string `json:"hostname"`
		OS       string `json:"os"`
		Kernel   string `json:"kernel"`
		CPU      string `json:"cpu"`
		Memory   string `json:"memory"`
		Storage  string `json:"storage"`
	} `json:"host"`

	// Podman runtime information
	Podman struct {
		Version    string `json:"version"`
		Rootful    bool   `json:"rootful"`
		APIVersion string `json:"apiVersion"`
	} `json:"podman"`

	// Available capabilities
	Capabilities []string `json:"capabilities"`
}

// ReconciliationResult reports the result of workload reconciliation.
type ReconciliationResult struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Host name
	Host string `json:"host"`

	// Desired state revision
	Revision int64 `json:"revision"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`

	// Workload results
	Workloads []WorkloadResult `json:"workloads"`
}

// WorkloadStatus reports the status of a single workload.
type WorkloadStatus struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`

	// Workload identification
	UID        string `json:"uid"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Generation int64  `json:"generation"`

	// Current phase
	Phase string `json:"phase"`

	// Runtime information
	Runtime struct {
		PodID      string          `json:"podId,omitempty"`
		Containers []ContainerInfo `json:"containers,omitempty"`
	} `json:"runtime,omitempty"`
}

// CertificateRotationRequest requests certificate rotation.
type CertificateRotationRequest struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`
}

// Validate performs basic validation on AgentHello.
func (a *AgentHello) Validate() error {
	if a.Agent.ID == "" {
		return &ProtocolError{Code: "invalid_agent_id", Message: "agent ID is required"}
	}
	if a.Agent.Version == "" {
		return &ProtocolError{Code: "invalid_agent_version", Message: "agent version is required"}
	}
	if a.Host.Name == "" {
		return &ProtocolError{Code: "invalid_host_name", Message: "host name is required"}
	}
	if len(a.ProtocolVersions) == 0 {
		return &ProtocolError{Code: "no_protocol_versions", Message: "at least one protocol version is required"}
	}
	return nil
}

// Validate performs basic validation on EnrollmentRequest.
func (e *EnrollmentRequest) Validate() error {
	if e.Token == "" {
		return &ProtocolError{Code: "missing_token", Message: "enrollment token is required"}
	}
	if e.Host.Name == "" {
		return &ProtocolError{Code: "invalid_host_name", Message: "host name is required"}
	}
	return nil
}

// Validate performs basic validation on Heartbeat.
func (h *Heartbeat) Validate() error {
	if h.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	return nil
}

// Validate performs basic validation on HostInventory.
func (h *HostInventory) Validate() error {
	if h.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	if h.Host.Name == "" {
		return &ProtocolError{Code: "missing_host_name", Message: "host name is required"}
	}
	return nil
}

// Validate performs basic validation on ReconciliationResult.
func (r *ReconciliationResult) Validate() error {
	if r.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	if r.Host == "" {
		return &ProtocolError{Code: "missing_host", Message: "host name is required"}
	}
	return nil
}

// Validate performs basic validation on WorkloadStatus.
func (w *WorkloadStatus) Validate() error {
	if w.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	if w.UID == "" {
		return &ProtocolError{Code: "missing_uid", Message: "workload UID is required"}
	}
	return nil
}

// Validate performs basic validation on CertificateRotationRequest.
func (c *CertificateRotationRequest) Validate() error {
	if c.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	return nil
}

// ToProto converts AgentHello to proto AgentHello.
func (a *AgentHello) ToProto() *agentv1alpha1.AgentHello {
	supportedProtocols := make([]*agentv1alpha1.ProtocolVersion, len(a.ProtocolVersions))
	for i := range a.ProtocolVersions {
		supportedProtocols[i] = &agentv1alpha1.ProtocolVersion{}
	}
	return &agentv1alpha1.AgentHello{
		AgentId:              a.Agent.ID,
		HostName:             a.Host.Name,
		AgentVersion:         a.Agent.Version,
		SupportedProtocols:   supportedProtocols,
		Runtime:              &agentv1alpha1.RuntimeInformation{},
		LastSessionId:        "",
		LastReceivedRevision: 0,
		LastAppliedRevision:  0,
	}
}

// AgentHelloFromProto converts proto AgentHello to AgentHello.
func AgentHelloFromProto(p *agentv1alpha1.AgentHello) *AgentHello {
	if p == nil {
		return nil
	}
	return &AgentHello{
		Type: MessageTypeAgentHello,
	}
}

// ToProto converts HostInventory to proto HostInventory.
func (h *HostInventory) ToProto() *agentv1alpha1.HostInventory {
	return &agentv1alpha1.HostInventory{
		SessionId: h.SessionID,
		Host:      &agentv1alpha1.HostInfo{},
		Podman:    &agentv1alpha1.PodmanInfo{},
		Capabilities: h.Capabilities,
	}
}

// HostInventoryFromProto converts proto HostInventory to HostInventory.
func HostInventoryFromProto(p *agentv1alpha1.HostInventory) *HostInventory {
	if p == nil {
		return nil
	}
	return &HostInventory{
		Type:       MessageTypeHostInventory,
		SessionID:  p.GetSessionId(),
		Capabilities: p.GetCapabilities(),
	}
}

// ToProto converts Heartbeat to proto Heartbeat.
func (h *Heartbeat) ToProto() *agentv1alpha1.Heartbeat {
	return &agentv1alpha1.Heartbeat{
		SessionId: h.SessionID,
		Runtime:   &agentv1alpha1.RuntimeHealth{},
		Workloads: &agentv1alpha1.WorkloadStatus{},
		HostName:  h.Host.Name,
	}
}

// HeartbeatFromProto converts proto Heartbeat to Heartbeat.
func HeartbeatFromProto(p *agentv1alpha1.Heartbeat) *Heartbeat {
	if p == nil {
		return nil
	}
	return &Heartbeat{
		Type:      MessageTypeHeartbeat,
		SessionID: p.GetSessionId(),
	}
}

// ToProto converts ReconciliationResult to proto ReconcileResult.
func (r *ReconciliationResult) ToProto() *agentv1alpha1.ReconcileResult {
	protoWorkloads := make([]*agentv1alpha1.WorkloadResult, len(r.Workloads))
	for i, w := range r.Workloads {
		protoWorkloads[i] = w.ToProto()
	}
	return &agentv1alpha1.ReconcileResult{
		SessionId: r.SessionID,
		Host:      r.Host,
		Revision:  uint64(r.Revision),
		Workloads: protoWorkloads,
	}
}

// ReconciliationResultFromProto converts proto ReconcileResult to ReconciliationResult.
func ReconciliationResultFromProto(p *agentv1alpha1.ReconcileResult) *ReconciliationResult {
	if p == nil {
		return nil
	}
	workloads := make([]WorkloadResult, len(p.GetWorkloads()))
	for i, w := range p.GetWorkloads() {
		workloads[i] = *WorkloadResultFromProto(w)
	}
	return &ReconciliationResult{
		Type:      MessageTypeReconciliationResult,
		SessionID: p.GetSessionId(),
		Host:      p.GetHost(),
		Revision:  int64(p.GetRevision()),
		Workloads: workloads,
	}
}

// ToProto converts WorkloadStatus to proto WorkloadStatus.
func (w *WorkloadStatus) ToProto() *agentv1alpha1.WorkloadStatus {
	return &agentv1alpha1.WorkloadStatus{
		SessionId: w.SessionID,
		Identity: &agentv1alpha1.ResourceIdentity{
			Uid:        w.UID,
			Namespace:  w.Namespace,
			Name:       w.Name,
			Generation: w.Generation,
		},
		Phase: w.Phase,
		Runtime: &agentv1alpha1.RuntimeInformation{},
	}
}

// WorkloadStatusFromProto converts proto WorkloadStatus to WorkloadStatus.
func WorkloadStatusFromProto(p *agentv1alpha1.WorkloadStatus) *WorkloadStatus {
	if p == nil {
		return nil
	}
	identity := p.GetIdentity()
	return &WorkloadStatus{
		Type:       MessageTypeWorkloadStatus,
		SessionID:  p.GetSessionId(),
		UID:        identity.GetUid(),
		Namespace:  identity.GetNamespace(),
		Name:       identity.GetName(),
		Generation: identity.GetGeneration(),
		Phase:      p.GetPhase(),
	}
}
