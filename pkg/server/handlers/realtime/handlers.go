package realtime

import (
	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/security"
)

// RealtimeHandler interface for backend-specific realtime handling
type RealtimeHandler interface {
	HandleWSRequest(c backends.Context)
}

// Server interface for servers that support realtime handlers
type Server interface {
	Validator() security.Validator
	ProcessDB() database.ProcessDatabase
	RealtimeHandler() RealtimeHandler
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{server: server}
}

// HandleWSRequest delegates to the backend-specific realtime handler
func (h *Handlers) HandleWSRequest(c backends.Context) {
	h.server.RealtimeHandler().HandleWSRequest(c)
}