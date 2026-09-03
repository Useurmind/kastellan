package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// MessageHandler processes incoming protocol messages
type MessageHandler interface {
	Handle(ctx context.Context, msg []byte) error
	MessageType() messages.MessageType
}

// HandlerRegistry manages message handlers
type HandlerRegistry struct {
	handlers map[messages.MessageType]MessageHandler
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[messages.MessageType]MessageHandler),
	}
}

// Register adds a handler for a message type
func (r *HandlerRegistry) Register(handler MessageHandler) {
	r.handlers[handler.MessageType()] = handler
}

// Get returns the handler for a message type
func (r *HandlerRegistry) Get(msgType messages.MessageType) (MessageHandler, error) {
	handler, ok := r.handlers[msgType]
	if !ok {
		return nil, fmt.Errorf("no handler for message type: %s", msgType)
	}
	return handler, nil
}

// Handle processes a raw message and dispatches to the appropriate handler
func (r *HandlerRegistry) Handle(ctx context.Context, msgType messages.MessageType, data []byte) error {
	handler, err := r.Get(msgType)
	if err != nil {
		return err
	}
	return handler.Handle(ctx, data)
}

// StreamMessageHandler handles messages from a gRPC stream
type StreamMessageHandler struct {
	registry  *HandlerRegistry
	stream    grpc.ClientStream
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	running   bool
	stopChan  chan struct{}
	errorChan chan error
}

// NewStreamMessageHandler creates a new stream message handler
func NewStreamMessageHandler(ctx context.Context, stream grpc.ClientStream, registry *HandlerRegistry) *StreamMessageHandler {
	ctx, cancel := context.WithCancel(ctx)
	return &StreamMessageHandler{
		registry:  registry,
		stream:    stream,
		ctx:       ctx,
		cancel:    cancel,
		stopChan:  make(chan struct{}),
		errorChan: make(chan error, 100),
	}
}

// Start begins the message receive loop
func (h *StreamMessageHandler) Start() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return fmt.Errorf("message handler already running")
	}
	h.running = true
	h.mu.Unlock()

	go h.receiveLoop()
	return nil
}

// Stop halts the message receive loop
func (h *StreamMessageHandler) Stop() {
	h.mu.Lock()
	if !h.running {
		h.mu.Unlock()
		return
	}
	h.running = false
	h.mu.Unlock()

	close(h.stopChan)
	h.cancel()
}

// ErrorChannel returns a channel for receiving errors
func (h *StreamMessageHandler) ErrorChannel() <-chan error {
	return h.errorChan
}

// receiveLoop continuously receives messages from the stream
func (h *StreamMessageHandler) receiveLoop() {
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-h.stopChan:
			return
		default:
		}

		// Read message from stream
		// Note: This would use actual gRPC streaming to receive messages
		// For now, we'll just sleep and continue
		select {
		case <-time.After(1 * time.Second):
			continue
		case <-h.ctx.Done():
			return
		case <-h.stopChan:
			return
		}
	}
}

// SendMessage sends a message through the stream
func (h *StreamMessageHandler) SendMessage(ctx context.Context, msg interface{}) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.running || h.stream == nil {
		return fmt.Errorf("not running or no stream")
	}

	// Encode message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send message through gRPC stream
	// Note: This would use actual gRPC streaming to send messages
	_ = data

	return nil
}

// SendHeartbeat sends a heartbeat message
func (h *StreamMessageHandler) SendHeartbeat(ctx context.Context, sessionID, hostName string) error {
	msg := messages.Heartbeat{
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
			Assigned: 0,
			Ready:    0,
			Failed:   0,
		},
		Host: struct {
			Name string `json:"name"`
		}{
			Name: hostName,
		},
	}

	return h.SendMessage(ctx, &msg)
}

// SendAgentHello sends the initial AgentHello message
func (h *StreamMessageHandler) SendAgentHello(ctx context.Context, agentID, agentVersion, hostName, hostHostname string, capabilities []string) error {
	msg := messages.AgentHello{
		Type: messages.MessageTypeAgentHello,
		Agent: struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		}{
			ID:      agentID,
			Version: agentVersion,
		},
		Host: struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		}{
			Name:     hostName,
			Hostname: hostHostname,
		},
		ProtocolVersions: []string{"v1alpha1"},
		Runtime: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "podman",
			Version: "5.6.0",
		},
		Capabilities: capabilities,
		Timestamp:    time.Now(),
	}

	return h.SendMessage(ctx, &msg)
}

// SendEnrollmentRequest sends an enrollment request
func (h *StreamMessageHandler) SendEnrollmentRequest(ctx context.Context, token, agentID, agentVersion, hostName, hostHostname string) error {
	msg := messages.EnrollmentRequest{
		Type:  messages.MessageTypeEnrollmentRequest,
		Token: token,
		Agent: struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		}{
			ID:      agentID,
			Version: agentVersion,
		},
		Host: struct {
			Name      string `json:"name"`
			Hostname  string `json:"hostname"`
			IPAddress string `json:"ipAddress,omitempty"`
		}{
			Name:     hostName,
			Hostname: hostHostname,
		},
		Timestamp: time.Now(),
	}

	return h.SendMessage(ctx, &msg)
}
