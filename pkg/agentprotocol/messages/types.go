// Package messages defines the protocol message structures for the Kastellan Agent Protocol.
package messages

import (
	"fmt"
	"time"
)

// MessageType represents the type of a protocol message.
type MessageType string

const (
	// Agent messages
	MessageTypeAgentHello                 MessageType = "AgentHello"
	MessageTypeEnrollmentRequest          MessageType = "EnrollmentRequest"
	MessageTypeHeartbeat                  MessageType = "Heartbeat"
	MessageTypeHostInventory              MessageType = "HostInventory"
	MessageTypeReconciliationResult       MessageType = "ReconciliationResult"
	MessageTypeWorkloadStatus             MessageType = "WorkloadStatus"
	MessageTypeCertificateRotationRequest MessageType = "CertificateRotationRequest"

	// Operator messages
	MessageTypeOperatorHello      MessageType = "OperatorHello"
	MessageTypeEnrollmentResponse MessageType = "EnrollmentResponse"
	MessageTypeDesiredState       MessageType = "DesiredState"
	MessageTypeDesiredStateUpdate MessageType = "DesiredStateUpdate"
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

// OperatorHello is the server's response to AgentHello.
type OperatorHello struct {
	Type MessageType `json:"type"`

	// Session identification
	Session struct {
		ID string `json:"id"`
	} `json:"session"`

	// Negotiated protocol version
	ProtocolVersion string `json:"protocolVersion"`

	// Configuration from operator
	Configuration struct {
		HeartbeatInterval    string `json:"heartbeatInterval"`
		StateReportInterval  string `json:"stateReportInterval"`
		OfflineAfter         string `json:"offlineAfter"`
		MaxManifestSizeBytes int    `json:"maxManifestSizeBytes,omitempty"`
	} `json:"configuration"`

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

// EnrollmentResponse is sent by the operator in response to EnrollmentRequest.
type EnrollmentResponse struct {
	Type MessageType `json:"type"`

	// Success indicates if enrollment was successful
	Success bool `json:"success"`

	// Error message if enrollment failed
	Error string `json:"error,omitempty"`

	// Agent identity (certificate and key)
	Identity struct {
		Certificate string `json:"certificate"`
		PrivateKey  string `json:"privateKey"`
		CABundle    string `json:"caBundle"`
	} `json:"identity,omitempty"`

	// Session ID for subsequent connections
	SessionID string `json:"sessionId,omitempty"`

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

// WorkloadResult represents the result for a single workload.
type WorkloadResult struct {
	// Workload identification
	UID        string `json:"uid"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Generation int64  `json:"generation"`

	// Current phase
	Phase string `json:"phase"`

	// Error details if phase is Failed
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`

	// Manifest digest
	ManifestDigest string `json:"manifestDigest"`

	// Runtime information
	Runtime struct {
		PodID      string          `json:"podId,omitempty"`
		Containers []ContainerInfo `json:"containers,omitempty"`
	} `json:"runtime,omitempty"`
}

// ContainerInfo represents information about a container.
type ContainerInfo struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	State string `json:"state"`
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

// DesiredState is sent by the operator to deliver workload assignments.
type DesiredState struct {
	Type MessageType `json:"type"`

	// Host name
	Host string `json:"host"`

	// Revision number (monotonically increasing)
	Revision int64 `json:"revision"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`

	// PodmanPlay resources assigned to this host
	PodmanPlays []PodmanPlay `json:"podmanPlays"`
}

// PodmanPlay represents a workload assignment.
type PodmanPlay struct {
	// Unique identifier
	UID string `json:"uid"`

	// Kubernetes namespace
	Namespace string `json:"namespace"`

	// Name of the resource
	Name string `json:"name"`

	// Generation number
	Generation int64 `json:"generation"`

	// Full YAML manifest
	Manifest string `json:"manifest"`
}

// DesiredStateUpdate is sent by the operator for incremental updates.
type DesiredStateUpdate struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Host name
	Host string `json:"host"`

	// Revision number
	Revision int64 `json:"revision"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`

	// Additions
	Additions []PodmanPlay `json:"additions,omitempty"`

	// Deletions
	Deletions []string `json:"deletions,omitempty"` // UIDs of workloads to delete
}

// ProtocolError represents a protocol-level error.
type ProtocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns the error message.
func (e *ProtocolError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
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

// ReconcileRequest is sent by the operator to request reconciliation.
type ReconcileRequest struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Host name
	Host string `json:"host"`

	// Revision number
	Revision int64 `json:"revision"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`
}

// CertificateRotationResponse is sent by the operator in response to certificate rotation request.
type CertificateRotationResponse struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Success indicates if rotation was successful
	Success bool `json:"success"`

	// Error message if rotation failed
	Error string `json:"error,omitempty"`

	// New certificate and key
	Identity struct {
		Certificate string `json:"certificate,omitempty"`
		PrivateKey  string `json:"privateKey,omitempty"`
		CABundle    string `json:"caBundle,omitempty"`
	} `json:"identity,omitempty"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`
}

// ConnectionClose is sent by the operator to close the connection.
type ConnectionClose struct {
	Type MessageType `json:"type"`

	// Session ID
	SessionID string `json:"sessionId"`

	// Reason for closing
	Reason string `json:"reason,omitempty"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`
}

// Validate performs basic validation on ReconcileRequest.
func (r *ReconcileRequest) Validate() error {
	if r.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	if r.Host == "" {
		return &ProtocolError{Code: "missing_host", Message: "host name is required"}
	}
	return nil
}

// Validate performs basic validation on CertificateRotationResponse.
func (c *CertificateRotationResponse) Validate() error {
	if c.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	return nil
}

// Validate performs basic validation on ConnectionClose.
func (c *ConnectionClose) Validate() error {
	if c.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	return nil
}

// Validate performs basic validation on OperatorHello.
func (o *OperatorHello) Validate() error {
	if o.Session.ID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	if o.ProtocolVersion == "" {
		return &ProtocolError{Code: "missing_protocol_version", Message: "protocol version is required"}
	}
	return nil
}

// Validate performs basic validation on EnrollmentResponse.
func (e *EnrollmentResponse) Validate() error {
	if e.SessionID == "" {
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

// Validate performs basic validation on DesiredState.
func (d *DesiredState) Validate() error {
	if d.Host == "" {
		return &ProtocolError{Code: "missing_host", Message: "host name is required"}
	}
	return nil
}

// Validate performs basic validation on DesiredStateUpdate.
func (d *DesiredStateUpdate) Validate() error {
	if d.SessionID == "" {
		return &ProtocolError{Code: "missing_session_id", Message: "session ID is required"}
	}
	if d.Host == "" {
		return &ProtocolError{Code: "missing_host", Message: "host name is required"}
	}
	return nil
}

// Validate performs basic validation on PodmanPlay.
func (p *PodmanPlay) Validate() error {
	if p.UID == "" {
		return &ProtocolError{Code: "missing_uid", Message: "workload UID is required"}
	}
	if p.Name == "" {
		return &ProtocolError{Code: "missing_name", Message: "workload name is required"}
	}
	if p.Manifest == "" {
		return &ProtocolError{Code: "missing_manifest", Message: "manifest is required"}
	}
	return nil
}
