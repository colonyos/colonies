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

	err = server.validator.RequireMembership(recoveredID, msg.Function.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executor, err := server.db.GetExecutorByName(msg.Function.ColonyName, msg.Function.ExecutorName)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to add function, Executor with name <"+msg.Function.ExecutorName+"> does not exist"), http.StatusBadRequest)
		return
	}

	if executor.ID != recoveredID {
		if server.handleHTTPError(c, errors.New("Not allowed to add a function to another executor"), http.StatusForbidden) {
			return
		}
	}

	functions, err := server.controller.getFunctionsByExecutorName(msg.Function.ColonyName, msg.Function.ExecutorName)
	for _, function := range functions {
		if function.FuncName == msg.Function.FuncName {
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

	log.WithFields(log.Fields{"FunctionId": addedFunction.FunctionID, "ExecutorName": addedFunction.ExecutorName, "ColonyName": addedFunction.ColonyName, "FuncName": addedFunction.FuncName}).Debug("Adding function")

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

	var functions []*core.Function

	if msg.ExecutorName == "" {
		err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
		functions, err = server.controller.getFunctionsByColonyName(msg.ColonyName)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	} else {
		targetExecutor, err := server.db.GetExecutorByName(msg.ColonyName, msg.ExecutorName)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
		if targetExecutor == nil {
			if server.handleHTTPError(c, errors.New("Executor not found"), http.StatusForbidden) {
				return
			}
		}
		err = server.validator.RequireMembership(recoveredID, targetExecutor.ColonyName, true)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
		functions, err = server.controller.getFunctionsByExecutorName(msg.ColonyName, msg.ExecutorName)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	jsonString, err = core.ConvertFunctionArrayToJSON(functions)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleRemoveFunctionHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveFunctionMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to remove function, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to remove function, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.FunctionID == "" {
		server.handleHTTPError(c, errors.New("Failed to remove function, msg.FunctionID is empty"), http.StatusBadRequest)
		return
	}

	function, err := server.controller.getFunctionByID(msg.FunctionID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executor, err := server.db.GetExecutorByName(function.ColonyName, function.ExecutorName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to remove function, Executor does not exist"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, executor.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if executor.ID != recoveredID {
		if server.handleHTTPError(c, errors.New("Not allowed to add a function to another executor"), http.StatusForbidden) {
			return
		}
	}

	err = server.controller.removeFunction(msg.FunctionID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{"FunctionId": msg.FunctionID}).Debug("Removing function")

	server.sendEmptyHTTPReply(c, payloadType)
}
