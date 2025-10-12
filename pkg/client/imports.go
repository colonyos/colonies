package client

import (
	"github.com/colonyos/colonies/pkg/client/coap"
	"github.com/colonyos/colonies/pkg/client/gin"
	"github.com/colonyos/colonies/pkg/client/grpc"
)

func init() {
	// Register the gin backend factory when this package is imported
	ginFactory := gin.GetGinClientBackendFactory()
	RegisterBackendFactory(ginFactory)

	// Register the gRPC backend factory when this package is imported
	grpcFactory := grpc.GetGRPCClientBackendFactory()
	RegisterBackendFactory(grpcFactory)

	// Register the CoAP backend factory when this package is imported
	coapFactory := coap.GetCoAPClientBackendFactory()
	RegisterBackendFactory(coapFactory)
}