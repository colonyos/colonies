package colony

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/service/registry"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	AddColony(colony *core.Colony) (*core.Colony, error)
	RemoveColony(colonyName string) error
	GetColonies() ([]*core.Colony, error)
	GetColony(colonyName string) (*core.Colony, error)
	GetColonyStatistics(colonyName string) (*core.Statistics, error)
}

type ColoniesServer interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	GetServerID() (string, error)
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	Validator() security.Validator
	ColonyDB() database.ColonyDatabase
	Controller() Controller
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
	if err := handlerRegistry.Register(rpc.AddColonyPayloadType, h.HandleAddColony); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveColonyPayloadType, h.HandleRemoveColony); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetColoniesPayloadType, h.HandleGetColonies); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetColonyPayloadType, h.HandleGetColony); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetColonyStatisticsPayloadType, h.HandleColonyStatistics); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddColony(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddColonyMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add colony, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add colony, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	serverID, err := h.server.GetServerID()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	err = h.server.Validator().RequireServerOwner(recoveredID, serverID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Colony == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add colony, colony is <nil>"), http.StatusBadRequest)
		return
	}

	if len(msg.Colony.ID) != 64 {
		h.server.HandleHTTPError(c, errors.New("Failed to add colony, invalid colony Id length"), http.StatusBadRequest)
		return
	}

	colonyExist, err := h.server.ColonyDB().GetColonyByName(msg.Colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if colonyExist != nil {
		if h.server.HandleHTTPError(c, errors.New("A Colony with name <"+msg.Colony.Name+"> already exists"), http.StatusBadRequest) {
			return
		}
	}

	addedColony, err := h.server.Controller().AddColony(msg.Colony)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if addedColony == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add colony, addedColony is <nil>"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedColony.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.Colony.Name, "ColonyId": addedColony.ID}).Debug("Adding colony")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRemoveColony(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveColonyMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove colony, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove colony, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	serverID, err := h.server.GetServerID()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	err = h.server.Validator().RequireServerOwner(recoveredID, serverID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := h.server.ColonyDB().GetColonyByName(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if colony == nil {
		if h.server.HandleHTTPError(c, errors.New("Colony with name <"+msg.ColonyName+"> not found"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.Controller().RemoveColony(colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": colony.ID, "ColonyName": colony.Name}).Debug("Removing colony")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleGetColonies(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColoniesMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get colonies, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get colonies, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	serverID, err := h.server.GetServerID()

	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	err = h.server.Validator().RequireServerOwner(recoveredID, serverID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colonies, err := h.server.Controller().GetColonies()
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertColonyArrayToJSON(colonies)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.Debug("Getting colonies")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetColony(c *gin.Context, recoveredID, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColonyMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get colony, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get colony, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := h.server.Controller().GetColony(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if colony == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get colony, colony is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = colony.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName}).Debug("Getting colony")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleColonyStatistics(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColonyStatisticsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get colony statistics, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get colony statistics, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	colony, err := h.server.ColonyDB().GetColonyByName(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if colony == nil {
		if h.server.HandleHTTPError(c, errors.New("Colony with name <"+msg.ColonyName+"> not found"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.Validator().RequireMembership(recoveredID, colony.Name, true)
	if err != nil {
		return
	}

	stat, err := h.server.Controller().GetColonyStatistics(colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = stat.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName}).Debug("Getting colony statistics")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}