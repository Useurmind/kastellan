// Package client provides the gRPC client for the Kastellan Agent Protocol.
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// Client represents the agent protocol client.
type Client struct {
	// Configuration
	serverAddress string
	enrollmentToken string
	agentID       string
	agentVersion  string
	hostName      string
	hostHostname  string

	// TLS configuration
	certPath      string
	keyPath       string
	caPath        string
	serverName    string

	// Connection state
	conn          *grpc.ClientConn
	stream        grpc.ClientStream
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex

	// Reconnection settings
	reconnectDelay time.Duration
	maxDelay       time.Duration

	// Heartbeat settings
	heartbeatInterval time.Duration
	heartbeatStop     chan struct{}
	heartbeatWG       sync.WaitGroup

	// Status reporting
	statusStop chan struct{}
	statusWG   sync.WaitGroup

	// Event channels
	connectedCh   chan struct{}
	disconnectedCh chan struct{}
	messageCh     chan messages.Message
	errorCh       chan error
}

// Message represents a protocol message that can be sent or received.
type Message interface {
	GetType() messages.MessageType
}

// New creates a new agent protocol client.
func New(serverAddress, agentID, agentVersion, hostName, hostHostname string) *Client {
	return &Client{
		serverAddress:   serverAddress,
		agentID:         agentID,
		agentVersion:    agentVersion,
		hostName:        hostName,
		hostHostname:    hostHostname,
		reconnectDelay:  time.Second,
		maxDelay:        time.Minute,
		heartbeatInterval: 30 * time.Second,
		connectedCh:     make(chan struct{}, 1),
		disconnectedCh:  make(chan struct{}, 1),
		messageCh:       make(chan messages.Message, 100),
		errorCh:         make(chan error, 100),
	}
}

// WithEnrollmentToken sets the enrollment token for initial connection.
func (c *Client) WithEnrollmentToken(token string) *Client {
	c.enrollmentToken = token
	return c
}

// WithCertificates sets the paths to TLS certificates.
func (c *Client) WithCertificates(certPath, keyPath, caPath string) *Client {
	c.certPath = certPath
	c.keyPath = keyPath
	c.caPath = caPath
	return c
}

// WithServerName sets the server name for TLS verification.
func (c *Client) WithServerName(serverName string) *Client {
	c.serverName = serverName
	return c
}

// WithReconnectDelay sets the initial reconnection delay.
func (c *Client) WithReconnectDelay(delay time.Duration) *Client {
	c.reconnectDelay = delay
	return c
}

// WithMaxReconnectDelay sets the maximum reconnection delay.
func (c *Client) WithMaxReconnectDelay(delay time.Duration) *Client {
	c.maxDelay = delay
	return c
}

// WithHeartbeatInterval sets the heartbeat interval.
func (c *Client) WithHeartbeatInterval(interval time.Duration) *Client {
	c.heartbeatInterval = interval
	return c
}

// Connect establishes a connection to the operator.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.mu.Unlock()

	return c.connectWithReconnect(c.ctx)
}

// connectWithReconnect handles the reconnection loop.
func (c *Client) connectWithReconnect(ctx context.Context) error {
	delay := c.reconnectDelay

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := c.connectOnce(ctx); err != nil {
			c.notifyDisconnected()

			// Log the error (would use proper logging in production)
			fmt.Printf("Connection failed: %v, retrying in %v\n", err, delay)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Exponential backoff with jitter
				delay = min(delay*2, c.maxDelay)
			}
			continue
		}

		// Connection successful
		c.notifyConnected()
		delay = c.reconnectDelay // Reset delay

		// Wait for connection to be closed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.disconnectedCh:
			// Connection was closed, restart
			continue
		}
	}
}

// connectOnce establishes a single connection.
func (c *Client) connectOnce(ctx context.Context) error {
	// Create TLS credentials
	tlsConfig, err := c.createTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}

	// Create gRPC connection
	conn, err := grpc.DialContext(
		ctx,
		c.serverAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// Create bidirectional stream
	stream, err := c.createStream(ctx, conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create stream: %w", err)
	}

	c.mu.Lock()
	c.stream = stream
	c.mu.Unlock()

	// Start heartbeat and status goroutines
	c.startHeartbeat(ctx)
	c.startStatusReporting(ctx)

	return nil
}

// createTLSConfig creates the TLS configuration for mTLS.
func (c *Client) createTLSConfig() (*tls.Config, error) {
	// Load CA certificate
	caCert, err := c.loadCertificate(c.caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Load client certificate
	clientCert, err := c.loadCertificate(c.certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load client private key
	clientKey, err := c.loadPrivateKey(c.keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client private key: %w", err)
	}

	// Create TLS config
	config := &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{clientCert},
			PrivateKey:  clientKey,
			Leaf:        clientCert,
		}},
		RootCAs:    caCert,
		ServerName: c.serverName,
		MinVersion: tls.VersionTLS13,
	}

	return config, nil
}

// loadCertificate loads a certificate from a file.
func (c *Client) loadCertificate(path string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	data, err := io.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("failed to append certificate from %s", path)
	}

	return pool, nil
}

// loadPrivateKey loads a private key from a file.
func (c *Client) loadPrivateKey(path string) (interface{}, error) {
	data, err := io.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", path)
	}

	// Try to parse as RSA private key
	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	// Try to parse as EC private key
	if block.Type == "EC PRIVATE KEY" {
		return x509.ParseECPrivateKey(block.Bytes)
	}

	// Try to parse as PKCS#8 private key
	if block.Type == "PRIVATE KEY" {
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	}

	return nil, fmt.Errorf("unknown private key type: %s", block.Type)
}

// createStream creates the bidirectional gRPC stream.
func (c *Client) createStream(ctx context.Context, conn *grpc.ClientConn) (grpc.ClientStream, error) {
	// Create metadata with agent info
	md := metadata.New(map[string]string{
		"agent-id":      c.agentID,
		"agent-version": c.agentVersion,
		"host-name":     c.hostName,
	})

	ctx = metadata.NewOutgoingContext(ctx, md)

	// Create stream (this would use the actual gRPC service)
	// For now, we'll use a placeholder
	stream, err := conn.NewStream(ctx, &grpc.StreamDesc{}, "/kastellan.AgentService/Connect")
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// startHeartbeat starts the heartbeat goroutine.
func (c *Client) startHeartbeat(ctx context.Context) {
	c.heartbeatStop = make(chan struct{})
	c.heartbeatWG.Add(1)

	go func() {
		defer c.heartbeatWG.Done()

		ticker := time.NewTicker(c.heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.heartbeatStop:
				return
			case <-ticker.C:
				if err := c.sendHeartbeat(ctx); err != nil {
					fmt.Printf("Failed to send heartbeat: %v\n", err)
					// Don't stop on heartbeat failure
				}
			}
		}
	}()
}

// sendHeartbeat sends a heartbeat message to the operator.
func (c *Client) sendHeartbeat(ctx context.Context) error {
	heartbeat := &messages.Heartbeat{
		Type:      messages.MessageTypeHeartbeat,
		SessionID: c.getSessionID(),
		Timestamp: time.Now(),
	}

	// Send heartbeat through the stream
	// This would use the actual gRPC message sending mechanism
	return c.sendMessage(ctx, heartbeat)
}

// startStatusReporting starts the status reporting goroutine.
func (c *Client) startStatusReporting(ctx context.Context) {
	c.statusStop = make(chan struct{})
	c.statusWG.Add(1)

	go func() {
		defer c.statusWG.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.statusStop:
				return
			case msg := <-c.messageCh:
				if err := c.sendMessage(ctx, msg); err != nil {
					fmt.Printf("Failed to send message: %v\n", err)
				}
			}
		}
	}()
}

// sendMessage sends a message through the stream.
func (c *Client) sendMessage(ctx context.Context, msg messages.Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.stream == nil {
		return fmt.Errorf("stream is not connected")
	}

	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send message through stream
	// This would use the actual gRPC stream sending mechanism
	_, err = c.stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}

	return nil
}

// getSessionID returns the current session ID.
func (c *Client) getSessionID() string {
	// This would be set during enrollment or connection
	return ""
}

// notifyConnected notifies listeners that the client is connected.
func (c *Client) notifyConnected() {
	select {
	case c.connectedCh <- struct{}{}:
	default:
	}
}

// notifyDisconnected notifies listeners that the client is disconnected.
func (c *Client) notifyDisconnected() {
	select {
	case c.disconnectedCh <- struct{}{}:
	default:
	}
}

// Close closes the connection and stops all goroutines.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop heartbeat and status goroutines
	if c.heartbeatStop != nil {
		close(c.heartbeatStop)
	}
	if c.statusStop != nil {
		close(c.statusStop)
	}

	// Wait for goroutines to finish
	c.heartbeatWG.Wait()
	c.statusWG.Wait()

	// Cancel context
	if c.cancel != nil {
		c.cancel()
	}

	// Close stream and connection
	if c.stream != nil {
		c.stream.CloseSend()
	}
	if c.conn != nil {
		c.conn.Close()
	}

	return nil
}

// GetConnectedChannel returns a channel that receives notifications when connected.
func (c *Client) GetConnectedChannel() <-chan struct{} {
	return c.connectedCh
}

// GetDisconnectedChannel returns a channel that receives notifications when disconnected.
func (c *Client) GetDisconnectedChannel() <-chan struct{} {
	return c.disconnectedCh
}

// GetMessageChannel returns a channel for sending messages to the operator.
func (c *Client) GetMessageChannel() chan<- messages.Message {
	return c.messageCh
}

// GetErrorChannel returns a channel for receiving errors.
func (c *Client) GetErrorChannel() <-chan error {
	return c.errorCh
}

// min returns the minimum of two durations.
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
