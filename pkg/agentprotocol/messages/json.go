package messages

import (
	"encoding/json"
	"fmt"
	"time"
)

// MarshalJSON marshals the message to JSON bytes.
func (a *AgentHello) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type MessageType `json:"type"`
		Agent struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		} `json:"agent"`
		Host struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		} `json:"host"`
		ProtocolVersions []string `json:"protocolVersions"`
		Runtime struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"runtime"`
		Capabilities []string `json:"capabilities"`
		Timestamp    time.Time `json:"timestamp"`
	}{
		Type:             a.Type,
		Agent:            a.Agent,
		Host:             a.Host,
		ProtocolVersions: a.ProtocolVersions,
		Runtime:          a.Runtime,
		Capabilities:     a.Capabilities,
		Timestamp:        a.Timestamp,
	})
}

// UnmarshalJSON unmarshals JSON bytes into AgentHello.
func (a *AgentHello) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type MessageType `json:"type"`
		Agent struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		} `json:"agent"`
		Host struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		} `json:"host"`
		ProtocolVersions []string `json:"protocolVersions"`
		Runtime struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"runtime"`
		Capabilities []string `json:"capabilities"`
		Timestamp    time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	a.Type = raw.Type
	a.Agent.ID = raw.Agent.ID
	a.Agent.Version = raw.Agent.Version
	a.Host.Name = raw.Host.Name
	a.Host.Hostname = raw.Host.Hostname
	a.Host.IPAddress = raw.Host.IPAddress
	a.ProtocolVersions = raw.ProtocolVersions
	a.Runtime.Name = raw.Runtime.Name
	a.Runtime.Version = raw.Runtime.Version
	a.Capabilities = raw.Capabilities
	a.Timestamp = raw.Timestamp

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (o *OperatorHello) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type MessageType `json:"type"`
		Session struct {
			ID string `json:"id"`
		} `json:"session"`
		ProtocolVersion string `json:"protocolVersion"`
		Configuration struct {
			HeartbeatInterval    string `json:"heartbeatInterval"`
			StateReportInterval  string `json:"stateReportInterval"`
			OfflineAfter         string `json:"offlineAfter"`
			MaxManifestSizeBytes int    `json:"maxManifestSizeBytes,omitempty"`
		} `json:"configuration"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Type:              o.Type,
		Session:           o.Session,
		ProtocolVersion:   o.ProtocolVersion,
		Configuration:     o.Configuration,
		Timestamp:         o.Timestamp,
	})
}

// UnmarshalJSON unmarshals JSON bytes into OperatorHello.
func (o *OperatorHello) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type MessageType `json:"type"`
		Session struct {
			ID string `json:"id"`
		} `json:"session"`
		ProtocolVersion string `json:"protocolVersion"`
		Configuration struct {
			HeartbeatInterval    string `json:"heartbeatInterval"`
			StateReportInterval  string `json:"stateReportInterval"`
			OfflineAfter         string `json:"offlineAfter"`
			MaxManifestSizeBytes int    `json:"maxManifestSizeBytes,omitempty"`
		} `json:"configuration"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	o.Type = raw.Type
	o.Session.ID = raw.Session.ID
	o.ProtocolVersion = raw.ProtocolVersion
	o.Configuration.HeartbeatInterval = raw.Configuration.HeartbeatInterval
	o.Configuration.StateReportInterval = raw.Configuration.StateReportInterval
	o.Configuration.OfflineAfter = raw.Configuration.OfflineAfter
	o.Configuration.MaxManifestSizeBytes = raw.Configuration.MaxManifestSizeBytes
	o.Timestamp = raw.Timestamp

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (e *EnrollmentRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type MessageType `json:"type"`
		Token  string      `json:"token"`
		Agent  struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		} `json:"agent"`
		Host struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		} `json:"host"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Type:      e.Type,
		Token:     e.Token,
		Agent:     e.Agent,
		Host:      e.Host,
		Timestamp: e.Timestamp,
	})
}

// UnmarshalJSON unmarshals JSON bytes into EnrollmentRequest.
func (e *EnrollmentRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type MessageType `json:"type"`
		Token  string      `json:"token"`
		Agent  struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		} `json:"agent"`
		Host struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		} `json:"host"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	e.Type = raw.Type
	e.Token = raw.Token
	e.Agent.ID = raw.Agent.ID
	e.Agent.Version = raw.Agent.Version
	e.Host.Name = raw.Host.Name
	e.Host.Hostname = raw.Host.Hostname
	e.Host.IPAddress = raw.Host.IPAddress
	e.Timestamp = raw.Timestamp

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (e *EnrollmentResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType `json:"type"`
		Success   bool        `json:"success"`
		Error     string      `json:"error,omitempty"`
		Identity  struct {
			Certificate string `json:"certificate"`
			PrivateKey  string `json:"privateKey"`
			CABundle    string `json:"caBundle"`
		} `json:"identity,omitempty"`
		SessionID string    `json:"sessionId,omitempty"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Type:      e.Type,
		Success:   e.Success,
		Error:     e.Error,
		Identity:  e.Identity,
		SessionID: e.SessionID,
		Timestamp: e.Timestamp,
	})
}

// UnmarshalJSON unmarshals JSON bytes into EnrollmentResponse.
func (e *EnrollmentResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType `json:"type"`
		Success   bool        `json:"success"`
		Error     string      `json:"error,omitempty"`
		Identity  struct {
			Certificate string `json:"certificate"`
			PrivateKey  string `json:"privateKey"`
			CABundle    string `json:"caBundle"`
		} `json:"identity,omitempty"`
		SessionID string    `json:"sessionId,omitempty"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	e.Type = raw.Type
	e.Success = raw.Success
	e.Error = raw.Error
	e.Identity.Certificate = raw.Identity.Certificate
	e.Identity.PrivateKey = raw.Identity.PrivateKey
	e.Identity.CABundle = raw.Identity.CABundle
	e.SessionID = raw.SessionID
	e.Timestamp = raw.Timestamp

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (h *Heartbeat) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
		Runtime   struct {
			Available bool   `json:"available"`
			Error     string `json:"error,omitempty"`
		} `json:"runtime"`
		Workloads struct {
			Assigned int `json:"assigned"`
			Ready    int `json:"ready"`
			Failed   int `json:"failed"`
			Updating int `json:"updating,omitempty"`
			Unknown  int `json:"unknown,omitempty"`
		} `json:"workloads"`
		Host struct {
			Name string `json:"name"`
		} `json:"host"`
	}{
		Type:      h.Type,
		SessionID: h.SessionID,
		Timestamp: h.Timestamp,
		Runtime:   h.Runtime,
		Workloads: h.Workloads,
		Host:      h.Host,
	})
}

// UnmarshalJSON unmarshals JSON bytes into Heartbeat.
func (h *Heartbeat) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
		Runtime   struct {
			Available bool   `json:"available"`
			Error     string `json:"error,omitempty"`
		} `json:"runtime"`
		Workloads struct {
			Assigned int `json:"assigned"`
			Ready    int `json:"ready"`
			Failed   int `json:"failed"`
			Updating int `json:"updating,omitempty"`
			Unknown  int `json:"unknown,omitempty"`
		} `json:"workloads"`
		Host struct {
			Name string `json:"name"`
		} `json:"host"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	h.Type = raw.Type
	h.SessionID = raw.SessionID
	h.Timestamp = raw.Timestamp
	h.Runtime.Available = raw.Runtime.Available
	h.Runtime.Error = raw.Runtime.Error
	h.Workloads.Assigned = raw.Workloads.Assigned
	h.Workloads.Ready = raw.Workloads.Ready
	h.Workloads.Failed = raw.Workloads.Failed
	h.Workloads.Updating = raw.Workloads.Updating
	h.Workloads.Unknown = raw.Workloads.Unknown
	h.Host.Name = raw.Host.Name

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (h *HostInventory) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
		Host      struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			OS        string `json:"os"`
			Kernel    string `json:"kernel"`
			CPU       string `json:"cpu"`
			Memory    string `json:"memory"`
			Storage   string `json:"storage"`
		} `json:"host"`
		Podman struct {
			Version    string `json:"version"`
			Rootful    bool   `json:"rootful"`
			APIVersion string `json:"apiVersion"`
		} `json:"podman"`
		Capabilities []string `json:"capabilities"`
	}{
		Type:         h.Type,
		SessionID:    h.SessionID,
		Timestamp:    h.Timestamp,
		Host:         h.Host,
		Podman:       h.Podman,
		Capabilities: h.Capabilities,
	})
}

// UnmarshalJSON unmarshals JSON bytes into HostInventory.
func (h *HostInventory) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
		Host      struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			OS        string `json:"os"`
			Kernel    string `json:"kernel"`
			CPU       string `json:"cpu"`
			Memory    string `json:"memory"`
			Storage   string `json:"storage"`
		} `json:"host"`
		Podman struct {
			Version    string `json:"version"`
			Rootful    bool   `json:"rootful"`
			APIVersion string `json:"apiVersion"`
		} `json:"podman"`
		Capabilities []string `json:"capabilities"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	h.Type = raw.Type
	h.SessionID = raw.SessionID
	h.Timestamp = raw.Timestamp
	h.Host.Name = raw.Host.Name
	h.Host.Hostname = raw.Host.Hostname
	h.Host.OS = raw.Host.OS
	h.Host.Kernel = raw.Host.Kernel
	h.Host.CPU = raw.Host.CPU
	h.Host.Memory = raw.Host.Memory
	h.Host.Storage = raw.Host.Storage
	h.Podman.Version = raw.Podman.Version
	h.Podman.Rootful = raw.Podman.Rootful
	h.Podman.APIVersion = raw.Podman.APIVersion
	h.Capabilities = raw.Capabilities

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (r *ReconciliationResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType    `json:"type"`
		SessionID string         `json:"sessionId"`
		Host      string         `json:"host"`
		Revision  int64          `json:"revision"`
		Timestamp time.Time      `json:"timestamp"`
		Workloads []WorkloadResult `json:"workloads"`
	}{
		Type:      r.Type,
		SessionID: r.SessionID,
		Host:      r.Host,
		Revision:  r.Revision,
		Timestamp: r.Timestamp,
		Workloads: r.Workloads,
	})
}

// UnmarshalJSON unmarshals JSON bytes into ReconciliationResult.
func (r *ReconciliationResult) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType      `json:"type"`
		SessionID string           `json:"sessionId"`
		Host      string           `json:"host"`
		Revision  int64            `json:"revision"`
		Timestamp time.Time        `json:"timestamp"`
		Workloads []WorkloadResult `json:"workloads"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	r.Type = raw.Type
	r.SessionID = raw.SessionID
	r.Host = raw.Host
	r.Revision = raw.Revision
	r.Timestamp = raw.Timestamp
	r.Workloads = raw.Workloads

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (w *WorkloadResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		UID            string `json:"uid"`
		Namespace      string `json:"namespace"`
		Name           string `json:"name"`
		Generation     int64  `json:"generation"`
		Phase          string `json:"phase"`
		Reason         string `json:"reason,omitempty"`
		Message        string `json:"message,omitempty"`
		ManifestDigest string `json:"manifestDigest"`
		Runtime        struct {
			PodID      string          `json:"podId,omitempty"`
			Containers []ContainerInfo `json:"containers,omitempty"`
		} `json:"runtime,omitempty"`
	}{
		UID:            w.UID,
		Namespace:      w.Namespace,
		Name:           w.Name,
		Generation:     w.Generation,
		Phase:          w.Phase,
		Reason:         w.Reason,
		Message:        w.Message,
		ManifestDigest: w.ManifestDigest,
		Runtime: struct {
			PodID      string          `json:"podId,omitempty"`
			Containers []ContainerInfo `json:"containers,omitempty"`
		}{
			PodID:      w.Runtime.PodID,
			Containers: w.Runtime.Containers,
		},
	})
}

// UnmarshalJSON unmarshals JSON bytes into WorkloadResult.
func (w *WorkloadResult) UnmarshalJSON(data []byte) error {
	var raw struct {
		UID            string `json:"uid"`
		Namespace      string `json:"namespace"`
		Name           string `json:"name"`
		Generation     int64  `json:"generation"`
		Phase          string `json:"phase"`
		Reason         string `json:"reason,omitempty"`
		Message        string `json:"message,omitempty"`
		ManifestDigest string `json:"manifestDigest"`
		Runtime        struct {
			PodID      string          `json:"podId,omitempty"`
			Containers []ContainerInfo `json:"containers,omitempty"`
		} `json:"runtime,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	w.UID = raw.UID
	w.Namespace = raw.Namespace
	w.Name = raw.Name
	w.Generation = raw.Generation
	w.Phase = raw.Phase
	w.Reason = raw.Reason
	w.Message = raw.Message
	w.ManifestDigest = raw.ManifestDigest
	w.Runtime.PodID = raw.Runtime.PodID
	w.Runtime.Containers = raw.Runtime.Containers

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (c *ContainerInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name  string `json:"name"`
		ID    string `json:"id"`
		State string `json:"state"`
	}{
		Name:  c.Name,
		ID:    c.ID,
		State: c.State,
	})
}

// UnmarshalJSON unmarshals JSON bytes into ContainerInfo.
func (c *ContainerInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name  string `json:"name"`
		ID    string `json:"id"`
		State string `json:"state"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Name = raw.Name
	c.ID = raw.ID
	c.State = raw.State

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (w *WorkloadStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
		UID       string      `json:"uid"`
		Namespace string      `json:"namespace"`
		Name      string      `json:"name"`
		Generation int64      `json:"generation"`
		Phase     string      `json:"phase"`
		Runtime   struct {
			PodID      string          `json:"podId,omitempty"`
			Containers []ContainerInfo `json:"containers,omitempty"`
		} `json:"runtime,omitempty"`
	}{
		Type:       w.Type,
		SessionID:  w.SessionID,
		Timestamp:  w.Timestamp,
		UID:        w.UID,
		Namespace:  w.Namespace,
		Name:       w.Name,
		Generation: w.Generation,
		Phase:      w.Phase,
		Runtime: struct {
			PodID      string          `json:"podId,omitempty"`
			Containers []ContainerInfo `json:"containers,omitempty"`
		}{
			PodID:      w.Runtime.PodID,
			Containers: w.Runtime.Containers,
		},
	})
}

// UnmarshalJSON unmarshals JSON bytes into WorkloadStatus.
func (w *WorkloadStatus) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
		UID       string      `json:"uid"`
		Namespace string      `json:"namespace"`
		Name      string      `json:"name"`
		Generation int64      `json:"generation"`
		Phase     string      `json:"phase"`
		Runtime   struct {
			PodID      string          `json:"podId,omitempty"`
			Containers []ContainerInfo `json:"containers,omitempty"`
		} `json:"runtime,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	w.Type = raw.Type
	w.SessionID = raw.SessionID
	w.Timestamp = raw.Timestamp
	w.UID = raw.UID
	w.Namespace = raw.Namespace
	w.Name = raw.Name
	w.Generation = raw.Generation
	w.Phase = raw.Phase
	w.Runtime.PodID = raw.Runtime.PodID
	w.Runtime.Containers = raw.Runtime.Containers

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (c *CertificateRotationRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
	}{
		Type:      c.Type,
		SessionID: c.SessionID,
		Timestamp: c.Timestamp,
	})
}

// UnmarshalJSON unmarshals JSON bytes into CertificateRotationRequest.
func (c *CertificateRotationRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType `json:"type"`
		SessionID string      `json:"sessionId"`
		Timestamp time.Time   `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Type = raw.Type
	c.SessionID = raw.SessionID
	c.Timestamp = raw.Timestamp

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (d *DesiredState) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType `json:"type"`
		Host      string      `json:"host"`
		Revision  int64       `json:"revision"`
		Timestamp time.Time   `json:"timestamp"`
		PodmanPlays []PodmanPlay `json:"podmanPlays"`
	}{
		Type:        d.Type,
		Host:        d.Host,
		Revision:    d.Revision,
		Timestamp:   d.Timestamp,
		PodmanPlays: d.PodmanPlays,
	})
}

// UnmarshalJSON unmarshals JSON bytes into DesiredState.
func (d *DesiredState) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType    `json:"type"`
		Host      string         `json:"host"`
		Revision  int64          `json:"revision"`
		Timestamp time.Time      `json:"timestamp"`
		PodmanPlays []PodmanPlay `json:"podmanPlays"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	d.Type = raw.Type
	d.Host = raw.Host
	d.Revision = raw.Revision
	d.Timestamp = raw.Timestamp
	d.PodmanPlays = raw.PodmanPlays

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (p *PodmanPlay) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		UID        string `json:"uid"`
		Namespace  string `json:"namespace"`
		Name       string `json:"name"`
		Generation int64  `json:"generation"`
		Manifest   string `json:"manifest"`
	}{
		UID:        p.UID,
		Namespace:  p.Namespace,
		Name:       p.Name,
		Generation: p.Generation,
		Manifest:   p.Manifest,
	})
}

// UnmarshalJSON unmarshals JSON bytes into PodmanPlay.
func (p *PodmanPlay) UnmarshalJSON(data []byte) error {
	var raw struct {
		UID        string `json:"uid"`
		Namespace  string `json:"namespace"`
		Name       string `json:"name"`
		Generation int64  `json:"generation"`
		Manifest   string `json:"manifest"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.UID = raw.UID
	p.Namespace = raw.Namespace
	p.Name = raw.Name
	p.Generation = raw.Generation
	p.Manifest = raw.Manifest

	return nil
}

// MarshalJSON marshals the message to JSON bytes.
func (d *DesiredStateUpdate) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      MessageType  `json:"type"`
		SessionID string       `json:"sessionId"`
		Host      string       `json:"host"`
		Revision  int64        `json:"revision"`
		Timestamp time.Time    `json:"timestamp"`
		Additions []PodmanPlay `json:"additions,omitempty"`
		Deletions []string     `json:"deletions,omitempty"`
	}{
		Type:      d.Type,
		SessionID: d.SessionID,
		Host:      d.Host,
		Revision:  d.Revision,
		Timestamp: d.Timestamp,
		Additions: d.Additions,
		Deletions: d.Deletions,
	})
}

// UnmarshalJSON unmarshals JSON bytes into DesiredStateUpdate.
func (d *DesiredStateUpdate) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      MessageType  `json:"type"`
		SessionID string       `json:"sessionId"`
		Host      string       `json:"host"`
		Revision  int64        `json:"revision"`
		Timestamp time.Time    `json:"timestamp"`
		Additions []PodmanPlay `json:"additions,omitempty"`
		Deletions []string     `json:"deletions,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	d.Type = raw.Type
	d.SessionID = raw.SessionID
	d.Host = raw.Host
	d.Revision = raw.Revision
	d.Timestamp = raw.Timestamp
	d.Additions = raw.Additions
	d.Deletions = raw.Deletions

	return nil
}

// MarshalJSON marshals the ProtocolError to JSON bytes.
func (p *ProtocolError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    p.Code,
		Message: p.Message,
	})
}

// UnmarshalJSON unmarshals JSON bytes into ProtocolError.
func (p *ProtocolError) UnmarshalJSON(data []byte) error {
	var raw struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.Code = raw.Code
	p.Message = raw.Message

	return nil
}

// String returns a string representation of the message type.
func (t MessageType) String() string {
	return string(t)
}

// String returns a string representation of the ProtocolError.
func (p *ProtocolError) String() string {
	return fmt.Sprintf("%s: %s", p.Code, p.Message)
}
