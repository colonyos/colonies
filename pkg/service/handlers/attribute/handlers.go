package attribute

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/service/registry"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ColoniesServer interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	Validator() security.Validator
	AttributeController() Controller
}

type Controller interface {
	GetProcess(processID string) (*core.Process, error)
	AddAttribute(attribute *core.Attribute) (*core.Attribute, error)
	GetAttribute(attributeID string) (*core.Attribute, error)
}

type Handlers struct {
	server ColoniesServer
}

func NewHandlers(server ColoniesServer) *Handlers {
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

func (h *Handlers) HandleAddAttribute(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
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

	process, err := h.server.AttributeController().GetProcess(msg.Attribute.TargetID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add attribute, process is nil"), http.StatusInternalServerError)
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

	addedAttribute, err := h.server.AttributeController().AddAttribute(&msg.Attribute)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = addedAttribute.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"AttributeId": msg.Attribute.ID}).Debug("Adding attribute")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetAttribute(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
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

	attribute, err := h.server.AttributeController().GetAttribute(msg.AttributeID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	process, err := h.server.AttributeController().GetProcess(attribute.TargetID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get attribute, process is nil"), http.StatusInternalServerError)
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

