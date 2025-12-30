package attribute

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
	ProcessDB() database.ProcessDatabase
	AttributeDB() database.AttributeDatabase
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
	if err := handlerRegistry.Register(rpc.AddAttributePayloadType, h.HandleAddAttribute); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetAttributePayloadType, h.HandleGetAttribute); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddAttribute(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddAttributeMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add attribute, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add attribute, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := h.server.ProcessDB().GetProcessByID(msg.Attribute.TargetID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add attribute, process not found"), http.StatusNotFound)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.State != core.RUNNING {
		err := errors.New("Failed to add attribute, process is not running")
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		err := errors.New("Failed to add attribute, only executor with id <" + process.AssignedExecutorID + "> is allowed to set attributes")
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	msg.Attribute.GenerateID()
	msg.Attribute.TargetProcessGraphID = process.ProcessGraphID

	err = h.server.AttributeDB().AddAttribute(msg.Attribute)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	addedAttribute, err := h.server.AttributeDB().GetAttributeByID(msg.Attribute.ID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = addedAttribute.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"AttributeId": msg.Attribute.ID}).Debug("Adding attribute")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetAttribute(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetAttributeMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get attribute, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get attribute, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	attribute, err := h.server.AttributeDB().GetAttributeByID(msg.AttributeID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	process, err := h.server.ProcessDB().GetProcessByID(attribute.TargetID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get attribute, process not found"), http.StatusNotFound)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = attribute.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"AttributeId": msg.AttributeID}).Debug("Getting attribute")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

