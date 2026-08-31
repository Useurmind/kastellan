// Package client provides status reporting for the Kastellan Agent Protocol.
package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// StatusReporter reports workload status to the operator.
type StatusReporter struct {
	// Configuration
	reportInterval time.Duration

	// State
	mu            sync.Mutex
	running       bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
	workloadState map[string]*WorkloadState
}

// WorkloadState tracks the state of a workload.
type WorkloadState struct {
	UID            string
	Namespace      string
	Name           string
	Generation     int64
	Phase          string
	Reason         string
	Message        string
	ManifestDigest string
	Runtime        struct {
		PodID      string
		Containers []messages.ContainerInfo
	}
	LastUpdate time.Time
}

// StatusReporterConfig configures the status reporter.
type StatusReporterConfig struct {
	ReportInterval time.Duration
}

// DefaultStatusReporterConfig returns the default status reporter configuration.
func DefaultStatusReporterConfig() *StatusReporterConfig {
	return &StatusReporterConfig{
		ReportInterval: 60 * time.Second,
	}
}

// NewStatusReporter creates a new status reporter.
func NewStatusReporter(config *StatusReporterConfig) *StatusReporter {
	if config == nil {
		config = DefaultStatusReporterConfig()
	}

	return &StatusReporter{
		reportInterval: config.ReportInterval,
		workloadState:  make(map[string]*WorkloadState),
		stopChan:       make(chan struct{}),
	}
}

// Start starts the status reporter.
func (r *StatusReporter) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("status reporter already running")
	}
	r.running = true
	r.mu.Unlock()

	r.wg.Add(1)
	go r.run(ctx)

	return nil
}

// Stop stops the status reporter.
func (r *StatusReporter) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	r.mu.Unlock()

	close(r.stopChan)
	r.wg.Wait()
}

// run runs the status reporting loop.
func (r *StatusReporter) run(ctx context.Context) {
	defer r.wg.Done()

	ticker := time.NewTicker(r.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopChan:
			return
		case <-ticker.C:
			if err := r.reportStatus(ctx); err != nil {
				fmt.Printf("Failed to report status: %v\n", err)
			}
		}
	}
}

// reportStatus reports the current status to the operator.
func (r *StatusReporter) reportStatus(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Collect all workload states
	var workloads []messages.WorkloadResult
	for _, state := range r.workloadState {
		result := messages.WorkloadResult{
			UID:            state.UID,
			Namespace:      state.Namespace,
			Name:           state.Name,
			Generation:     state.Generation,
			Phase:          state.Phase,
			Reason:         state.Reason,
			Message:        state.Message,
			ManifestDigest: state.ManifestDigest,
			Runtime: struct {
				PodID      string                   `json:"podId,omitempty"`
				Containers []messages.ContainerInfo `json:"containers,omitempty"`
			}{
				PodID:      state.Runtime.PodID,
				Containers: state.Runtime.Containers,
			},
		}
		workloads = append(workloads, result)
	}

	if len(workloads) == 0 {
		return nil
	}

	// Create reconciliation result
	result := &messages.ReconciliationResult{
		Type:      messages.MessageTypeReconciliationResult,
		SessionID: "", // Would be set from connection
		Host:      "", // Would be set from connection
		Revision:  0,  // Would be set from connection
		Timestamp: time.Now(),
		Workloads: workloads,
	}

	// Send result (would use actual gRPC stream)
	// For now, just return nil
	return nil
}

// UpdateWorkloadState updates the state of a workload.
func (r *StatusReporter) UpdateWorkloadState(state *WorkloadState) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.workloadState[state.UID] = state
}

// GetWorkloadState returns the state of a workload.
func (r *StatusReporter) GetWorkloadState(uid string) *WorkloadState {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.workloadState[uid]
}

// DeleteWorkloadState deletes the state of a workload.
func (r *StatusReporter) DeleteWorkloadState(uid string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.workloadState, uid)
}

// GetWorkloadStates returns all workload states.
func (r *StatusReporter) GetWorkloadStates() map[string]*WorkloadState {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy of the map
	states := make(map[string]*WorkloadState, len(r.workloadState))
	for k, v := range r.workloadState {
		states[k] = v
	}

	return states
}

// CreateWorkloadResult creates a workload result from a workload state.
func CreateWorkloadResult(state *WorkloadState) messages.WorkloadResult {
	return messages.WorkloadResult{
		UID:            state.UID,
		Namespace:      state.Namespace,
		Name:           state.Name,
		Generation:     state.Generation,
		Phase:          state.Phase,
		Reason:         state.Reason,
		Message:        state.Message,
		ManifestDigest: state.ManifestDigest,
		Runtime: struct {
			PodID      string                   `json:"podId,omitempty"`
			Containers []messages.ContainerInfo `json:"containers,omitempty"`
		}{
			PodID:      state.Runtime.PodID,
			Containers: state.Runtime.Containers,
		},
	}
}

// CreateReconciliationResult creates a reconciliation result.
func CreateReconciliationResult(sessionID, host string, revision int64, workloads []messages.WorkloadResult) *messages.ReconciliationResult {
	return &messages.ReconciliationResult{
		Type:      messages.MessageTypeReconciliationResult,
		SessionID: sessionID,
		Host:      host,
		Revision:  revision,
		Timestamp: time.Now(),
		Workloads: workloads,
	}
}

// WorkloadPhase constants.
const (
	PhasePending  = "Pending"
	PhaseRunning  = "Running"
	PhaseReady    = "Ready"
	PhaseFailed   = "Failed"
	PhaseUpdating = "Updating"
	PhaseUnknown  = "Unknown"
)

// WorkloadStatus represents the status of a workload.
type WorkloadStatus struct {
	UID        string
	Namespace  string
	Name       string
	Generation int64
	Phase      string
	Reason     string
	Message    string
	Runtime    struct {
		PodID      string
		Containers []messages.ContainerInfo
	}
}

// NewWorkloadStatus creates a new workload status.
func NewWorkloadStatus(uid, namespace, name string, generation int64, phase, reason, message string) *WorkloadStatus {
	return &WorkloadStatus{
		UID:        uid,
		Namespace:  namespace,
		Name:       name,
		Generation: generation,
		Phase:      phase,
		Reason:     reason,
		Message:    message,
	}
}

// SetRuntime sets the runtime information for the workload.
func (w *WorkloadStatus) SetRuntime(podID string, containers []messages.ContainerInfo) {
	w.Runtime.PodID = podID
	w.Runtime.Containers = containers
}

// ToWorkloadResult converts the status to a result.
func (w *WorkloadStatus) ToWorkloadResult() messages.WorkloadResult {
	return messages.WorkloadResult{
		UID:        w.UID,
		Namespace:  w.Namespace,
		Name:       w.Name,
		Generation: w.Generation,
		Phase:      w.Phase,
		Reason:     w.Reason,
		Message:    w.Message,
		Runtime: struct {
			PodID      string                   `json:"podId,omitempty"`
			Containers []messages.ContainerInfo `json:"containers,omitempty"`
		}{
			PodID:      w.Runtime.PodID,
			Containers: w.Runtime.Containers,
		},
	}
}
