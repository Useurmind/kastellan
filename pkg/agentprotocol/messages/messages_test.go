package messages_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/useurmind/kastellan/pkg/agentprotocol/messages"
)

func TestAgentHello(t *testing.T) {
	// Create AgentHello message
	agentHello := &messages.AgentHello{
		Type: messages.MessageTypeAgentHello,
		Agent: struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		}{
			ID:      "agent-2fd317",
			Version: "0.1.0",
		},
		Host: struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		}{
			Name:     "lb01",
			Hostname: "lb01.example.internal",
		},
		ProtocolVersions: []string{"v1alpha1"},
		Runtime: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "podman",
			Version: "5.6.0",
		},
		Capabilities: []string{"play-kube", "replace", "configmap", "secret", "host-path"},
		Timestamp:    time.Now(),
	}

	// Test validation
	if err := agentHello.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	// Test JSON marshaling
	data, err := json.Marshal(agentHello)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.AgentHello
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled.Agent.ID != agentHello.Agent.ID {
		t.Errorf("Agent.ID mismatch: got %v, want %v", unmarshaled.Agent.ID, agentHello.Agent.ID)
	}
	if unmarshaled.Host.Name != agentHello.Host.Name {
		t.Errorf("Host.Name mismatch: got %v, want %v", unmarshaled.Host.Name, agentHello.Host.Name)
	}
	if len(unmarshaled.ProtocolVersions) != len(agentHello.ProtocolVersions) {
		t.Errorf("ProtocolVersions length mismatch: got %v, want %v", len(unmarshaled.ProtocolVersions), len(agentHello.ProtocolVersions))
	}
}

func TestOperatorHello(t *testing.T) {
	// Create OperatorHello message
	operatorHello := &messages.OperatorHello{
		Type: messages.MessageTypeOperatorHello,
		Session: struct {
			ID string `json:"id"`
		}{
			ID: "session-844d12",
		},
		ProtocolVersion: "v1alpha1",
		Configuration: struct {
			HeartbeatInterval    string `json:"heartbeatInterval"`
			StateReportInterval  string `json:"stateReportInterval"`
			OfflineAfter         string `json:"offlineAfter"`
			MaxManifestSizeBytes int    `json:"maxManifestSizeBytes,omitempty"`
		}{
			HeartbeatInterval:    "30s",
			StateReportInterval:  "60s",
			OfflineAfter:         "2m",
			MaxManifestSizeBytes: 1048576,
		},
		Timestamp: time.Now(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(operatorHello)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.OperatorHello
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled.Session.ID != operatorHello.Session.ID {
		t.Errorf("Session.ID mismatch: got %v, want %v", unmarshaled.Session.ID, operatorHello.Session.ID)
	}
	if unmarshaled.ProtocolVersion != operatorHello.ProtocolVersion {
		t.Errorf("ProtocolVersion mismatch: got %v, want %v", unmarshaled.ProtocolVersion, operatorHello.ProtocolVersion)
	}
}

func TestEnrollmentRequest(t *testing.T) {
	// Create EnrollmentRequest message
	enrollmentRequest := &messages.EnrollmentRequest{
		Type:  messages.MessageTypeEnrollmentRequest,
		Token: "test-token-123",
		Agent: struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		}{
			ID:      "agent-2fd317",
			Version: "0.1.0",
		},
		Host: struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		}{
			Name:     "lb01",
			Hostname: "lb01.example.internal",
		},
		Timestamp: time.Now(),
	}

	// Test validation
	if err := enrollmentRequest.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	// Test JSON marshaling
	data, err := json.Marshal(enrollmentRequest)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.EnrollmentRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled.Token != enrollmentRequest.Token {
		t.Errorf("Token mismatch: got %v, want %v", unmarshaled.Token, enrollmentRequest.Token)
	}
	if unmarshaled.Host.Name != enrollmentRequest.Host.Name {
		t.Errorf("Host.Name mismatch: got %v, want %v", unmarshaled.Host.Name, enrollmentRequest.Host.Name)
	}
}

func TestHeartbeat(t *testing.T) {
	// Create Heartbeat message
	heartbeat := &messages.Heartbeat{
		Type:      messages.MessageTypeHeartbeat,
		SessionID: "session-844d12",
		Timestamp: time.Now(),
		Runtime: struct {
			Available bool   `json:"available"`
			Error     string `json:"error,omitempty"`
		}{
			Available: true,
		},
		Workloads: struct {
			Assigned int `json:"assigned"`
			Ready    int `json:"ready"`
			Failed   int `json:"failed"`
			Updating int `json:"updating,omitempty"`
			Unknown  int `json:"unknown,omitempty"`
		}{
			Assigned: 3,
			Ready:    3,
			Failed:   0,
		},
		Host: struct {
			Name string `json:"name"`
		}{
			Name: "lb01",
		},
	}

	// Test validation
	if err := heartbeat.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	// Test JSON marshaling
	data, err := json.Marshal(heartbeat)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.Heartbeat
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled.SessionID != heartbeat.SessionID {
		t.Errorf("SessionID mismatch: got %v, want %v", unmarshaled.SessionID, heartbeat.SessionID)
	}
	if unmarshaled.Workloads.Ready != heartbeat.Workloads.Ready {
		t.Errorf("Workloads.Ready mismatch: got %v, want %v", unmarshaled.Workloads.Ready, heartbeat.Workloads.Ready)
	}
}

func TestReconciliationResult(t *testing.T) {
	// Create ReconciliationResult message
	result := &messages.ReconciliationResult{
		Type:      messages.MessageTypeReconciliationResult,
		SessionID: "session-844d12",
		Host:      "lb01",
		Revision:  42,
		Timestamp: time.Now(),
		Workloads: []messages.WorkloadResult{
			{
				UID:            "0f831f62-342a-4add-b05c-c968ec71b679",
				Namespace:      "infrastructure",
				Name:           "haproxy",
				Generation:     4,
				Phase:          "Ready",
				ManifestDigest: "sha256:93f3d2...",
				Runtime: struct {
					PodID      string                   `json:"podId,omitempty"`
					Containers []messages.ContainerInfo `json:"containers,omitempty"`
				}{
					PodID: "12d769d4...",
					Containers: []messages.ContainerInfo{
						{
							Name:  "haproxy",
							ID:    "fdd471ee...",
							State: "running",
						},
					},
				},
			},
		},
	}

	// Test validation
	if err := result.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.ReconciliationResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled.Host != result.Host {
		t.Errorf("Host mismatch: got %v, want %v", unmarshaled.Host, result.Host)
	}
	if unmarshaled.Revision != result.Revision {
		t.Errorf("Revision mismatch: got %v, want %v", unmarshaled.Revision, result.Revision)
	}
	if len(unmarshaled.Workloads) != len(result.Workloads) {
		t.Errorf("Workloads length mismatch: got %v, want %v", len(unmarshaled.Workloads), len(result.Workloads))
	}
}

func TestContainerInfo(t *testing.T) {
	// Create ContainerInfo message
	container := &messages.ContainerInfo{
		Name:  "haproxy",
		ID:    "fdd471ee-1234-5678-9abc-def012345678",
		State: "running",
	}

	// Test JSON marshaling
	data, err := json.Marshal(container)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.ContainerInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled.Name != container.Name {
		t.Errorf("Name mismatch: got %v, want %v", unmarshaled.Name, container.Name)
	}
	if unmarshaled.ID != container.ID {
		t.Errorf("ID mismatch: got %v, want %v", unmarshaled.ID, container.ID)
	}
	if unmarshaled.State != container.State {
		t.Errorf("State mismatch: got %v, want %v", unmarshaled.State, container.State)
	}
}

func TestDesiredState(t *testing.T) {
	// Create DesiredState message
	desiredState := &messages.DesiredState{
		Type:      messages.MessageTypeDesiredState,
		Host:      "lb01",
		Revision:  42,
		Timestamp: time.Now(),
		PodmanPlays: []messages.PodmanPlay{
			{
				UID:        "0f831f62-342a-4add-b05c-c968ec71b679",
				Namespace:  "infrastructure",
				Name:       "haproxy",
				Generation: 4,
				Manifest: `apiVersion: v1
kind: Pod
metadata:
  name: haproxy
spec:
  containers:
    - name: haproxy
      image: docker.io/library/haproxy:lts`,
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(desiredState)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.DesiredState
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled.Host != desiredState.Host {
		t.Errorf("Host mismatch: got %v, want %v", unmarshaled.Host, desiredState.Host)
	}
	if unmarshaled.Revision != desiredState.Revision {
		t.Errorf("Revision mismatch: got %v, want %v", unmarshaled.Revision, desiredState.Revision)
	}
	if len(unmarshaled.PodmanPlays) != len(desiredState.PodmanPlays) {
		t.Errorf("PodmanPlays length mismatch: got %v, want %v", len(unmarshaled.PodmanPlays), len(desiredState.PodmanPlays))
	}
}

func TestProtocolError(t *testing.T) {
	// Create ProtocolError
	err := &messages.ProtocolError{
		Code:    "invalid_agent_id",
		Message: "agent ID is required",
	}

	// Test Error() method
	if err.Error() == "" {
		t.Error("Error() returned empty string")
	}

	// Test JSON marshaling
	marshaledData, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Errorf("MarshalJSON() error = %v", marshalErr)
	}

	// Test JSON unmarshaling
	var unmarshaled messages.ProtocolError
	unmarshalErr := json.Unmarshal(marshaledData, &unmarshaled)
	if unmarshalErr != nil {
		t.Errorf("UnmarshalJSON() error = %v", unmarshalErr)
	}

	// Verify unmarshaled data
	if unmarshaled.Code != err.Code {
		t.Errorf("Code mismatch: got %v, want %v", unmarshaled.Code, err.Code)
	}
	if unmarshaled.Message != err.Message {
		t.Errorf("Message mismatch: got %v, want %v", unmarshaled.Message, err.Message)
	}
}
