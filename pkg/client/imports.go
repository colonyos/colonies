package client

import (
	"github.com/colonyos/colonies/pkg/client/gin"
)

func init() {
	// Register the gin backend factory when this package is imported
	factory := gin.GetGinClientBackendFactory()
	RegisterBackendFactory(factory)
}