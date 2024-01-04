package server

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddLogHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddLogMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add log, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		errmsg := "Failed to add log, process is nil"
		log.Error(errmsg)
		server.handleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	executor, err := server.db.GetExecutorByID(recoveredID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if process.State != core.RUNNING {
		errmsg := "Failed to set output, process is not running"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to add log, not allowed to add log, only the assigned Executor may att logs"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	colonyName := process.FunctionSpec.Conditions.ColonyName
	err = server.db.AddLog(process.ID, colonyName, executor.Name, time.Now().UTC().UnixNano(), msg.Message)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to add log")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Adding log")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleGetLogsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetLogsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get logs, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get logs, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ExecutorName != "" {
		executor, err := server.db.GetExecutorByName(msg.ColonyName, msg.ExecutorName)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}

		if executor == nil {
			errmsg := "Failed to get logs, executor does not exist"
			log.Error(errmsg)
			server.handleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
			return
		}

		err = server.validator.RequireMembership(recoveredID, executor.ColonyName, true)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			log.Error(err)
			return
		}
	} else {
		process, err := server.controller.getProcess(msg.ProcessID)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		if process == nil {
			errmsg := "Failed to get logs, process does not exist"
			log.Error(errmsg)
			server.handleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
			return
		}

		err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			log.Error(err)
			return
		}
	}

	if msg.Count > MAX_LOG_COUNT {
		if server.handleHTTPError(c, errors.New("Count exceeds max log count ("+strconv.Itoa(MAX_LOG_COUNT)+")"), http.StatusForbidden) {
			return
		}
	}

	var logs []core.Log
	if msg.ExecutorName != "" {
		if msg.Since > 0 {
			logs, err = server.db.GetLogsByExecutorSince(msg.ExecutorName, msg.Count, msg.Since)
		} else {
			logs, err = server.db.GetLogsByExecutor(msg.ExecutorName, msg.Count)
		}
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			log.WithFields(log.Fields{"Error": err, "ColonyName": msg.ColonyName}).Debug("Failed to get logs for executor")
			server.handleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
	} else {
		if msg.Since > 0 {
			logs, err = server.db.GetLogsByProcessIDSince(msg.ProcessID, msg.Count, msg.Since)
		} else {
			logs, err = server.db.GetLogsByProcessID(msg.ProcessID, msg.Count)
		}
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			log.WithFields(log.Fields{"Error": err, "ColonyName": msg.ColonyName}).Debug("Failed to get logs for process")
			server.handleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
	}

	jsonStr, err := core.ConvertLogArrayToJSON(logs)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to parse log")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.Debug("Getting logs")
	server.sendHTTPReply(c, payloadType, jsonStr)
}
