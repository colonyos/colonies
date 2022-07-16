package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleSubmitProcessSpecHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSubmitProcessSpecMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to submit process spec, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to submit process spec, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.ProcessSpec == nil {
		server.handleHTTPError(c, errors.New("Failed to submit process spec, msg.ProcessSpec is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ProcessSpec.Conditions.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	process := core.CreateProcess(msg.ProcessSpec)
	addedProcess, err := server.controller.addProcess(process)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedProcess == nil {
		server.handleHTTPError(c, errors.New("Failed to submit process spec, addedProcess is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedProcess.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Submitting process")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleAssignProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAssignProcessMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to assign process, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to assign process, msg.msgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	process, err := server.controller.assignRuntime(recoveredID, msg.ColonyID, msg.Latest)
	if server.handleHTTPError(c, err, http.StatusNoContent) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to assign process, process is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = process.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessID": process.ID, "RuntimeID": process.AssignedRuntimeID}).Info("Assigning process")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessHistHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessHistMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get process history, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get process history, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID, true)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyID)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	processes, err := server.controller.findProcessHistory(msg.ColonyID, msg.RuntimeID, msg.Seconds, msg.State)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	jsonString, err = core.ConvertProcessArrayToJSON(processes)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyID":  msg.ColonyID,
		"RuntimeID": msg.RuntimeID,
		"Seconds":   msg.Seconds,
		"State":     msg.State}).
		Info("Finding process history")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessesMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get processes, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get processes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID, true)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyID)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	log.WithFields(log.Fields{"ColonyID": msg.ColonyID, "Count": msg.Count}).Info("Getting processes")

	switch msg.State {
	case core.WAITING:
		processes, err := server.controller.findWaitingProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.RUNNING:
		processes, err := server.controller.findRunningProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.SUCCESS:
		processes, err := server.controller.findSuccessfulProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.FAILED:
		processes, err := server.controller.findFailedProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	default:
		err := errors.New("Failed to get processes, invalid msg.State")
		server.handleHTTPError(c, err, http.StatusBadRequest)
		return
	}
}

func (server *ColoniesServer) handleGetProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get process, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get process, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to get process, process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = process.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Getting process")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteProcessMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to delete process, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete process, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to delete process, process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Deleting process")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleDeleteAllProcessesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteAllProcessesMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to delete all processes, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete all processes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteAllProcesses(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyID": msg.ColonyID}).Info("Deleting all processes")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleCloseSuccessfulHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseSuccessfulMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to close process as successful, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to close process as successful, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to close process as successful, process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("Failed to close process as successful, not allowed to close process")
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = server.controller.closeSuccessful(process.ID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Close successful")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleCloseFailedHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseFailedMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to close process as failed, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to close process as failed, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to close process as failed, process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("Failed to close process as failed, not allowed to close process")
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = server.controller.closeFailed(process.ID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Close failed")

	server.sendEmptyHTTPReply(c, payloadType)
}
