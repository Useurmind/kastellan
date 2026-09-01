package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/spf13/cobra"

	"github.com/kastellan/kastellan/agent/handlers"
	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

const (
	defaultAgentVersion = "0.1.0"
	defaultServerPort   = 443
)

var (
	serverAddress     string
	hostName          string
	hostHostname      string
	agentID           string
	agentVersion      string
	certPath          string
	keyPath           string
	caPath            string
	enrollmentToken   string
	enableEnrollment  bool
	heartbeatInterval time.Duration
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Kastellan Agent",
	Long: `Run the Kastellan Agent which connects to the operator and manages workloads.

The agent establishes a bidirectional gRPC connection over mTLS to the Kastellan
Operator and receives desired state for workloads assigned to this host.`,
	RunE: runAgent,
}

func init() {
	runCmd.Flags().StringVar(&serverAddress, "server", "", "Operator server address (host:port)")
	runCmd.Flags().StringVar(&hostName, "host-name", "", "External host name (required)")
	runCmd.Flags().StringVar(&hostHostname, "host-hostname", "", "External host hostname (optional, defaults to host-name)")
	runCmd.Flags().StringVar(&agentID, "agent-id", "", "Agent identifier (optional, auto-generated if not set)")
	runCmd.Flags().StringVar(&agentVersion, "agent-version", defaultAgentVersion, "Agent version")
	runCmd.Flags().StringVar(&certPath, "cert-path", "", "Path to client certificate (PEM format)")
	runCmd.Flags().StringVar(&keyPath, "key-path", "", "Path to client private key (PEM format)")
	runCmd.Flags().StringVar(&caPath, "ca-path", "", "Path to CA certificate (PEM format)")
	runCmd.Flags().StringVar(&enrollmentToken, "enrollment-token", "", "Enrollment token for initial connection")
	runCmd.Flags().DurationVar(&heartbeatInterval, "heartbeat-interval", 30*time.Second, "Heartbeat interval")
	runCmd.Flags().BoolVar(&enableEnrollment, "enrollment", false, "Enable enrollment mode (for first-time connection)")
}

func runAgent(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	return runAgentLoop(ctx)
}

func runAgentLoop(ctx context.Context) error {
	if serverAddress == "" {
		if val, ok := os.LookupEnv("KASTELLAN_OPERATOR_ADDRESS"); ok {
			serverAddress = val
		} else {
			serverAddress = fmt.Sprintf("localhost:%d", defaultServerPort)
		}
	}

	if hostName == "" {
		if val, ok := os.LookupEnv("KASTELLAN_HOST_NAME"); ok {
			hostName = val
		} else {
			return errors.New("host name is required (use --host-name or KASTELLAN_HOST_NAME)")
		}
	}

	if hostHostname == "" {
		hostHostname = hostName
	}

	if agentID == "" {
		agentID = generateAgentID()
	}

	if certPath != "" {
		certPath, _ = filepath.Abs(certPath)
	}
	if keyPath != "" {
		keyPath, _ = filepath.Abs(keyPath)
	}
	if caPath != "" {
		caPath, _ = filepath.Abs(caPath)
	}

	if enableEnrollment {
		if enrollmentToken == "" {
			return errors.New("enrollment token is required when enrollment mode is enabled")
		}
		if caPath == "" {
			return errors.New("CA certificate path is required for enrollment")
		}
		if certPath != "" || keyPath != "" {
			return errors.New("certificates should not be provided during enrollment")
		}
	} else {
		if certPath == "" || keyPath == "" || caPath == "" {
			return errors.New("certificates are required (cert-path, key-path, ca-path)")
		}
	}

	fmt.Printf("Starting Kastellan Agent\n")
	fmt.Printf("  Server: %s\n", serverAddress)
	fmt.Printf("  Host: %s (%s)\n", hostName, hostHostname)
	fmt.Printf("  Agent: %s v%s\n", agentID, agentVersion)

	fmt.Printf("  Certificates:\n")
	if certPath != "" {
		fmt.Printf("    Client: %s\n", certPath)
		fmt.Printf("    Key: %s\n", keyPath)
	}
	fmt.Printf("    CA: %s\n", caPath)

	fmt.Printf("  Mode: ")
	if enableEnrollment {
		fmt.Printf("Enrollment (using token)\n")
	} else {
		fmt.Printf("mTLS (using certificates)\n")
	}
	fmt.Printf("Starting connection loop...\n\n")

	if enableEnrollment {
		return runWithEnrollment(ctx)
	}

	return runWithMTLS(ctx)
}

func runWithMTLS(ctx context.Context) error {
	tlsConfig, err := createTLSConfig(certPath, keyPath, caPath)
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}

	conn, err := grpc.DialContext(
		ctx,
		serverAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}
	defer conn.Close()

	stream, err := conn.NewStream(ctx, &grpc.StreamDesc{}, "/kastellan.AgentService/Connect")
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	registry := handlers.NewHandlerRegistry()

	heartbeatHandler := handlers.NewOperatorHelloHandler()
	desiredStateHandler := handlers.NewDesiredStateHandler()
	reconcileHandler := handlers.NewReconcileRequestHandler()
	connectionCloseHandler := handlers.NewConnectionCloseHandler()

	registry.Register(heartbeatHandler)
	registry.Register(desiredStateHandler)
	registry.Register(reconcileHandler)
	registry.Register(connectionCloseHandler)

	msgHandler := handlers.NewStreamMessageHandler(ctx, stream, registry)
	if err := msgHandler.Start(); err != nil {
		return fmt.Errorf("failed to start message handler: %w", err)
	}

	if err := sendAgentHello(ctx, msgHandler); err != nil {
		return fmt.Errorf("failed to send AgentHello: %w", err)
	}

	fmt.Println("Agent connected to operator")
	fmt.Println()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			msgHandler.Stop()
			return nil
		case <-ticker.C:
			if err := msgHandler.SendHeartbeat(ctx, "", hostName); err != nil {
				fmt.Printf("Failed to send heartbeat: %v\n", err)
			}
		}
	}
}

func runWithEnrollment(ctx context.Context) error {
	tlsConfig, err := createTLSConfig("", "", caPath)
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}

	conn, err := grpc.DialContext(
		ctx,
		serverAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}
	defer conn.Close()

	stream, err := conn.NewStream(ctx, &grpc.StreamDesc{}, "/kastellan.AgentService/Connect")
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	registry := handlers.NewHandlerRegistry()

	enrollmentHandler := handlers.NewEnrollmentResponseHandler()
	registry.Register(enrollmentHandler)

	msgHandler := handlers.NewStreamMessageHandler(ctx, stream, registry)
	if err := msgHandler.Start(); err != nil {
		return fmt.Errorf("failed to start message handler: %w", err)
	}

	if err := sendEnrollmentRequest(ctx, msgHandler); err != nil {
		return fmt.Errorf("failed to send EnrollmentRequest: %w", err)
	}

	fmt.Println("Enrollment request sent, waiting for response...")
	fmt.Println()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			msgHandler.Stop()
			return nil
		case <-ticker.C:
			if err := msgHandler.SendHeartbeat(ctx, "", hostName); err != nil {
				fmt.Printf("Failed to send heartbeat: %v\n", err)
			}
		}
	}
}

func createTLSConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
	if certPath == "" && keyPath == "" {
		return &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS13,
		}, nil
	}

	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append CA certificate")
	}

	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			MinVersion:   tls.VersionTLS13,
		}, nil
	}

	return &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS13,
	}, nil
}

func generateAgentID() string {
	return fmt.Sprintf("agent-%s", time.Now().Format("20060102150405"))
}

func sendAgentHello(ctx context.Context, handler *handlers.StreamMessageHandler) error {
	capabilities := []string{
		messages.CapabilityPlayKube,
		messages.CapabilityReplace,
	}

	return handler.SendAgentHello(ctx, agentID, agentVersion, hostName, hostHostname, capabilities)
}

func sendEnrollmentRequest(ctx context.Context, handler *handlers.StreamMessageHandler) error {
	return handler.SendEnrollmentRequest(ctx, enrollmentToken, agentID, agentVersion, hostName, hostHostname)
}

func unmarshalMessage(data []byte) (messages.MessageType, interface{}, error) {
	var raw struct {
		Type messages.MessageType `json:"type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal message type: %w", err)
	}

	switch raw.Type {
	case messages.MessageTypeOperatorHello:
		var msg messages.OperatorHello
		if err := json.Unmarshal(data, &msg); err != nil {
			return "", nil, fmt.Errorf("failed to unmarshal OperatorHello: %w", err)
		}
		return raw.Type, &msg, nil

	case messages.MessageTypeDesiredState:
		var msg messages.DesiredState
		if err := json.Unmarshal(data, &msg); err != nil {
			return "", nil, fmt.Errorf("failed to unmarshal DesiredState: %w", err)
		}
		return raw.Type, &msg, nil

	case messages.MessageTypeReconcileRequest:
		var msg messages.ReconcileRequest
		if err := json.Unmarshal(data, &msg); err != nil {
			return "", nil, fmt.Errorf("failed to unmarshal ReconcileRequest: %w", err)
		}
		return raw.Type, &msg, nil

	case messages.MessageTypeConnectionClose:
		var msg messages.ConnectionClose
		if err := json.Unmarshal(data, &msg); err != nil {
			return "", nil, fmt.Errorf("failed to unmarshal ConnectionClose: %w", err)
		}
		return raw.Type, &msg, nil

	case messages.MessageTypeEnrollmentResponse:
		var msg messages.EnrollmentResponse
		if err := json.Unmarshal(data, &msg); err != nil {
			return "", nil, fmt.Errorf("failed to unmarshal EnrollmentResponse: %w", err)
		}
		return raw.Type, &msg, nil

	default:
		return "", nil, fmt.Errorf("unknown message type: %s", raw.Type)
	}
}
