// Package client provides connection management for the Kastellan Agent Protocol.
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Connection represents a gRPC connection to the operator.
type Connection struct {
	// Configuration
	serverAddress string
	serverName    string
	certPath      string
	keyPath       string
	caPath        string

	// TLS configuration
	tlsConfig *tls.Config

	// gRPC connection
	conn *grpc.ClientConn

	// Connection state
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	connected bool
}

// NewConnection creates a new connection manager.
func NewConnection(serverAddress string) *Connection {
	return &Connection{
		serverAddress: serverAddress,
	}
}

// WithServerName sets the server name for TLS verification.
func (c *Connection) WithServerName(serverName string) *Connection {
	c.serverName = serverName
	return c
}

// WithCertificates sets the paths to TLS certificates.
func (c *Connection) WithCertificates(certPath, keyPath, caPath string) *Connection {
	c.certPath = certPath
	c.keyPath = keyPath
	c.caPath = caPath
	return c
}

// CreateTLSConfig creates the TLS configuration for mTLS.
func (c *Connection) CreateTLSConfig() (*tls.Config, error) {
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
func (c *Connection) loadCertificate(path string) (*x509.CertPool, error) {
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
func (c *Connection) loadPrivateKey(path string) (interface{}, error) {
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

// Connect establishes a gRPC connection to the operator.
func (c *Connection) Connect(ctx context.Context) error {
	c.mu.Lock()
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.mu.Unlock()

	// Create TLS credentials
	tlsConfig, err := c.CreateTLSConfig()
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
	c.connected = true
	c.mu.Unlock()

	return nil
}

// Disconnect closes the gRPC connection.
func (c *Connection) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.connected = false
	return nil
}

// IsConnected returns whether the connection is active.
func (c *Connection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetConn returns the gRPC connection.
func (c *Connection) GetConn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// NewStream creates a new bidirectional stream.
func (c *Connection) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	return c.conn.NewStream(ctx, desc, method, opts...)
}

// SendRequest sends a request through the stream.
func (c *Connection) SendRequest(ctx context.Context, stream grpc.ClientStream, data []byte) error {
	_, err := stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}
	return nil
}

// RecvResponse receives a response from the stream.
func (c *Connection) RecvResponse(ctx context.Context, stream grpc.ClientStream) ([]byte, error) {
	data := make([]byte, 4096)
	n, err := stream.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}
	return data[:n], nil
}

// SendMetadata sends metadata with the request.
func (c *Connection) SendMetadata(ctx context.Context, md metadata.MD) context.Context {
	return metadata.NewOutgoingContext(ctx, md)
}

// CreateMetadata creates metadata from agent information.
func (c *Connection) CreateMetadata(agentID, agentVersion, hostName string) metadata.MD {
	return metadata.New(map[string]string{
		"agent-id":      agentID,
		"agent-version": agentVersion,
		"host-name":     hostName,
	})
}

// Wait for connection to be established or timeout.
func (c *Connection) WaitForConnection(ctx context.Context, timeout time.Duration) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutChan:
			return fmt.Errorf("connection timeout")
		case <-ticker.C:
			if c.IsConnected() {
				return nil
			}
		}
	}
}

// GetServerAddress returns the server address.
func (c *Connection) GetServerAddress() string {
	return c.serverAddress
}

// GetServerName returns the server name.
func (c *Connection) GetServerName() string {
	return c.serverName
}
