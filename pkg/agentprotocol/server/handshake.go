package server

import (
	"fmt"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol"
	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// Protocol version constant.
const ProtocolVersionV1Alpha1 = "v1alpha1"

// Error codes for server errors.
const (
	ErrorCodeSessionExists           = "session_exists"
	ErrorCodeInvalidCertificate      = "invalid_certificate"
	ErrorCodeHostNotFound            = "host_not_found"
	ErrorCodeHostDisabled            = "host_disabled"
	ErrorCodeProtocolVersionMismatch = "protocol_version_mismatch"
	ErrorCodeEnrollmentTokenInvalid  = "enrollment_token_invalid"
	ErrorCodeEnrollmentTokenExpired  = "enrollment_token_expired"
	ErrorCodeInvalidAgentID          = "invalid_agent_id"
	ErrorCodeInvalidAgentVersion     = "invalid_agent_version"
	ErrorCodeInvalidHostName         = "invalid_host_name"
	ErrorCodeMissingProtocolVersions = "missing_protocol_versions"
	ErrorCodeRuntimeUnavailable      = "runtime_unavailable"
)

// ExternalHostResolver resolves agent certificates to ExternalHost names.
type ExternalHostResolver interface {
	ResolveHost(cert *messages.AgentHello) (string, error)
	IsHostEnabled(hostName string) (bool, error)
}

// EnrollmentTokenStore stores and validates enrollment tokens.
type EnrollmentTokenStore interface {
	ValidateToken(token string) (string, error)
	InvalidateToken(token string) error
}

// HandshakeHandler handles the AgentHello/OperatorHello handshake.
type HandshakeHandler struct {
	externalHosts   ExternalHostResolver
	enrollmentStore EnrollmentTokenStore
}

// NewHandshakeHandler creates a new handshake handler.
func NewHandshakeHandler() *HandshakeHandler {
	return &HandshakeHandler{}
}

// SetExternalHostResolver sets the external host resolver.
func (h *HandshakeHandler) SetExternalHostResolver(resolver ExternalHostResolver) {
	h.externalHosts = resolver
}

// SetEnrollmentStore sets the enrollment token store.
func (h *HandshakeHandler) SetEnrollmentStore(store EnrollmentTokenStore) {
	h.enrollmentStore = store
}

// HandleAgentHello processes an AgentHello message and returns an OperatorHello response.
func (h *HandshakeHandler) HandleAgentHello(agentHello interface{}) (*messages.OperatorHello, error) {
	// Type assertion to AgentHello
	ah, ok := agentHello.(*messages.AgentHello)
	if !ok {
		return nil, agentprotocol.NewProtocolError(
			ErrorCodeInvalidAgentID,
			"invalid AgentHello type",
		)
	}

	// Validate AgentHello message
	if err := ah.Validate(); err != nil {
		return nil, err
	}

	hostName := ah.Host.Name

	// Check if host is enabled (if resolver is available)
	if h.externalHosts != nil {
		enabled, err := h.externalHosts.IsHostEnabled(hostName)
		if err != nil {
			return nil, agentprotocol.NewProtocolError(
				ErrorCodeHostNotFound,
				fmt.Sprintf("failed to check host status: %v", err),
			)
		}

		if !enabled {
			return nil, agentprotocol.NewAuthorizationError(fmt.Sprintf("host %s is disabled", hostName))
		}
	}

	// Validate runtime (Podman)
	if ah.Runtime.Name != messages.RuntimePodman {
		return nil, agentprotocol.NewProtocolError(
			ErrorCodeRuntimeUnavailable,
			fmt.Sprintf("unsupported runtime: %s (expected %s)", ah.Runtime.Name, messages.RuntimePodman),
		)
	}

	// Validate protocol version
	if !h.isProtocolSupported(ah.ProtocolVersions) {
		return nil, agentprotocol.NewProtocolError(
			ErrorCodeProtocolVersionMismatch,
			fmt.Sprintf("no compatible protocol version. Agent supports: %v", ah.ProtocolVersions),
		)
	}

	// Create response
	response := &messages.OperatorHello{
		Type:            messages.MessageTypeOperatorHello,
		ProtocolVersion: ProtocolVersionV1Alpha1,
		Configuration: struct {
			HeartbeatInterval    string `json:"heartbeatInterval"`
			StateReportInterval  string `json:"stateReportInterval"`
			OfflineAfter         string `json:"offlineAfter"`
			MaxManifestSizeBytes int    `json:"maxManifestSizeBytes,omitempty"`
		}{
			HeartbeatInterval:    "30s",
			StateReportInterval:  "60s",
			OfflineAfter:         "2m",
			MaxManifestSizeBytes: 1024 * 1024,
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// isProtocolSupported checks if any of the agent's protocol versions are supported.
func (h *HandshakeHandler) isProtocolSupported(agentVersions []string) bool {
	supported := []string{
		ProtocolVersionV1Alpha1,
	}

	for _, agentVer := range agentVersions {
		for _, supportedVer := range supported {
			if agentVer == supportedVer {
				return true
			}
		}
	}

	return false
}
