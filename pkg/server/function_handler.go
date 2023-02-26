package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddFunctionHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddFunctionMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add function, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add function, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Function == nil {
		server.handleHTTPError(c, errors.New("Failed to add function, msg.Function is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, msg.Function.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Function.ExecutorID != recoveredID {
		if server.handleHTTPError(c, errors.New("Not allowed to add a function to another executor"), http.StatusForbidden) {
			return
		}
	}

	functions, err := server.controller.getFunctionByExecutorID(msg.Function.ExecutorID)
	for _, function := range functions {
		if function.Name == msg.Function.Name {
			if server.handleHTTPError(c, errors.New("Function already exists"), http.StatusForbidden) {
				return
			}
		}
	}

	msg.Function.FunctionID = core.GenerateRandomID()
	addedFunction, err := server.controller.addFunction(msg.Function)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = addedFunction.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"FunctionId": addedFunction.FunctionID, "ExecutorId": addedFunction.ExecutorID, "ColonyID": addedFunction.ColonyID, "Name": addedFunction.Name}).Debug("Adding function")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetFunctionsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFunctionsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get function, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get function, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ExecutorID == "" {
		server.handleHTTPError(c, errors.New("Failed to get functions, msg.ExecutorID is empty"), http.StatusBadRequest)
		return
	}

	executor, err := server.controller.db.GetExecutorByID(msg.ExecutorID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, executor.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	functions, err := server.controller.getFunctionByExecutorID(msg.ExecutorID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = core.ConvertFunctionArrayToJSON(functions)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteFunctionHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteFunctionMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete function, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete function, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.FunctionID == "" {
		server.handleHTTPError(c, errors.New("Failed to delete function, msg.FunctionID is empty"), http.StatusBadRequest)
		return
	}

	function, err := server.controller.getFunctionByID(msg.FunctionID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executor, err := server.controller.db.GetExecutorByID(function.ExecutorID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, executor.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if function.ExecutorID != recoveredID {
		if server.handleHTTPError(c, errors.New("Not allowed to add a function to another executor"), http.StatusForbidden) {
			return
		}
	}

	err = server.controller.deleteFunction(msg.FunctionID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{"FunctionId": msg.FunctionID}).Debug("Deleting function")

	server.sendEmptyHTTPReply(c, payloadType)
}
