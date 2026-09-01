// Package agentprotocol provides constants for the Kastellan Agent Protocol.
package agentprotocol

import (
	"time"
)

// Protocol version constants.
const (
	ProtocolVersionV1Alpha1 = "v1alpha1"
	ProtocolVersionV1       = "v1"
)

// Default configuration values.
const (
	DefaultHeartbeatInterval     = 30 * time.Second
	DefaultStateReportInterval   = 60 * time.Second
	DefaultOfflineAfter          = 2 * time.Minute
	DefaultReconnectInitialDelay = time.Second
	DefaultReconnectMaxDelay     = time.Minute
	DefaultMaxManifestSizeBytes  = 1024 * 1024 // 1MB
)

// Default server port.
const (
	DefaultServerPort = 443
)

// Default server address.
const (
	DefaultServerAddress = "localhost:443"
)

// Workload phase constants.
const (
	PhasePending  = "Pending"
	PhaseRunning  = "Running"
	PhaseReady    = "Ready"
	PhaseFailed   = "Failed"
	PhaseUpdating = "Updating"
	PhaseUnknown  = "Unknown"
)

// Workload status constants.
const (
	StatusConnected    = "Connected"
	StatusDisconnected = "Disconnected"
	StatusReady        = "Ready"
	StatusOffline      = "Offline"
)

// Capability constants.
const (
	CapabilityPlayKube  = "play-kube"
	CapabilityReplace   = "replace"
	CapabilityConfigMap = "configmap"
	CapabilitySecret    = "secret"
	CapabilityHostPath  = "host-path"
)

// Runtime constants.
const (
	RuntimePodman = "podman"
)

// Annotation constants for workload identification.
const (
	AnnotationManaged     = "kastellan.io/managed"
	AnnotationNamespace   = "kastellan.io/namespace"
	AnnotationWorkload    = "kastellan.io/workload"
	AnnotationWorkloadUID = "kastellan.io/workload-uid"
	AnnotationGeneration  = "kastellan.io/generation"
)

// Label constants for workload identification.
const (
	LabelManaged     = "kastellan.io/managed"
	LabelNamespace   = "kastellan.io/namespace"
	LabelWorkload    = "kastellan.io/workload"
	LabelWorkloadUID = "kastellan.io/workload-uid"
	LabelGeneration  = "kastellan.io/generation"
)

// State directory paths.
const (
	StateDirectory          = "/var/lib/kastellan"
	IdentityDirectory       = StateDirectory + "/identity"
	StateStateDirectory     = StateDirectory + "/state"
	StateWorkloadsDirectory = StateDirectory + "/workloads"
)

// Identity file paths.
const (
	AgentIDFile     = IdentityDirectory + "/agent-id"
	CertificateFile = IdentityDirectory + "/certificate.pem"
	PrivateKeyFile  = IdentityDirectory + "/private-key.pem"
	CABundleFile    = IdentityDirectory + "/ca-bundle.pem"
)

// State file paths.
const (
	LastReceivedRevisionFile = StateStateDirectory + "/last-received-revision"
	LastAppliedRevisionFile  = StateStateDirectory + "/last-applied-revision"
)

// Workload state file paths.
const (
	ManifestFile       = "manifest.yaml"
	ManifestDigestFile = "manifest.sha256"
	MetadataFile       = "metadata.json"
	StatusFile         = "status.json"
)

// Message type constants.
const (
	MessageTypeAgentHello                  = "AgentHello"
	MessageTypeEnrollmentRequest           = "EnrollmentRequest"
	MessageTypeHeartbeat                   = "Heartbeat"
	MessageTypeHostInventory               = "HostInventory"
	MessageTypeReconciliationResult        = "ReconciliationResult"
	MessageTypeWorkloadStatus              = "WorkloadStatus"
	MessageTypeCertificateRotationRequest  = "CertificateRotationRequest"
	MessageTypeOperatorHello               = "OperatorHello"
	MessageTypeEnrollmentResponse          = "EnrollmentResponse"
	MessageTypeDesiredState                = "DesiredState"
	MessageTypeDesiredStateUpdate          = "DesiredStateUpdate"
	MessageTypeReconcileRequest            = "ReconcileRequest"
	MessageTypeCertificateRotationResponse = "CertificateRotationResponse"
	MessageTypeConnectionClose             = "ConnectionClose"
)

// Error type constants.
const (
	ErrorTypeTransient = "transient"
	ErrorTypePermanent = "permanent"
)

// Reconnection policy constants.
const (
	ReconnectPolicyExponential = "exponential"
	ReconnectPolicyLinear      = "linear"
	ReconnectPolicyFixed       = "fixed"
)

// Authentication type constants.
const (
	AuthTypeEnrollmentToken = "enrollment-token"
	AuthTypeMTLS            = "mtls"
)

// Manifest size limits.
const (
	MinManifestSizeBytes = 1024        // 1KB
	MaxManifestSizeBytes = 1024 * 1024 // 1MB
)

// Timeout constants.
const (
	EnrollmentTimeout     = 30 * time.Second
	ConnectTimeout        = 10 * time.Second
	HeartbeatTimeout      = 5 * time.Second
	ReconciliationTimeout = 5 * time.Minute
)

// Retry constants.
const (
	MaxEnrollmentRetries   = 3
	MaxReconnectionRetries = 0 // 0 means infinite
	MaxHeartbeatRetries    = 3
)
