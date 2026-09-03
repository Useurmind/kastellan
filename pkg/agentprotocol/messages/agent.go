// Package messages defines the protocol message structures for the Kastellan Agent Protocol.
package messages

import (
	"time"
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
