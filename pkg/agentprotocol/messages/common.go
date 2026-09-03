// Package messages defines the protocol message structures for the Kastellan Agent Protocol.
package messages

import (
	"fmt"

	agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)

// Capability constants
const (
	CapabilityPlayKube  = "play-kube"
	CapabilityReplace   = "replace"
	CapabilityConfigMap = "configmap"
	CapabilitySecret    = "secret"
	CapabilityHostPath  = "host-path"
)

// Runtime constants
const (
	RuntimePodman = "podman"
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
	MessageTypeOperatorHello               MessageType = "OperatorHello"
	MessageTypeEnrollmentResponse          MessageType = "EnrollmentResponse"
	MessageTypeDesiredState                MessageType = "DesiredState"
	MessageTypeDesiredStateUpdate          MessageType = "DesiredStateUpdate"
	MessageTypeReconcileRequest            MessageType = "ReconcileRequest"
	MessageTypeCertificateRotationResponse MessageType = "CertificateRotationResponse"
	MessageTypeConnectionClose             MessageType = "ConnectionClose"
)

// ProtocolError represents a protocol-level error.
type ProtocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns the error message.
func (e *ProtocolError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ContainerInfo represents information about a container.
type ContainerInfo struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	State string `json:"state"`
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

// Validate performs basic validation on ContainerInfo.
func (c *ContainerInfo) Validate() error {
	if c.Name == "" {
		return &ProtocolError{Code: "missing_name", Message: "container name is required"}
	}
	if c.ID == "" {
		return &ProtocolError{Code: "missing_id", Message: "container ID is required"}
	}
	return nil
}

// Validate performs basic validation on WorkloadResult.
func (w *WorkloadResult) Validate() error {
	if w.UID == "" {
		return &ProtocolError{Code: "missing_uid", Message: "workload UID is required"}
	}
	if w.Name == "" {
		return &ProtocolError{Code: "missing_name", Message: "workload name is required"}
	}
	if w.ManifestDigest == "" {
		return &ProtocolError{Code: "missing_manifest_digest", Message: "manifest digest is required"}
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

// ToProto converts ContainerInfo to proto ContainerInfo.
func (c *ContainerInfo) ToProto() *agentv1alpha1.ContainerInfo {
	return &agentv1alpha1.ContainerInfo{
		Name:  c.Name,
		Id:    c.ID,
		State: c.State,
	}
}

// ContainerInfoFromProto converts proto ContainerInfo to ContainerInfo.
func ContainerInfoFromProto(p *agentv1alpha1.ContainerInfo) *ContainerInfo {
	if p == nil {
		return nil
	}
	return &ContainerInfo{
		Name:  p.Name,
		ID:    p.Id,
		State: p.State,
	}
}

// ToProto converts WorkloadResult to proto WorkloadResult.
func (w *WorkloadResult) ToProto() *agentv1alpha1.WorkloadResult {
	protoContainers := make([]*agentv1alpha1.ContainerInfo, len(w.Runtime.Containers))
	for i, c := range w.Runtime.Containers {
		protoContainers[i] = c.ToProto()
	}
	return &agentv1alpha1.WorkloadResult{
		Identity: &agentv1alpha1.ResourceIdentity{
			ApiVersion: "",
			Kind:       "",
			Namespace:  w.Namespace,
			Name:       w.Name,
			Uid:        w.UID,
			Generation: w.Generation,
		},
		Phase:          w.Phase,
		Error:          nil,
		ManifestDigest: w.ManifestDigest,
		Containers:     protoContainers,
	}
}

// WorkloadResultFromProto converts proto WorkloadResult to WorkloadResult.
func WorkloadResultFromProto(p *agentv1alpha1.WorkloadResult) *WorkloadResult {
	if p == nil {
		return nil
	}
	identity := p.GetIdentity()
	return &WorkloadResult{
		UID:        identity.GetUid(),
		Namespace:  identity.GetNamespace(),
		Name:       identity.GetName(),
		Generation: identity.GetGeneration(),
		Phase:      p.GetPhase(),
		ManifestDigest: p.GetManifestDigest(),
	}
}

// ToProto converts PodmanPlay to proto PodmanPlayAssignment.
func (p *PodmanPlay) ToProto() *agentv1alpha1.PodmanPlayAssignment {
	return &agentv1alpha1.PodmanPlayAssignment{
		Identity: &agentv1alpha1.ResourceIdentity{
			ApiVersion: "",
			Kind:       "",
			Namespace:  p.Namespace,
			Name:       p.Name,
			Uid:        p.UID,
			Generation: p.Generation,
		},
		Manifest:       p.Manifest,
		ManifestDigest: "",
		UpdateStrategy: "",
	}
}

// PodmanPlayFromProto converts proto PodmanPlayAssignment to PodmanPlay.
func PodmanPlayFromProto(p *agentv1alpha1.PodmanPlayAssignment) *PodmanPlay {
	if p == nil {
		return nil
	}
	identity := p.GetIdentity()
	return &PodmanPlay{
		UID:        identity.GetUid(),
		Namespace:  identity.GetNamespace(),
		Name:       identity.GetName(),
		Generation: identity.GetGeneration(),
		Manifest:   p.GetManifest(),
	}
}
