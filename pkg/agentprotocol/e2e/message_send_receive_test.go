//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
	"github.com/kastellan/kastellan/pkg/agentprotocol/server"
)

func TestMessageSendReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()
	agentv1alpha1.RegisterAgentProtocolServer(s, server.NewGrpcServer())

	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	conn, err := grpc.DialContext(
		ctx,
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := agentv1alpha1.NewAgentProtocolClient(conn)

	t.Run("agent sends hello and receives operator hello", func(t *testing.T) {
		stream, err := client.Connect(ctx)
		require.NoError(t, err)

		agentHello := &agentv1alpha1.AgentHello{
			AgentId:              "test-agent",
			HostName:             "test-host",
			AgentVersion:         "1.0.0",
			SupportedProtocols:   []*agentv1alpha1.ProtocolVersion{{Major: 1, Minor: 0}},
			Runtime:              &agentv1alpha1.RuntimeInformation{Name: "podman", Version: "5.0.0"},
			LastSessionId:        "",
			LastReceivedRevision: 0,
			LastAppliedRevision:  0,
		}

		err = stream.Send(&agentv1alpha1.ProtocolMessage{
			Payload: &agentv1alpha1.ProtocolMessage_AgentHello{
				AgentHello: agentHello,
			},
		})
		require.NoError(t, err)

		msg, err := stream.Recv()
		require.NoError(t, err)

		assert.IsType(t, &agentv1alpha1.ProtocolMessage_OperatorHello{}, msg.Payload)
		operatorHello := msg.GetOperatorHello()
		assert.NotEmpty(t, operatorHello.SessionId)
		assert.Equal(t, uint32(1), operatorHello.SelectedProtocol.Major)
		assert.Equal(t, uint32(0), operatorHello.SelectedProtocol.Minor)
		assert.Equal(t, uint32(30), operatorHello.HeartbeatIntervalSeconds)
		assert.Equal(t, uint32(60), operatorHello.StateReportIntervalSeconds)
	})
}

func TestMessageReceiveLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()
	agentv1alpha1.RegisterAgentProtocolServer(s, server.NewGrpcServer())

	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	conn, err := grpc.DialContext(
		ctx,
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := agentv1alpha1.NewAgentProtocolClient(conn)

	t.Run("client can receive multiple operator messages", func(t *testing.T) {
		stream, err := client.Connect(ctx)
		require.NoError(t, err)

		agentHello := &agentv1alpha1.AgentHello{
			AgentId:              "test-agent-2",
			HostName:             "test-host-2",
			AgentVersion:         "1.0.0",
			SupportedProtocols:   []*agentv1alpha1.ProtocolVersion{{Major: 1, Minor: 0}},
			Runtime:              &agentv1alpha1.RuntimeInformation{Name: "podman", Version: "5.0.0"},
			LastSessionId:        "",
			LastReceivedRevision: 0,
			LastAppliedRevision:  0,
		}

		for i := 0; i < 3; i++ {
			err = stream.Send(&agentv1alpha1.ProtocolMessage{
				Payload: &agentv1alpha1.ProtocolMessage_AgentHello{
					AgentHello: agentHello,
				},
			})
			require.NoError(t, err)

			msg, err := stream.Recv()
			require.NoError(t, err)

			assert.IsType(t, &agentv1alpha1.ProtocolMessage_OperatorHello{}, msg.Payload)
		}
	})
}

func TestServerReceiveAgentHello(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()
	agentv1alpha1.RegisterAgentProtocolServer(s, server.NewGrpcServer())

	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	conn, err := grpc.DialContext(
		ctx,
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := agentv1alpha1.NewAgentProtocolClient(conn)

	t.Run("server receives agent hello message", func(t *testing.T) {
		stream, err := client.Connect(ctx)
		require.NoError(t, err)

		agentHello := &agentv1alpha1.AgentHello{
			AgentId:              "hello-agent",
			HostName:             "hello-host",
			AgentVersion:         "2.0.0",
			SupportedProtocols:   []*agentv1alpha1.ProtocolVersion{{Major: 1, Minor: 0}},
			Runtime:              &agentv1alpha1.RuntimeInformation{Name: "podman", Version: "5.0.0"},
			LastSessionId:        "",
			LastReceivedRevision: 0,
			LastAppliedRevision:  0,
		}

		err = stream.Send(&agentv1alpha1.ProtocolMessage{
			Payload: &agentv1alpha1.ProtocolMessage_AgentHello{
				AgentHello: agentHello,
			},
		})
		require.NoError(t, err)

		msg, err := stream.Recv()
		require.NoError(t, err)

		assert.NotNil(t, msg)
		assert.NotNil(t, msg.GetOperatorHello())
		assert.Contains(t, msg.GetOperatorHello().GetSessionId(), "hello-agent")
	})
}

func TestConcurrentMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()
	agentv1alpha1.RegisterAgentProtocolServer(s, server.NewGrpcServer())

	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	conn, err := grpc.DialContext(
		ctx,
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)
	defer conn.Close()

	t.Run("multiple agents can connect concurrently", func(t *testing.T) {
		numAgents := 5
		errors := make(chan error, numAgents)

		for i := 0; i < numAgents; i++ {
			go func(idx int) {
				client := agentv1alpha1.NewAgentProtocolClient(conn)
				stream, err := client.Connect(ctx)
				if err != nil {
					errors <- fmt.Errorf("failed to create stream for agent %d: %w", idx, err)
					return
				}

				agentHello := &agentv1alpha1.AgentHello{
					AgentId:              fmt.Sprintf("concurrent-agent-%d", idx),
					HostName:             fmt.Sprintf("concurrent-host-%d", idx),
					AgentVersion:         "1.0.0",
					SupportedProtocols:   []*agentv1alpha1.ProtocolVersion{{Major: 1, Minor: 0}},
					Runtime:              &agentv1alpha1.RuntimeInformation{Name: "podman", Version: "5.0.0"},
					LastSessionId:        "",
					LastReceivedRevision: 0,
					LastAppliedRevision:  0,
				}

				err = stream.Send(&agentv1alpha1.ProtocolMessage{
					Payload: &agentv1alpha1.ProtocolMessage_AgentHello{
						AgentHello: agentHello,
					},
				})
				if err != nil {
					errors <- fmt.Errorf("failed to send hello for agent %d: %w", idx, err)
					return
				}

				msg, err := stream.Recv()
				if err != nil {
					errors <- fmt.Errorf("failed to recv for agent %d: %w", idx, err)
					return
				}

				if msg.GetOperatorHello().GetSessionId() == "" {
					errors <- fmt.Errorf("empty session id for agent %d", idx)
					return
				}

				errors <- nil
			}(i)
		}

		for i := 0; i < numAgents; i++ {
			err := <-errors
			if err != nil {
				t.Errorf("agent %d failed: %v", i, err)
			}
		}
	})
}
