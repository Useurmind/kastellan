package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol"
	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// DesiredStateProvider provides desired state for hosts.
type DesiredStateProvider struct {
	mu        sync.RWMutex
	revisions map[string]int64
	hosts     map[string]map[string]*messages.PodmanPlay
}

// NewDesiredStateProvider creates a new desired state provider.
func NewDesiredStateProvider() *DesiredStateProvider {
	return &DesiredStateProvider{
		revisions: make(map[string]int64),
		hosts:     make(map[string]map[string]*messages.PodmanPlay),
	}
}

// SetHostWorkloads sets the workloads for a host.
func (p *DesiredStateProvider) SetHostWorkloads(host string, workloads map[string]*messages.PodmanPlay) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.hosts[host] = workloads
	p.revisions[host]++

	return nil
}

// GetDesiredState returns the desired state for a host at a specific revision.
func (p *DesiredStateProvider) GetDesiredState(host string, revision int64) (*messages.DesiredState, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	currentRevision, exists := p.revisions[host]
	if !exists {
		return nil, agentprotocol.NewProtocolError(
			"host_not_found",
			fmt.Sprintf("no workloads found for host: %s", host),
		)
	}

	// If requesting a specific revision that's older than current,
	// return the current state (idempotent delivery)
	if revision > 0 && revision < currentRevision {
		return nil, agentprotocol.NewProtocolError(
			"stale_revision",
			fmt.Sprintf("revision %d is stale, current is %d", revision, currentRevision),
		)
	}

	hostWorkloads, exists := p.hosts[host]
	if !exists || len(hostWorkloads) == 0 {
		return &messages.DesiredState{
			Type:        messages.MessageTypeDesiredState,
			Host:        host,
			Revision:    currentRevision,
			Timestamp:   time.Now(),
			PodmanPlays: []messages.PodmanPlay{},
		}, nil
	}

	// Convert map to slice
	var podmanPlays []messages.PodmanPlay
	for _, play := range hostWorkloads {
		podmanPlays = append(podmanPlays, *play)
	}

	return &messages.DesiredState{
		Type:        messages.MessageTypeDesiredState,
		Host:        host,
		Revision:    currentRevision,
		Timestamp:   time.Now(),
		PodmanPlays: podmanPlays,
	}, nil
}

// GetNextRevision returns the next revision number for a host.
func (p *DesiredStateProvider) GetNextRevision(host string) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	revision, exists := p.revisions[host]
	if !exists {
		return 1, nil
	}

	return revision + 1, nil
}

// MarkRevisionApplied marks a revision as applied.
func (p *DesiredStateProvider) MarkRevisionApplied(host string, revision int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if current, exists := p.revisions[host]; exists && revision >= current {
		p.revisions[host] = revision
	}

	return nil
}

// GetHostRevision returns the current revision for a host.
func (p *DesiredStateProvider) GetHostRevision(host string) (int64, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	revision, exists := p.revisions[host]
	return revision, exists
}

// RemoveHost removes all workloads for a host.
func (p *DesiredStateProvider) RemoveHost(host string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.hosts, host)
}

// HasHost returns whether a host has workloads.
func (p *DesiredStateProvider) HasHost(host string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.hosts[host]
	return exists
}
