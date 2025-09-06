package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/service/registry"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	GetStatistics() (*core.Statistics, error)
	GetEtcdServer() EtcdServer
}

type EtcdServer interface {
	CurrentCluster() cluster.Config
}

type Validator interface {
	RequireServerOwner(recoveredID string, serverID string) error
}

type ColoniesServer interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	GetServerID() (string, error)
	Validator() Validator
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
	if err := handlerRegistry.Register(rpc.GetStatisiticsPayloadType, h.HandleStatistics); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetClusterPayloadType, h.HandleGetCluster); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleVersion(c *gin.Context, payloadType string, jsonString string) {
	msg, err := rpc.CreateVersionMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get server version, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get server version, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	versionMsg := rpc.CreateVersionMsg(build.BuildVersion, build.BuildTime)

	jsonString, err = versionMsg.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.Debug("Getting server version")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleStatistics(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetStatisticsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get statistics, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get server version, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	stat, err := h.server.Controller().GetStatistics()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = stat.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetCluster(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetClusterMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get cluster info, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get cluster info, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	cluster := h.server.Controller().GetEtcdServer().CurrentCluster()
	jsonString, err = cluster.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}