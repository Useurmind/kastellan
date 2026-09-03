package server

import (
	"sync"
	"time"
)

// ReconcileRequestStore stores pending reconcile requests.
type ReconcileRequestStore struct {
	requests map[string][]ReconcileRequest
	mu       sync.RWMutex
}

// ReconcileRequest represents a request to reconcile a workload.
type ReconcileRequest struct {
	Host        string
	Revision    int64
	WorkloadUID string
	RequestedAt time.Time
}

// NewReconcileRequestStore creates a new reconcile request store.
func NewReconcileRequestStore() *ReconcileRequestStore {
	return &ReconcileRequestStore{
		requests: make(map[string][]ReconcileRequest),
	}
}

// AddRequest adds a reconcile request for a host.
func (s *ReconcileRequestStore) AddRequest(host string, request ReconcileRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests[host] = append(s.requests[host], request)
}

// GetRequests gets all reconcile requests for a host.
func (s *ReconcileRequestStore) GetRequests(host string) []ReconcileRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests := s.requests[host]
	result := make([]ReconcileRequest, len(requests))
	copy(result, requests)
	return result
}

// ClearRequests clears all requests for a host.
func (s *ReconcileRequestStore) ClearRequests(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.requests, host)
}

// Server is the main gRPC server for the agent protocol.
type Server struct {
	sessionManager       *SessionManager
	handshakeHandler     *HandshakeHandler
	desiredStateProvider *DesiredStateProvider
	statusProcessor      *StatusProcessor
	heartbeatProcessor   *HeartbeatProcessor
	reconcileStore       *ReconcileRequestStore

	mu sync.RWMutex
}

// ServerConfig configures the server.
type ServerConfig struct {
	SessionManager       *SessionManager
	HandshakeHandler     *HandshakeHandler
	DesiredStateProvider *DesiredStateProvider
	StatusProcessor      *StatusProcessor
	HeartbeatProcessor   *HeartbeatProcessor
	ReconcileStore       *ReconcileRequestStore
}

// NewServer creates a new agent protocol server.
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = &ServerConfig{}
	}

	if config.SessionManager == nil {
		config.SessionManager = NewSessionManager()
	}

	if config.HandshakeHandler == nil {
		config.HandshakeHandler = NewHandshakeHandler()
	}

	if config.DesiredStateProvider == nil {
		config.DesiredStateProvider = NewDesiredStateProvider()
	}

	if config.StatusProcessor == nil {
		config.StatusProcessor = NewStatusProcessor()
	}

	if config.HeartbeatProcessor == nil {
		config.HeartbeatProcessor = NewHeartbeatProcessor()
	}

	if config.ReconcileStore == nil {
		config.ReconcileStore = NewReconcileRequestStore()
	}

	return &Server{
		sessionManager:       config.SessionManager,
		handshakeHandler:     config.HandshakeHandler,
		desiredStateProvider: config.DesiredStateProvider,
		statusProcessor:      config.StatusProcessor,
		heartbeatProcessor:   config.HeartbeatProcessor,
		reconcileStore:       config.ReconcileStore,
	}
}

// GetSessionManager returns the session manager.
func (s *Server) GetSessionManager() *SessionManager {
	return s.sessionManager
}

// GetDesiredStateProvider returns the desired state provider.
func (s *Server) GetDesiredStateProvider() *DesiredStateProvider {
	return s.desiredStateProvider
}

// GetStatusProcessor returns the status processor.
func (s *Server) GetStatusProcessor() *StatusProcessor {
	return s.statusProcessor
}

// GetHeartbeatProcessor returns the heartbeat processor.
func (s *Server) GetHeartbeatProcessor() *HeartbeatProcessor {
	return s.heartbeatProcessor
}
