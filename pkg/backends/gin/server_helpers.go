package gin

import (
	"net/http"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ServerHelpers provides gin-specific server helper functions
type ServerHelpers struct{}

// NewServerHelpers creates new gin server helpers
func NewServerHelpers() *ServerHelpers {
	return &ServerHelpers{}
}

// HandleHTTPErrorGin handles HTTP errors for gin contexts
func (h *ServerHelpers) HandleHTTPErrorGin(c *gin.Context, err error, errorCode int) bool {
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("HTTP Error")
		c.JSON(errorCode, gin.H{"error": err.Error()})
		return true
	}
	return false
}

// HandleHTTPErrorContext handles HTTP errors for generic contexts by checking if they're gin contexts
func (h *ServerHelpers) HandleHTTPErrorContext(c backends.Context, err error, errorCode int) bool {
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("HTTP Error") 
		c.JSON(errorCode, map[string]interface{}{"error": err.Error()})
		return true
	}
	return false
}

// SendHTTPReplyGin sends HTTP reply for gin contexts
func (h *ServerHelpers) SendHTTPReplyGin(c *gin.Context, payloadType string, jsonString string) {
	c.Header("Content-Type", "application/json")
	c.Header("Payload-Type", payloadType)
	c.String(http.StatusOK, jsonString)
}

// SendEmptyHTTPReplyGin sends empty HTTP reply for gin contexts  
func (h *ServerHelpers) SendEmptyHTTPReplyGin(c *gin.Context, payloadType string) {
	c.Header("Content-Type", "application/json")
	c.Header("Payload-Type", payloadType)
	c.Status(http.StatusOK)
}

// ExtractGinContext extracts gin.Context from a generic Context
func (h *ServerHelpers) ExtractGinContext(c backends.Context) (*gin.Context, bool) {
	if ginAdapter, ok := c.(*ContextAdapter); ok {
		return ginAdapter.GinContext(), true
	}
	return nil, false
}

// HandleContextUnion handles both gin.Context and backends.Context types
func (h *ServerHelpers) HandleContextUnion(c interface{}) (backends.Context, *gin.Context, bool) {
	switch ctx := c.(type) {
	case *gin.Context:
		return NewContextAdapter(ctx), ctx, true
	case backends.Context:
		if ginCtx, ok := h.ExtractGinContext(ctx); ok {
			return ctx, ginCtx, true
		}
		return ctx, nil, true
	default:
		return nil, nil, false
	}
}