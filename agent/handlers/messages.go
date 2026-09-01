package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// OperatorHelloHandler handles OperatorHello messages
type OperatorHelloHandler struct {
	onHello func(ctx context.Context, msg *messages.OperatorHello) error
}

// NewOperatorHelloHandler creates a new OperatorHello handler
func NewOperatorHelloHandler() *OperatorHelloHandler {
	return &OperatorHelloHandler{
		onHello: func(ctx context.Context, msg *messages.OperatorHello) error {
			return nil
		},
	}
}

// WithOnHello sets the callback for handling OperatorHello
func (h *OperatorHelloHandler) WithOnHello(fn func(ctx context.Context, msg *messages.OperatorHello) error) *OperatorHelloHandler {
	h.onHello = fn
	return h
}

// Handle processes an OperatorHello message
func (h *OperatorHelloHandler) Handle(ctx context.Context, data []byte) error {
	var msg messages.OperatorHello
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal OperatorHello: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid OperatorHello: %w", err)
	}

	if h.onHello != nil {
		return h.onHello(ctx, &msg)
	}

	return nil
}

// MessageType returns the message type handled by this handler
func (h *OperatorHelloHandler) MessageType() messages.MessageType {
	return messages.MessageTypeOperatorHello
}

// DesiredStateHandler handles DesiredState messages
type DesiredStateHandler struct {
	onDesiredState func(ctx context.Context, msg *messages.DesiredState) error
}

// NewDesiredStateHandler creates a new DesiredState handler
func NewDesiredStateHandler() *DesiredStateHandler {
	return &DesiredStateHandler{
		onDesiredState: func(ctx context.Context, msg *messages.DesiredState) error {
			return nil
		},
	}
}

// WithOnDesiredState sets the callback for handling DesiredState
func (h *DesiredStateHandler) WithOnDesiredState(fn func(ctx context.Context, msg *messages.DesiredState) error) *DesiredStateHandler {
	h.onDesiredState = fn
	return h
}

// Handle processes a DesiredState message
func (h *DesiredStateHandler) Handle(ctx context.Context, data []byte) error {
	var msg messages.DesiredState
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal DesiredState: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid DesiredState: %w", err)
	}

	if h.onDesiredState != nil {
		return h.onDesiredState(ctx, &msg)
	}

	return nil
}

// MessageType returns the message type handled by this handler
func (h *DesiredStateHandler) MessageType() messages.MessageType {
	return messages.MessageTypeDesiredState
}

// ReconcileRequestHandler handles ReconcileRequest messages
type ReconcileRequestHandler struct {
	onReconcile func(ctx context.Context, msg *messages.ReconcileRequest) error
}

// NewReconcileRequestHandler creates a new ReconcileRequest handler
func NewReconcileRequestHandler() *ReconcileRequestHandler {
	return &ReconcileRequestHandler{
		onReconcile: func(ctx context.Context, msg *messages.ReconcileRequest) error {
			return nil
		},
	}
}

// WithOnReconcile sets the callback for handling ReconcileRequest
func (h *ReconcileRequestHandler) WithOnReconcile(fn func(ctx context.Context, msg *messages.ReconcileRequest) error) *ReconcileRequestHandler {
	h.onReconcile = fn
	return h
}

// Handle processes a ReconcileRequest message
func (h *ReconcileRequestHandler) Handle(ctx context.Context, data []byte) error {
	var msg messages.ReconcileRequest
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal ReconcileRequest: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid ReconcileRequest: %w", err)
	}

	if h.onReconcile != nil {
		return h.onReconcile(ctx, &msg)
	}

	return nil
}

// MessageType returns the message type handled by this handler
func (h *ReconcileRequestHandler) MessageType() messages.MessageType {
	return messages.MessageTypeReconcileRequest
}

// ConnectionCloseHandler handles ConnectionClose messages
type ConnectionCloseHandler struct {
	onClose func(ctx context.Context, msg *messages.ConnectionClose) error
}

// NewConnectionCloseHandler creates a new ConnectionClose handler
func NewConnectionCloseHandler() *ConnectionCloseHandler {
	return &ConnectionCloseHandler{
		onClose: func(ctx context.Context, msg *messages.ConnectionClose) error {
			return nil
		},
	}
}

// WithOnClose sets the callback for handling ConnectionClose
func (h *ConnectionCloseHandler) WithOnClose(fn func(ctx context.Context, msg *messages.ConnectionClose) error) *ConnectionCloseHandler {
	h.onClose = fn
	return h
}

// Handle processes a ConnectionClose message
func (h *ConnectionCloseHandler) Handle(ctx context.Context, data []byte) error {
	var msg messages.ConnectionClose
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal ConnectionClose: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid ConnectionClose: %w", err)
	}

	if h.onClose != nil {
		return h.onClose(ctx, &msg)
	}

	return nil
}

// MessageType returns the message type handled by this handler
func (h *ConnectionCloseHandler) MessageType() messages.MessageType {
	return messages.MessageTypeConnectionClose
}

// EnrollmentResponseHandler handles EnrollmentResponse messages
type EnrollmentResponseHandler struct {
	onEnrollment func(ctx context.Context, msg *messages.EnrollmentResponse) error
}

// NewEnrollmentResponseHandler creates a new EnrollmentResponse handler
func NewEnrollmentResponseHandler() *EnrollmentResponseHandler {
	return &EnrollmentResponseHandler{
		onEnrollment: func(ctx context.Context, msg *messages.EnrollmentResponse) error {
			return nil
		},
	}
}

// WithOnEnrollment sets the callback for handling EnrollmentResponse
func (h *EnrollmentResponseHandler) WithOnEnrollment(fn func(ctx context.Context, msg *messages.EnrollmentResponse) error) *EnrollmentResponseHandler {
	h.onEnrollment = fn
	return h
}

// Handle processes an EnrollmentResponse message
func (h *EnrollmentResponseHandler) Handle(ctx context.Context, data []byte) error {
	var msg messages.EnrollmentResponse
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal EnrollmentResponse: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid EnrollmentResponse: %w", err)
	}

	if h.onEnrollment != nil {
		return h.onEnrollment(ctx, &msg)
	}

	return nil
}

// MessageType returns the message type handled by this handler
func (h *EnrollmentResponseHandler) MessageType() messages.MessageType {
	return messages.MessageTypeEnrollmentResponse
}
