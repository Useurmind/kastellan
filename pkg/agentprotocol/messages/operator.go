// Package messages defines the protocol message structures for the Kastellan Agent Protocol.
package messages

import (
	"time"

	agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)

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

// Validate performs basic validation on DesiredState.
func (d *DesiredState) Validate() error {
	if d.Host == "" {
		return &ProtocolError{Code: "missing_host", Message: "host name is required"}
	}
	return nil
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

// ToProto converts OperatorHello to proto OperatorHello.
func (o *OperatorHello) ToProto() *agentv1alpha1.OperatorHello {
	return &agentv1alpha1.OperatorHello{
		SessionId:                 o.Session.ID,
		SelectedProtocol:          &agentv1alpha1.ProtocolVersion{},
		HeartbeatIntervalSeconds:  0,
		StateReportIntervalSeconds: 0,
		ServerTimeUnix:            0,
	}
}

// OperatorHelloFromProto converts proto OperatorHello to OperatorHello.
func OperatorHelloFromProto(p *agentv1alpha1.OperatorHello) *OperatorHello {
	if p == nil {
		return nil
	}
	return &OperatorHello{
		Type: MessageTypeOperatorHello,
	}
}

// ToProto converts DesiredState to proto DesiredState.
func (d *DesiredState) ToProto() *agentv1alpha1.DesiredState {
	protoPodmanPlays := make([]*agentv1alpha1.PodmanPlayAssignment, len(d.PodmanPlays))
	for i, p := range d.PodmanPlays {
		protoPodmanPlays[i] = p.ToProto()
	}
	return &agentv1alpha1.DesiredState{
		Host:        d.Host,
		Revision:    uint64(d.Revision),
		PodmanPlays: protoPodmanPlays,
	}
}

// DesiredStateFromProto converts proto DesiredState to DesiredState.
func DesiredStateFromProto(p *agentv1alpha1.DesiredState) *DesiredState {
	if p == nil {
		return nil
	}
	podmanPlays := make([]PodmanPlay, len(p.GetPodmanPlays()))
	for i, pp := range p.GetPodmanPlays() {
		podmanPlays[i] = *PodmanPlayFromProto(pp)
	}
	return &DesiredState{
		Type:      MessageTypeDesiredState,
		Host:      p.GetHost(),
		Revision:  int64(p.GetRevision()),
		PodmanPlays: podmanPlays,
	}
}

// ToProto converts ReconcileRequest to proto ReconcileRequest.
func (r *ReconcileRequest) ToProto() *agentv1alpha1.ReconcileRequest {
	return &agentv1alpha1.ReconcileRequest{
		SessionId: r.SessionID,
		Host:      r.Host,
		Revision:  uint64(r.Revision),
	}
}

// ReconcileRequestFromProto converts proto ReconcileRequest to ReconcileRequest.
func ReconcileRequestFromProto(p *agentv1alpha1.ReconcileRequest) *ReconcileRequest {
	if p == nil {
		return nil
	}
	return &ReconcileRequest{
		Type:      MessageTypeReconcileRequest,
		SessionID: p.GetSessionId(),
		Host:      p.GetHost(),
		Revision:  int64(p.GetRevision()),
	}
}

// ToProto converts DesiredStateUpdate to proto DesiredStateUpdate.
func (d *DesiredStateUpdate) ToProto() *agentv1alpha1.DesiredStateUpdate {
	protoAdditions := make([]*agentv1alpha1.PodmanPlayAssignment, len(d.Additions))
	for i, a := range d.Additions {
		protoAdditions[i] = a.ToProto()
	}
	return &agentv1alpha1.DesiredStateUpdate{
		SessionId:  d.SessionID,
		Host:       d.Host,
		Revision:   uint64(d.Revision),
		Additions:  protoAdditions,
		Deletions:  d.Deletions,
	}
}

// DesiredStateUpdateFromProto converts proto DesiredStateUpdate to DesiredStateUpdate.
func DesiredStateUpdateFromProto(p *agentv1alpha1.DesiredStateUpdate) *DesiredStateUpdate {
	if p == nil {
		return nil
	}
	additions := make([]PodmanPlay, len(p.GetAdditions()))
	for i, a := range p.GetAdditions() {
		additions[i] = *PodmanPlayFromProto(a)
	}
	return &DesiredStateUpdate{
		Type:      MessageTypeDesiredStateUpdate,
		SessionID: p.GetSessionId(),
		Host:      p.GetHost(),
		Revision:  int64(p.GetRevision()),
		Additions: additions,
		Deletions: p.GetDeletions(),
	}
}
