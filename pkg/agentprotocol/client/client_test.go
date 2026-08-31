package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol/client"
)

func TestReconnector(t *testing.T) {
	// Test reconnector with default config
	reconnector := client.NewReconnector(nil)

	// Test NextDelay
	delay1 := reconnector.NextDelay()
	if delay1 < time.Second {
		t.Errorf("NextDelay() returned %v, expected at least 1s", delay1)
	}

	// Test ResetDelay
	reconnector.ResetDelay()
	delay2 := reconnector.NextDelay()
	if delay2 != time.Second {
		t.Errorf("ResetDelay() NextDelay() returned %v, expected 1s", delay2)
	}

	// Test ShouldRetry
	if !reconnector.ShouldRetry() {
		t.Error("ShouldRetry() returned false, expected true")
	}

	// Test IncrementRetry
	reconnector.IncrementRetry()
	if reconnector.ShouldRetry() {
		t.Error("ShouldRetry() returned true after increment, expected false with maxRetries=0")
	}
}

func TestHeartbeatManager(t *testing.T) {
	// Test heartbeat manager
	heartbeatManager := client.NewHeartbeatManager(nil)

	// Test Start
	ctx := context.Background()
	if err := heartbeatManager.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Test IsConnected (should be false initially)
	if heartbeatManager.IsConnected() {
		t.Error("IsConnected() returned true, expected false")
	}

	// Test RecordAck
	heartbeatManager.RecordAck()
	if !heartbeatManager.IsConnected() {
		t.Error("IsConnected() returned false after RecordAck, expected true")
	}

	// Test Stop
	heartbeatManager.Stop()
}

func TestStatusReporter(t *testing.T) {
	// Test status reporter
	statusReporter := client.NewStatusReporter(nil)

	// Test UpdateWorkloadState
	state := &client.WorkloadState{
		UID:         "test-uid",
		Namespace:   "default",
		Name:        "test-workload",
		Generation:  1,
		Phase:       "Running",
		LastUpdate:  time.Now(),
	}
	statusReporter.UpdateWorkloadState(state)

	// Test GetWorkloadState
	retrievedState := statusReporter.GetWorkloadState("test-uid")
	if retrievedState == nil {
		t.Error("GetWorkloadState() returned nil")
	} else if retrievedState.UID != state.UID {
		t.Errorf("GetWorkloadState() returned wrong UID: got %v, want %v", retrievedState.UID, state.UID)
	}

	// Test GetWorkloadStates
	states := statusReporter.GetWorkloadStates()
	if len(states) != 1 {
		t.Errorf("GetWorkloadStates() returned %v states, expected 1", len(states))
	}

	// Test DeleteWorkloadState
	statusReporter.DeleteWorkloadState("test-uid")
	states = statusReporter.GetWorkloadStates()
	if len(states) != 0 {
		t.Errorf("GetWorkloadStates() returned %v states after delete, expected 0", len(states))
	}
}

func TestCreateHeartbeatMessage(t *testing.T) {
	// Test CreateHeartbeatMessage
	msg := client.CreateHeartbeatMessage("session-123", "lb01")

	if msg.Type != "Heartbeat" {
		t.Errorf("MessageType mismatch: got %v, want Heartbeat", msg.Type)
	}
	if msg.SessionID != "session-123" {
		t.Errorf("SessionID mismatch: got %v, want session-123", msg.SessionID)
	}
	if msg.Host.Name != "lb01" {
		t.Errorf("Host.Name mismatch: got %v, want lb01", msg.Host.Name)
	}
	if !msg.Runtime.Available {
		t.Error("Runtime.Available should be true")
	}
}

func TestCreateReconciliationResult(t *testing.T) {
	// Test CreateReconciliationResult
	workloads := []client.WorkloadResult{
		{
			UID:        "test-uid",
			Namespace:  "default",
			Name:       "test-workload",
			Generation: 1,
			Phase:      "Running",
		},
	}

	result := client.CreateReconciliationResult("session-123", "lb01", 42, workloads)

	if result.Type != "ReconciliationResult" {
		t.Errorf("MessageType mismatch: got %v, want ReconciliationResult", result.Type)
	}
	if result.SessionID != "session-123" {
		t.Errorf("SessionID mismatch: got %v, want session-123", result.SessionID)
	}
	if result.Host != "lb01" {
		t.Errorf("Host mismatch: got %v, want lb01", result.Host)
	}
	if result.Revision != 42 {
		t.Errorf("Revision mismatch: got %v, want 42", result.Revision)
	}
	if len(result.Workloads) != 1 {
		t.Errorf("Workloads length mismatch: got %v, want 1", len(result.Workloads))
	}
}

func TestBackoffDelay(t *testing.T) {
	// Test BackoffDelay
	initial := time.Second
	max := time.Minute

	// Test first attempt
	delay1 := client.BackoffDelay(0, initial, max)
	if delay1 < initial || delay1 > initial*2 {
		t.Errorf("BackoffDelay(0) returned %v, expected between %v and %v", delay1, initial, initial*2)
	}

	// Test second attempt
	delay2 := client.BackoffDelay(1, initial, max)
	if delay2 < initial*2 || delay2 > initial*3 {
		t.Errorf("BackoffDelay(1) returned %v, expected between %v and %v", delay2, initial*2, initial*3)
	}

	// Test max cap
	delay3 := client.BackoffDelay(10, initial, max)
	if delay3 > max {
		t.Errorf("BackoffDelay(10) returned %v, expected at most %v", delay3, max)
	}
}
