package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleSubmitHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSubmitFunctionSpecMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		if server.handleHTTPError(c, errors.New("Failed to submit process, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to submit process spec, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.FunctionSpec == nil {
		server.handleHTTPError(c, errors.New("Failed to submit process spec, msg.FunctionSpec is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = VerifyFunctionSpec(msg.FunctionSpec)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	process := core.CreateProcess(msg.FunctionSpec)
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

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Submitting process")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleAssignProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string, originalRequest string) {
	var err error
	if server.exclusiveAssign && !server.controller.isLeader() {
		// Find out who is the leader
		leader := server.controller.getEtcdServer().CurrentCluster().Leader
		leaderHost := leader.Host
		leaderPort := leader.APIPort
		insecure := !server.tls

		log.WithFields(log.Fields{"LeaderHost": leaderHost, "LeaderPort": leaderPort}).Debug("Redirecting request to leader")
		client := client.CreateColoniesClient(leaderHost, leaderPort, insecure, true)

		jsonReplyString, err := client.SendRawMessage(string(originalRequest), insecure)
		if server.handleHTTPError(c, err, http.StatusInternalServerError) {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to send raw message to leader")
			return
		}

		c.String(http.StatusOK, jsonReplyString)
		return
	}

	msg, err := rpc.CreateAssignProcessMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to assign process, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to assign process, msg.msgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Timeout == 0 {
		server.handleHTTPError(c, errors.New("Invalid timeout value, timeout cannot be zero"), http.StatusBadRequest)
		return
	}

	executor, err := server.controller.getExecutor(recoveredID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{
		"ExecutorType": executor.Type,
		"ExecutorId":   recoveredID,
		"ColonyId":     msg.ColonyName,
		"Timeout":      msg.Timeout}).
		Debug("Waiting for processes")

	process, assignErr := server.controller.assign(recoveredID, msg.ColonyName)
	if assignErr != nil {
		if msg.Timeout > 0 {
			ctx, cancelCtx := context.WithTimeout(c.Request.Context(), time.Duration(msg.Timeout)*time.Second)
			defer cancelCtx()

			// Wait for a new process to be submitted to a ColoniesServer in the cluster
			server.controller.getEventHandler().waitForProcess(executor.Type, core.WAITING, "", ctx)
			process, assignErr = server.controller.assign(recoveredID, msg.ColonyName)
		}
	}

	if server.handleHTTPError(c, assignErr, http.StatusNotFound) {
		log.WithFields(log.Fields{"ExecutorId": recoveredID, "ColonyId": msg.ColonyName}).Debug("No process can be assigned")
		return
	}
	if process == nil {
		errmsg := "Failed to assign process, process is nil"
		log.Error(errmsg)
		server.handleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	jsonString, err = process.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID, "ExecutorId": process.AssignedExecutorID}).Debug("Assigning process")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessHistHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessHistMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get process hist, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get process history, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyName)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	processes, err := server.controller.findProcessHistory(msg.ColonyName, msg.ExecutorID, msg.Seconds, msg.State)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	jsonString, err = core.ConvertProcessArrayToJSON(processes)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyId":   msg.ColonyName,
		"ExecutorId": msg.ExecutorID,
		"Seconds":    msg.Seconds,
		"State":      msg.State}).
		Debug("Finding process history")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessesMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get processes, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get processes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyName)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName, "Count": msg.Count}).Debug("Getting processes")

	switch msg.State {
	case core.WAITING:
		processes, err := server.controller.findWaitingProcesses(msg.ColonyName, msg.ExecutorType, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.RUNNING:
		processes, err := server.controller.findRunningProcesses(msg.ColonyName, msg.ExecutorType, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.SUCCESS:
		processes, err := server.controller.findSuccessfulProcesses(msg.ColonyName, msg.ExecutorType, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.FAILED:
		processes, err := server.controller.findFailedProcesses(msg.ColonyName, msg.ExecutorType, msg.Count)
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
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get process, invalid JSON"), http.StatusBadRequest) {
			return
		}
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

	err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = process.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Getting process")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteProcessMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete process, invalid JSON"), http.StatusBadRequest) {
			return
		}
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

	err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Deleting process")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleDeleteAllProcessesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteAllProcessesMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete all processes, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete all processes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteAllProcesses(msg.ColonyName, msg.State)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName}).Debug("Deleting all processes")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleSetOutputHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSetOutputMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to set output, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to set output, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		errmsg := "Failed to set output, process is nil"
		log.Error(errmsg)
		server.handleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to close process as successful, not allowed to close process as successful"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.State != core.RUNNING {
		errmsg := "Failed to set output, process is not running"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = server.controller.setOutput(process.ID, msg.Output)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to set output")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Set output")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleCloseSuccessfulHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseSuccessfulMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to close successful, invalid JSON"), http.StatusBadRequest) {
			return
		}
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
		errmsg := "Failed to close process as successful, process is nil"
		log.Error(errmsg)
		server.handleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if process.AssignedExecutorID == "" {
		errmsg := "Failed to close process as successful, process is not assigned"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to close process as successful, not allowed to close process as successful"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = server.controller.closeSuccessful(process.ID, recoveredID, msg.Output)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to close process as successful")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Close successful")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleCloseFailedHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseFailedMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to close failed, invalid JSON"), http.StatusBadRequest) {
			return
		}
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
		errmsg := "Failed to close process as failed, process is nil"
		log.Error(errmsg)
		server.handleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedExecutorID == "" {
		errmsg := "Failed to close process as failed, process is not assigned"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to close process as failed, not allowed to close process as failed"
		log.Error(errmsg)
		err := errors.New(errmsg)
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = server.controller.closeFailed(process.ID, msg.Errors)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to close process as failed")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Close failed")

	server.sendEmptyHTTPReply(c, payloadType)
}
