package grpc

import (
	"context"

	"github.com/colonyos/colonies/pkg/client/grpc/proto"
	log "github.com/sirupsen/logrus"
)

// ColoniesServiceImpl implements the gRPC ColoniesService
type ColoniesServiceImpl struct {
	proto.UnimplementedColoniesServiceServer
	handler RPCHandler
}

// RPCHandler is an interface for handling RPC messages
type RPCHandler interface {
	HandleRPC(jsonPayload string) (string, error)
}

// NewColoniesServiceImpl creates a new gRPC service implementation
func NewColoniesServiceImpl(handler RPCHandler) *ColoniesServiceImpl {
	return &ColoniesServiceImpl{
		handler: handler,
	}
}

// SendMessage handles incoming RPC messages
func (s *ColoniesServiceImpl) SendMessage(ctx context.Context, req *proto.RPCRequest) (*proto.RPCResponse, error) {
	if s.handler == nil {
		return &proto.RPCResponse{
			ErrorMessage: "no RPC handler configured",
		}, nil
	}

	// Process the JSON RPC message through the handler
	response, err := s.handler.HandleRPC(req.JsonPayload)
	if err != nil {
		log.WithError(err).Error("gRPC RPC handler error")
		return &proto.RPCResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	return &proto.RPCResponse{
		JsonPayload: response,
	}, nil
}

// CheckHealth handles health check requests
func (s *ColoniesServiceImpl) CheckHealth(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{
		Healthy: true,
		Version: "1.0.0",
	}, nil
}
