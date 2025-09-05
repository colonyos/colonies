package log

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const MAX_LOG_COUNT = 500
const MAX_COUNT = 100
const MAX_DAYS = 30

type ColoniesServer interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	Validator() security.Validator
	ExecutorDB() database.ExecutorDatabase
	ProcessDB() database.ProcessDatabase
	LogDB() database.LogDatabase
	LogProcessController() Controller
}

type Controller interface {
	GetProcess(processID string) (*core.Process, error)
}

type Handlers struct {
	server ColoniesServer
}

func NewHandlers(server ColoniesServer) *Handlers {
	return &Handlers{
		server: server,
	}
}

func (h *Handlers) HandleAddLog(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddLogMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add log, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := h.server.LogProcessController().GetProcess(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		errmsg := "Failed to add log, process is nil"
		log.Error(errmsg)
		h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByID(recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if process.State != core.RUNNING {
		errmsg := "Failed to set output, process is not running"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to add log, not allowed to add log, only the assigned Executor may att logs"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	colonyName := process.FunctionSpec.Conditions.ColonyName
	err = h.server.LogDB().AddLog(process.ID, colonyName, executor.Name, time.Now().UTC().UnixNano(), msg.Message)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to add log")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Adding log")

	h.server.SendEmptyHTTPReply(c, rpc.AddLogPayloadType)
}

func (h *Handlers) HandleGetLogs(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetLogsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get logs, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get logs, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ExecutorName != "" {
		executor, err := h.server.ExecutorDB().GetExecutorByName(msg.ColonyName, msg.ExecutorName)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}

		if executor == nil {
			errmsg := "Failed to get logs, executor does not exist"
			log.Error(errmsg)
			h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
			return
		}

		err = h.server.Validator().RequireMembership(recoveredID, executor.ColonyName, true)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			log.Error(err)
			return
		}
	} else {
		process, err := h.server.LogProcessController().GetProcess(msg.ProcessID)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		if process == nil {
			errmsg := "Failed to get logs, process does not exist"
			log.Error(errmsg)
			h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
			return
		}

		err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			log.Error(err)
			return
		}
	}

	if msg.Count > MAX_LOG_COUNT {
		if h.server.HandleHTTPError(c, errors.New("Count exceeds max log count ("+strconv.Itoa(MAX_LOG_COUNT)+")"), http.StatusForbidden) {
			return
		}
	}

	var logs []*core.Log
	if msg.ExecutorName != "" {
		if msg.Since > 0 {
			logs, err = h.server.LogDB().GetLogsByExecutorSince(msg.ExecutorName, msg.Count, msg.Since)
		} else {
			logs, err = h.server.LogDB().GetLogsByExecutor(msg.ExecutorName, msg.Count)
		}
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			log.WithFields(log.Fields{"Error": err, "ColonyName": msg.ColonyName}).Debug("Failed to get logs for executor")
			h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
	} else {
		if msg.Since > 0 {
			logs, err = h.server.LogDB().GetLogsByProcessIDSince(msg.ProcessID, msg.Count, msg.Since)
		} else {
			logs, err = h.server.LogDB().GetLogsByProcessID(msg.ProcessID, msg.Count)
		}
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			log.WithFields(log.Fields{"Error": err, "ColonyName": msg.ColonyName}).Debug("Failed to get logs for process")
			h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
	}

	jsonStr, err := core.ConvertLogArrayToJSON(logs)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to parse log")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.Debug("Getting logs")
	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleSearchLogs(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSearchLogsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to search logs, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get logs, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if msg.Count > MAX_COUNT {
		if h.server.HandleHTTPError(c, errors.New("Count exceeds max log count ("+strconv.Itoa(MAX_COUNT)+")"), http.StatusBadRequest) {
			return
		}
	}

	if msg.Days > MAX_DAYS {
		if h.server.HandleHTTPError(c, errors.New("Count exceeds max day count ("+strconv.Itoa(MAX_DAYS)+")"), http.StatusBadRequest) {
			return
		}
	}

	logs, err := h.server.LogDB().SearchLogs(msg.ColonyName, msg.Text, msg.Days, msg.Count)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err, "ColonyName": msg.ColonyName}).Debug("Failed to search logs")
		return
	}

	jsonStr, err := core.ConvertLogArrayToJSON(logs)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to parse log")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.Debug("Getting logs")
	h.server.SendHTTPReply(c, payloadType, jsonStr)
}