package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/client/grpc/proto"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClientBackend implements client backend using gRPC
type GRPCClientBackend struct {
	conn          *grpclib.ClientConn
	client        proto.ColoniesServiceClient
	host          string
	port          int
	isInsecure    bool
	skipTLSVerify bool
}

// NewGRPCClientBackend creates a new gRPC client backend
func NewGRPCClientBackend(config *backends.ClientConfig) (*GRPCClientBackend, error) {
	if config.BackendType != backends.GRPCClientBackendType {
		return nil, errors.New("invalid backend type for gRPC client")
	}

	backend := &GRPCClientBackend{
		host:          config.Host,
		port:          config.Port,
		isInsecure:    config.Insecure,
		skipTLSVerify: config.SkipTLSVerify,
	}

	// Create connection options
	var opts []grpclib.DialOption

	if config.Insecure {
		// Use insecure connection
		opts = append(opts, grpclib.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Use TLS
		tlsConfig := &tls.Config{}
		if config.SkipTLSVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		opts = append(opts, grpclib.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Connect to gRPC server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := grpclib.Dial(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	backend.conn = conn
	backend.client = proto.NewColoniesServiceClient(conn)

	return backend, nil
}

// SendRawMessage sends a raw JSON message via gRPC
func (g *GRPCClientBackend) SendRawMessage(jsonString string, insecure bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &proto.RPCRequest{
		JsonPayload: jsonString,
	}

	resp, err := g.client.SendMessage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("gRPC call failed: %w", err)
	}

	if resp.ErrorMessage != "" {
		return "", errors.New(resp.ErrorMessage)
	}

	return resp.JsonPayload, nil
}

// SendMessage sends an RPC message with authentication via gRPC
func (g *GRPCClientBackend) SendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	// Create RPC message
	var rpcMsg *rpc.RPCMsg
	var err error
	if insecure {
		rpcMsg, err = rpc.CreateInsecureRPCMsg(method, jsonString)
		if err != nil {
			return "", err
		}
	} else {
		rpcMsg, err = rpc.CreateRPCMsg(method, jsonString, prvKey)
		if err != nil {
			return "", err
		}
	}

	jsonString, err = rpcMsg.ToJSON()
	if err != nil {
		return "", err
	}

	// Send via gRPC
	if ctx == nil {
		ctx = context.Background()
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req := &proto.RPCRequest{
		JsonPayload: jsonString,
	}

	resp, err := g.client.SendMessage(ctxWithTimeout, req)
	if err != nil {
		return "", fmt.Errorf("gRPC call failed: %w", err)
	}

	if resp.ErrorMessage != "" {
		return "", errors.New(resp.ErrorMessage)
	}

	// Parse response
	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(resp.JsonPayload)
	if err != nil {
		return "", errors.New("expected a valid Colonies RPC message, but got: " + resp.JsonPayload)
	}

	if rpcReplyMsg.Error {
		failure, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
		if err != nil {
			return "", err
		}

		return "", &core.ColoniesError{Status: failure.Status, Message: failure.Message}
	}

	return rpcReplyMsg.DecodePayload(), nil
}

// CheckHealth checks the health of the gRPC server
func (g *GRPCClientBackend) CheckHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.HealthRequest{}

	resp, err := g.client.CheckHealth(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !resp.Healthy {
		return errors.New("server reported unhealthy status")
	}

	return nil
}

// Close closes the gRPC connection and cleans up services
func (g *GRPCClientBackend) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

// GRPCClientBackendFactory creates gRPC client backends
type GRPCClientBackendFactory struct{}

// NewGRPCClientBackendFactory creates a new gRPC client backend factory
func NewGRPCClientBackendFactory() *GRPCClientBackendFactory {
	return &GRPCClientBackendFactory{}
}

// CreateBackend creates a new gRPC client backend
func (f *GRPCClientBackendFactory) CreateBackend(config *backends.ClientConfig) (backends.ClientBackend, error) {
	return NewGRPCClientBackend(config)
}

// GetBackendType returns the backend type this factory creates
func (f *GRPCClientBackendFactory) GetBackendType() backends.ClientBackendType {
	return backends.GRPCClientBackendType
}

// Compile-time check that GRPCClientBackend implements the required interfaces
var _ backends.ClientBackend = (*GRPCClientBackend)(nil)
