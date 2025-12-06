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
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/colonyos/colonies/pkg/backends"
	log "github.com/sirupsen/logrus"
)

const MAX_LOG_COUNT = 500
const MAX_COUNT = 100
const MAX_DAYS = 30

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
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
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{
		server: server,
	}
}

// RegisterHandlers implements the HandlerRegistrar interface
func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	if err := handlerRegistry.Register(rpc.AddLogPayloadType, h.HandleAddLog); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.AddExecutorLogPayloadType, h.HandleAddExecutorLog); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetLogsPayloadType, h.HandleGetLogs); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.SearchLogsPayloadType, h.HandleSearchLogs); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddLog(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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

// HandleAddExecutorLog adds a log entry for an executor without requiring a process context.
// This is useful for executor startup logs, background operations, and diagnostics.
func (h *Handlers) HandleAddExecutorLog(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddExecutorLogMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add executor log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add executor log, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Verify the caller is a member of the colony
	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Verify the caller is the executor they claim to be
	executor, err := h.server.ExecutorDB().GetExecutorByID(recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if executor == nil {
		errmsg := "Failed to add executor log, executor not found"
		log.Error(errmsg)
		h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusForbidden)
		return
	}

	if executor.Name != msg.ExecutorName {
		errmsg := "Failed to add executor log, executor name mismatch"
		log.Error(errmsg)
		h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusForbidden)
		return
	}

	// Add the log with a special process ID to indicate executor-level log
	// Using empty processID which will be stored and retrievable via executor name
	err = h.server.LogDB().AddLog("", msg.ColonyName, executor.Name, time.Now().UTC().UnixNano(), msg.Message)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to add executor log")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ExecutorName": executor.Name}).Debug("Adding executor log")

	h.server.SendEmptyHTTPReply(c, rpc.AddExecutorLogPayloadType)
}

func (h *Handlers) HandleGetLogs(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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
		if msg.Latest {
			// Get latest logs (most recent count logs)
			logs, err = h.server.LogDB().GetLogsByExecutorLatest(msg.ExecutorName, msg.Count)
		} else if msg.Since > 0 {
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
		if msg.Latest {
			// Get latest logs (most recent count logs)
			logs, err = h.server.LogDB().GetLogsByProcessIDLatest(msg.ProcessID, msg.Count)
		} else if msg.Since > 0 {
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

func (h *Handlers) HandleSearchLogs(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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