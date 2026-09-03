package server

import (
	"fmt"

	"google.golang.org/grpc"

	agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)

// GrpcServer implements the AgentProtocol gRPC service.
type GrpcServer struct {
	agentv1alpha1.UnimplementedAgentProtocolServer
}

// NewGrpcServer creates a new agent protocol gRPC server.
func NewGrpcServer() *GrpcServer {
	return &GrpcServer{}
}

// Connect handles the bidirectional streaming connection from agents.
func (s *GrpcServer) Connect(stream grpc.BidiStreamingServer[agentv1alpha1.ProtocolMessage, agentv1alpha1.ProtocolMessage]) error {
	for {
		msg := agentv1alpha1.ProtocolMessage{}
		if err := stream.RecvMsg(&msg); err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}

		response, err := s.handleMessage(&msg)
		if err != nil {
			return fmt.Errorf("failed to handle message: %w", err)
		}

		if response != nil {
			if err := stream.SendMsg(response); err != nil {
				return fmt.Errorf("failed to send response: %w", err)
			}
		}
	}
}

// handleMessage processes incoming messages and returns responses.
func (s *GrpcServer) handleMessage(msg *agentv1alpha1.ProtocolMessage) (*agentv1alpha1.ProtocolMessage, error) {
	switch payload := msg.Payload.(type) {
	case *agentv1alpha1.ProtocolMessage_AgentHello:
		return s.handleAgentHello(payload.AgentHello)
	default:
		return nil, nil
	}
}

// handleAgentHello processes AgentHello messages.
func (s *GrpcServer) handleAgentHello(hello *agentv1alpha1.AgentHello) (*agentv1alpha1.ProtocolMessage, error) {
	response := &agentv1alpha1.OperatorHello{
		SessionId:                  "session-" + hello.AgentId,
		SelectedProtocol:           &agentv1alpha1.ProtocolVersion{Major: 1, Minor: 0},
		HeartbeatIntervalSeconds:   30,
		StateReportIntervalSeconds: 60,
		ServerTimeUnix:             0,
	}

	return &agentv1alpha1.ProtocolMessage{
		Payload: &agentv1alpha1.ProtocolMessage_OperatorHello{
			OperatorHello: response,
		},
	}, nil
}
