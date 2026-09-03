package server

import (
	"testing"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// Test phase constants.
const (
	PhasePending  = "Pending"
	PhaseRunning  = "Running"
	PhaseReady    = "Ready"
	PhaseFailed   = "Failed"
	PhaseUpdating = "Updating"
	PhaseUnknown  = "Unknown"
)

// TestSessionManager_CreateSession tests session creation.
func TestSessionManager_CreateSession(t *testing.T) {
	sm := NewSessionManager()
	session := NewSession("test-session", "test-host", "test-agent", "1.0.0")

	err := sm.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify session was created
	retrieved, exists := sm.GetSession("test-session")
	if !exists {
		t.Fatal("Session not found")
	}

	if retrieved.ID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", retrieved.ID)
	}

	if retrieved.ExternalHost != "test-host" {
		t.Errorf("Expected host 'test-host', got '%s'", retrieved.ExternalHost)
	}
}

// TestSessionManager_GetSessionByHost tests session lookup by host.
func TestSessionManager_GetSessionByHost(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", "host-1", "agent-1", "1.0.0")
	sm.CreateSession(session)

	retrieved, exists := sm.GetSessionByHost("host-1")
	if !exists {
		t.Fatal("Session not found by host")
	}

	if retrieved.ID != "session-1" {
		t.Errorf("Expected session ID 'session-1', got '%s'", retrieved.ID)
	}
}

// TestSession_UpdateHeartbeat tests heartbeat updates.
func TestSession_UpdateHeartbeat(t *testing.T) {
	session := NewSession("test", "host", "agent", "1.0.0")

	session.UpdateHeartbeat(struct {
		Assigned int
		Ready    int
		Failed   int
		Updating int
		Unknown  int
	}{
		Assigned: 3,
		Ready:    2,
		Failed:   1,
		Updating: 0,
		Unknown:  0,
	}, true)

	if !session.HostStatus.RuntimeAvailable {
		t.Error("Expected RuntimeAvailable to be true")
	}

	if session.HostStatus.Workloads.Ready != 2 {
		t.Errorf("Expected 2 ready workloads, got %d", session.HostStatus.Workloads.Ready)
	}
}

// TestDesiredStateProvider_SetHostWorkloads tests setting host workloads.
func TestDesiredStateProvider_SetHostWorkloads(t *testing.T) {
	p := NewDesiredStateProvider()

	workloads := map[string]*messages.PodmanPlay{
		"workload-1": {
			UID:        "workload-1",
			Namespace:  "default",
			Name:       "test-workload",
			Generation: 1,
			Manifest:   "apiVersion: v1\nkind: Pod\n",
		},
	}

	err := p.SetHostWorkloads("test-host", workloads)
	if err != nil {
		t.Fatalf("SetHostWorkloads failed: %v", err)
	}

	state, err := p.GetDesiredState("test-host", 0)
	if err != nil {
		t.Fatalf("GetDesiredState failed: %v", err)
	}

	if state.Revision != 1 {
		t.Errorf("Expected revision 1, got %d", state.Revision)
	}

	if len(state.PodmanPlays) != 1 {
		t.Errorf("Expected 1 workload, got %d", len(state.PodmanPlays))
	}
}

// TestStatusProcessor_ProcessResult tests processing reconciliation results.
func TestStatusProcessor_ProcessResult(t *testing.T) {
	p := NewStatusProcessor()

	result := &messages.ReconciliationResult{
		Type:      messages.MessageTypeReconciliationResult,
		SessionID: "test-session",
		Host:      "test-host",
		Revision:  1,
		Workloads: []messages.WorkloadResult{
			{
				UID:        "workload-1",
				Namespace:  "default",
				Name:       "test-workload",
				Generation: 1,
				Phase:      PhaseReady,
			},
		},
	}

	err := p.ProcessResult(result)
	if err != nil {
		t.Fatalf("ProcessResult failed: %v", err)
	}

	// Verify status was stored
	status, exists := p.GetWorkloadStatus("default", "test-workload")
	if !exists {
		t.Fatal("Workload status not found")
	}

	if status.Phase != PhaseReady {
		t.Errorf("Expected phase 'Ready', got '%s'", status.Phase)
	}
}

// TestHeartbeatProcessor_IsSessionAlive tests session alive check.
func TestHeartbeatProcessor_IsSessionAlive(t *testing.T) {
	p := NewHeartbeatProcessor()
	sessionID := "test-session"

	// Process a heartbeat
	heartbeat := &messages.Heartbeat{
		Type:      messages.MessageTypeHeartbeat,
		SessionID: sessionID,
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
			Assigned: 1,
			Ready:    1,
			Failed:   0,
		},
		Host: struct {
			Name string `json:"name"`
		}{
			Name: "test-host",
		},
	}

	err := p.ProcessHeartbeat(sessionID, heartbeat)
	if err != nil {
		t.Fatalf("ProcessHeartbeat failed: %v", err)
	}

	// Check if session is alive
	if !p.IsSessionAlive(sessionID) {
		t.Error("Expected session to be alive")
	}
}

// TestServer_Initialize tests server initialization.
func TestServer_Initialize(t *testing.T) {
	server := NewServer(nil)

	if server.sessionManager == nil {
		t.Error("Expected sessionManager to be initialized")
	}

	if server.handshakeHandler == nil {
		t.Error("Expected handshakeHandler to be initialized")
	}

	if server.desiredStateProvider == nil {
		t.Error("Expected desiredStateProvider to be initialized")
	}

	if server.statusProcessor == nil {
		t.Error("Expected statusProcessor to be initialized")
	}

	if server.heartbeatProcessor == nil {
		t.Error("Expected heartbeatProcessor to be initialized")
	}
}

// TestHandshakeHandler_HandleAgentHello tests agent hello handling.
func TestHandshakeHandler_HandleAgentHello(t *testing.T) {
	hh := NewHandshakeHandler()

	// Create a minimal AgentHello (without validation to test basic flow)
	agentHello := &messages.AgentHello{
		Agent: struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		}{
			ID:      "test-agent",
			Version: "1.0.0",
		},
		Host: struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		}{
			Name:     "test-host",
			Hostname: "test-host.example.com",
		},
		ProtocolVersions: []string{ProtocolVersionV1Alpha1},
		Runtime: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    messages.RuntimePodman,
			Version: "5.6.0",
		},
		Capabilities: []string{"play-kube", "replace"},
	}

	response, err := hh.HandleAgentHello(agentHello)
	if err != nil {
		t.Fatalf("HandleAgentHello failed: %v", err)
	}

	if response == nil {
		t.Error("Expected non-nil response")
	}

	if response.ProtocolVersion != ProtocolVersionV1Alpha1 {
		t.Errorf("Expected protocol version %s, got %s", ProtocolVersionV1Alpha1, response.ProtocolVersion)
	}
}

// TestServer_Getters tests server getter methods.
func TestServer_Getters(t *testing.T) {
	server := NewServer(nil)

	sm := server.GetSessionManager()
	if sm == nil {
		t.Error("Expected non-nil SessionManager")
	}

	dsp := server.GetDesiredStateProvider()
	if dsp == nil {
		t.Error("Expected non-nil DesiredStateProvider")
	}

	sp := server.GetStatusProcessor()
	if sp == nil {
		t.Error("Expected non-nil StatusProcessor")
	}

	hbp := server.GetHeartbeatProcessor()
	if hbp == nil {
		t.Error("Expected non-nil HeartbeatProcessor")
	}
}
