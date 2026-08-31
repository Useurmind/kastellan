package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/useurmind/kastellan/pkg/agentprotocol/client"
	"github.com/useurmind/kastellan/pkg/agentprotocol/messages"
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
	// ResetDelay sets currentDelay to initialDelay, but NextDelay adds jitter (0-50%)
	// So we expect delay2 to be between 1s and 1.5s
	maxExpected := time.Duration(float64(time.Second) * 1.5)
	if delay2 < time.Second || delay2 > maxExpected {
		t.Errorf("ResetDelay() NextDelay() returned %v, expected between 1s and 1.5s", delay2)
	}

	// Test ShouldRetry
	if !reconnector.ShouldRetry() {
		t.Error("ShouldRetry() returned false, expected true")
	}

	// Test IncrementRetry
	reconnector.IncrementRetry()
	// maxRetries=0 means infinite retries, so ShouldRetry should return true
	if !reconnector.ShouldRetry() {
		t.Error("ShouldRetry() returned false after increment, expected true with maxRetries=0")
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
	workloads := []messages.WorkloadResult{
		{
			UID:        "test-uid",
			Namespace:  "default",
			Name:       "test-workload",
			Generation: 1,
			Phase:      "Running",
		},
	}

	result := messages.ReconciliationResult{
		Type:      messages.MessageTypeReconciliationResult,
		SessionID: "session-123",
		Host:      "lb01",
		Revision:  42,
		Timestamp: time.Now(),
		Workloads: workloads,
	}

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

	// Test first attempt (attempt 0: initial * 1, with 0-50% jitter)
	delay1 := client.BackoffDelay(0, initial, max)
	// With jitter (0-50%), delay1 should be between 1s and 1.5s
	maxExpected1 := time.Duration(float64(initial) * 1.5)
	if delay1 < initial || delay1 > maxExpected1 {
		t.Errorf("BackoffDelay(0) returned %v, expected between %v and %v", delay1, initial, maxExpected1)
	}

	// Test second attempt (attempt 1: initial * 2, with 0-50% jitter)
	delay2 := client.BackoffDelay(1, initial, max)
	// With jitter (0-50%), delay2 should be between 2s and 3s
	maxExpected2 := time.Duration(float64(initial*2) * 1.5)
	if delay2 < initial*2 || delay2 > maxExpected2 {
		t.Errorf("BackoffDelay(1) returned %v, expected between %v and %v", delay2, initial*2, maxExpected2)
	}

	// Test max cap (attempt 10: initial * 1024, capped at max, with 0-50% jitter)
	delay3 := client.BackoffDelay(10, initial, max)
	// With jitter (0-50%), delay3 should be between max and 1.5*max
	maxExpected3 := time.Duration(float64(max) * 1.5)
	if delay3 > maxExpected3 {
		t.Errorf("BackoffDelay(10) returned %v, expected at most %v", delay3, maxExpected3)
	}
}
