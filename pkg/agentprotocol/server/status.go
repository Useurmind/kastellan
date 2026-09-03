package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// StatusProcessor processes workload status updates from agents.
type StatusProcessor struct {
	mu               sync.RWMutex
	workloadStatuses map[string]*WorkloadStatus
}

// WorkloadStatus tracks the status of a workload on a specific host.
type WorkloadStatus struct {
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

// NewStatusProcessor creates a new status processor.
func NewStatusProcessor() *StatusProcessor {
	return &StatusProcessor{
		workloadStatuses: make(map[string]*WorkloadStatus),
	}
}

// ProcessResult processes a reconciliation result from an agent.
func (p *StatusProcessor) ProcessResult(result *messages.ReconciliationResult) error {
	for _, workloadResult := range result.Workloads {
		status := &WorkloadStatus{
			UID:            workloadResult.UID,
			Namespace:      workloadResult.Namespace,
			Name:           workloadResult.Name,
			Generation:     workloadResult.Generation,
			Phase:          workloadResult.Phase,
			Reason:         workloadResult.Reason,
			Message:        workloadResult.Message,
			ManifestDigest: workloadResult.ManifestDigest,
		}

		status.Runtime.PodID = workloadResult.Runtime.PodID
		status.Runtime.Containers = workloadResult.Runtime.Containers

		p.UpdateWorkloadStatus(status)
	}

	return nil
}

// UpdateWorkloadStatus updates the status of a workload.
func (p *StatusProcessor) UpdateWorkloadStatus(status *WorkloadStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status.LastUpdate = time.Now()
	key := fmt.Sprintf("%s/%s", status.Namespace, status.Name)
	p.workloadStatuses[key] = status
}

// GetWorkloadStatus returns the status of a workload.
func (p *StatusProcessor) GetWorkloadStatus(namespace, name string) (*WorkloadStatus, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	status, exists := p.workloadStatuses[key]
	return status, exists
}

// GetWorkloadStatuses returns all workload statuses.
func (p *StatusProcessor) GetWorkloadStatuses() map[string]*WorkloadStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	statuses := make(map[string]*WorkloadStatus)
	for k, v := range p.workloadStatuses {
		statuses[k] = v
	}
	return statuses
}

// ClearWorkloadStatus clears the status of a workload.
func (p *StatusProcessor) ClearWorkloadStatus(namespace, name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	delete(p.workloadStatuses, key)
}

// GetPhaseCounts returns counts of workloads by phase.
func (p *StatusProcessor) GetPhaseCounts() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	counts := make(map[string]int)
	for _, status := range p.workloadStatuses {
		counts[status.Phase]++
	}
	return counts
}

// ProcessWorkloadStatus processes a single workload status update.
func (p *StatusProcessor) ProcessWorkloadStatus(status *messages.WorkloadStatus) error {
	workloadStatus := &WorkloadStatus{
		UID:        status.UID,
		Namespace:  status.Namespace,
		Name:       status.Name,
		Generation: status.Generation,
		Phase:      status.Phase,
	}

	workloadStatus.Runtime.PodID = status.Runtime.PodID
	workloadStatus.Runtime.Containers = status.Runtime.Containers

	p.UpdateWorkloadStatus(workloadStatus)

	return nil
}
