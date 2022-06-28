package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to add runtime, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add runtime, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Runtime == nil {
		server.handleHTTPError(c, errors.New("Failed to add runtime, runtime is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.Runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	addedRuntime, err := server.controller.addRuntime(msg.Runtime)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedRuntime == nil {
		server.handleHTTPError(c, errors.New("Failed to add runtime, addedRuntime is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedRuntime.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyID": msg.Runtime.ColonyID, "RuntimeID": addedRuntime.ID}).Info("Adding runtime")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetRuntimesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetRuntimesMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get runtimes, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get runtimes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID, false)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyID)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	runtimes, err := server.controller.getRuntimeByColonyID(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertRuntimeArrayToJSON(runtimes)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyID": msg.ColonyID}).Info("Getting runtimes")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get runtime, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get runtime, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleHTTPError(c, errors.New("Failed to get runtime, runtime is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, runtime.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"RuntimeID": runtime.ID}).Info("Getting runtime")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleApproveRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateApproveRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to approve runtime, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to approve runtime, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleHTTPError(c, errors.New("Failed to approve runtime, runtime is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.approveRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"RuntimeID": runtime.ID}).Info("Approving runtime")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleRejectRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRejectRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to reject runtime, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to reject runtime, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleHTTPError(c, errors.New("Failed to reject runtime, runtime is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.rejectRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"RuntimeID": runtime.ID}).Info("Rejecting runtime")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleDeleteRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to delete runtime, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete runtime, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleHTTPError(c, errors.New("Failed to delete runtime, runtime is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"RuntimeID": runtime.ID}).Info("Deleting runtime")

	server.sendEmptyHTTPReply(c, payloadType)
}
