package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddExecutorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddExecutorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Executor == nil {
		server.handleHTTPError(c, errors.New("Failed to add executor, executor is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.Executor.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	addedExecutor, err := server.controller.addExecutor(msg.Executor, server.allowExecutorReregister)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if addedExecutor == nil {
		server.handleHTTPError(c, errors.New("Failed to add executor, addedExecutor is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedExecutor.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyId":     msg.Executor.ColonyName,
		"ExecutorName": addedExecutor.Name,
		"ExecutorType": addedExecutor.Type,
		"ExecutorId":   addedExecutor.ID}).
		Debug("Adding executor")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetExecutorsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetExecutorsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get executors, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get executors, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, false)
	if err != nil {
		return
	}

	executors, err := server.controller.getExecutorByColonyName(msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertExecutorArrayToJSON(executors)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName}).Debug("Getting executors")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetExecutorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetExecutorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := server.db.GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to get executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireMembership(recoveredID, executor.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = executor.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Getting executor")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleApproveExecutorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateApproveExecutorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to approve executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to approve executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := server.db.GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to approve executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, executor.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.db.ApproveExecutor(executor)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Approving executor")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleRejectExecutorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRejectExecutorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to reject executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to reject executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := server.db.GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to reject executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, executor.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.db.RejectExecutor(executor)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Rejecting executor")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleRemoveExecutorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveExecutorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to remove executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to remove executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := server.db.GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to remove executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, executor.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.db.RemoveExecutorByName(msg.ColonyName, msg.ExecutorName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Removing executor")

	server.sendEmptyHTTPReply(c, payloadType)
}
