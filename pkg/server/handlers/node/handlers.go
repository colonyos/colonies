package node

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/colonyos/colonies/pkg/backends"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	Validator() security.Validator
	NodeDB() database.NodeDatabase
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{
		server: server,
	}
}

// RegisterHandlers implements the HandlerRegistrar interface
func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	if err := handlerRegistry.Register(rpc.GetNodesPayloadType, h.HandleGetNodes); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetNodePayloadType, h.HandleGetNode); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetNodesByLocationPayloadType, h.HandleGetNodesByLocation); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleGetNodes(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetNodesMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get nodes, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get nodes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	nodes, err := h.server.NodeDB().GetNodes(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	// Convert nodes array to JSON
	jsonBytes, err := core.ConvertNodesToJSON(nodes)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "NodeCount": len(nodes)}).Debug("Getting nodes")

	h.server.SendHTTPReply(c, payloadType, jsonBytes)
}

func (h *Handlers) HandleGetNode(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetNodeMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get node, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get node, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	node, err := h.server.NodeDB().GetNodeByName(msg.ColonyName, msg.NodeName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if node == nil {
		h.server.HandleHTTPError(c, errors.New("Node not found"), http.StatusNotFound)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, node.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = node.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"NodeID": node.ID, "NodeName": node.Name}).Debug("Getting node")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetNodesByLocation(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetNodesByLocationMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get nodes by location, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get nodes by location, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	nodes, err := h.server.NodeDB().GetNodesByLocation(msg.ColonyName, msg.Location)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	// Convert nodes array to JSON
	jsonBytes, err := core.ConvertNodesToJSON(nodes)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "Location": msg.Location, "NodeCount": len(nodes)}).Debug("Getting nodes by location")

	h.server.SendHTTPReply(c, payloadType, jsonBytes)
}
