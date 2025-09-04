package server

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/parsers"
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

	initiatorName, err := resolveInitiator(msg.FunctionSpec.Conditions.ColonyName, recoveredID, server.executorDB, server.userDB)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	process.InitiatorID = recoveredID
	process.InitiatorName = initiatorName

	executor, err := server.executorDB.GetExecutorByID(recoveredID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if executor != nil {
		process.InitiatorName = executor.Name
	} else {
		user, err := server.userDB.GetUserByID(msg.FunctionSpec.Conditions.ColonyName, recoveredID)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		if user != nil {
			process.InitiatorName = user.Name
		} else {
			if server.handleHTTPError(c, errors.New("Could not derive InitiatorName"), http.StatusBadRequest) {
				return
			}
		}
	}

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

	log.WithFields(log.Fields{"ColonyName": process.FunctionSpec.Conditions.ColonyName, "ProcessId": process.ID}).Debug("Submitting process")

	server.sendHTTPReply(c, payloadType, jsonString)
}

// handleAssignProcessHTTPRequest handles HTTP requests for process assignment to executors.
//
// This function implements leader-based exclusive assignment with retry loop and timeout support.
// It handles four main scenarios:
// 1. Non-leader node - redirects request to cluster leader for exclusive assignment
// 2. Colony assignments are paused - waits for resume signal or timeout
// 3. Process found - assigns and returns the process immediately  
// 4. No process available - waits for new processes to be submitted
//
// Flow:
//   - If exclusiveAssign is enabled and this node is not leader: redirect to leader
//   - Validates request parameters and executor permissions
//   - Enters retry loop that continues until timeout is reached
//   - Each iteration calls controller.assign() to attempt process assignment
//   - If assignments are paused: waits on resume channel or timeout
//   - If no process found: waits for new process events via waitForProcess()
//   - Returns assigned process or appropriate error (timeout, forbidden, etc.)
//
// Parameters:
//   - c: Gin HTTP context for the request
//   - recoveredID: Executor ID recovered from authentication 
//   - payloadType: Expected message type for validation
//   - jsonString: Request body containing assignment parameters
//   - originalRequest: Raw request for leader redirection in cluster mode
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

	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to assign process, executor not found"), http.StatusInternalServerError)
		return
	}

	var cpu int64
	if msg.AvailableCPU == "" {
		cpu = math.MaxInt64
	} else {
		cpu, err = parsers.ConvertCPUToInt(msg.AvailableCPU)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	}

	var memory int64
	if msg.AvailableMemory == "" {
		memory = math.MaxInt64
	} else {
		memory, err = parsers.ConvertMemoryToBytes(msg.AvailableMemory)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	}

	log.WithFields(log.Fields{
		"ExecutorType": executor.Type,
		"ExecutorId":   recoveredID,
		"ColonyName":   msg.ColonyName,
		"AvailableCPU": msg.AvailableCPU,
		"AvailableMem": msg.AvailableMemory,
		"CPU":          cpu,
		"Memory":       memory,
		"Timeout":      msg.Timeout}).
		Debug("Waiting for processes")

	var process *core.Process
	var ctx context.Context
	var cancelCtx context.CancelFunc

	if msg.Timeout > 0 {
		ctx, cancelCtx = context.WithTimeout(c.Request.Context(), time.Duration(msg.Timeout)*time.Second)
		defer cancelCtx()
	}

	for {
		result, assignErr := server.controller.assign(recoveredID, msg.ColonyName, cpu, memory)
		if assignErr != nil {
			server.handleHTTPError(c, assignErr, http.StatusInternalServerError)
			return
		}

		if result.IsPaused {
			// Assignments are paused
			if msg.Timeout > 0 {
				// Wait for resume signal or timeout
				select {
				case <-result.ResumeChannel:
					// Assignments resumed, continue to next retry
					continue
				case <-ctx.Done():
					// Timeout
					server.handleHTTPError(c, errors.New("Assignment timeout: colony assignments are paused"), http.StatusRequestTimeout)
					return
				}
			} else {
				// No timeout specified, return immediately
				server.handleHTTPError(c, errors.New("No processes available: colony assignments are paused"), http.StatusServiceUnavailable)
				return
			}
		}

		if result.Process != nil {
			// Got a process, we're done
			process = result.Process
			break
		}

		// No process available, wait for new processes if timeout is specified
		if msg.Timeout > 0 {
			// Wait for a new process to be submitted to a ColoniesServer in the cluster
			server.controller.getEventHandler().waitForProcess(executor.Type, core.WAITING, "", ctx)
			// Check if we timed out during the wait
			select {
			case <-ctx.Done():
				// Timeout occurred while waiting
				break
			default:
				// Continue to next retry to try assignment again
				continue
			}
		}
		
		// No timeout specified or timeout occurred, exit loop
		break
	}

	// Check if we still don't have a process after all attempts
	if process == nil {
		log.WithFields(log.Fields{"ExecutorId": recoveredID, "ColonyName": msg.ColonyName}).Debug("No process can be assigned")
		server.handleHTTPError(c, errors.New("No process available for assignment"), http.StatusNotFound)
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
		return
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
		"ColonyName": msg.ColonyName,
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
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "Count": msg.Count}).Debug("Getting processes")

	if msg.Count > MAX_COUNT {
		if server.handleHTTPError(c, errors.New("Count is larger than MaxCount limit <"+strconv.Itoa(MAX_COUNT)+">"), http.StatusBadRequest) {
			return
		}
	}

	if msg.Initiator != "" {
		users, err := server.userDB.GetUsersByColonyName(msg.ColonyName)
		if server.handleHTTPError(c, err, http.StatusInternalServerError) {
			return
		}

		found := false
		for _, user := range users {
			if user.Name == msg.Initiator {
				found = true
				break
			}
		}
		if !found {
			if server.handleHTTPError(c, errors.New("User <"+msg.Initiator+"> does not exist"), http.StatusBadRequest) {
				return
			}
		}
	}

	switch msg.State {
	case core.WAITING:
		processes, err := server.processDB.FindWaitingProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.RUNNING:
		processes, err := server.processDB.FindRunningProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.SUCCESS:
		processes, err := server.processDB.FindSuccessfulProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.FAILED:
		processes, err := server.processDB.FindFailedProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
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

func (server *ColoniesServer) handleRemoveProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveProcessMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to remove process, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to remove process, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to remove process, process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.ProcessGraphID != "" {
		err := errors.New("Failed to remove, cannot remove a process part of a workflow, delete the entire workflow instead")
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
		}
		return
	}

	err = server.controller.removeProcess(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Removing process")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleRemoveAllProcessesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveAllProcessesMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to remove all processes, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to remove all processes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.removeAllProcesses(msg.ColonyName, msg.State)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName}).Debug("Removing all processes")

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

func (server *ColoniesServer) handlePauseAssignmentsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreatePauseAssignmentsMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		server.handleHTTPError(c, errors.New("Failed to pause assignments, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to pause assignments, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Check if user is colony owner
	err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.pauseColonyAssignments(msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"Colony": msg.ColonyName}).Debug("Colony assignments paused successfully")
	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleResumeAssignmentsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateResumeAssignmentsMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		server.handleHTTPError(c, errors.New("Failed to resume assignments, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to resume assignments, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Check if user is colony owner
	err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.resumeColonyAssignments(msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"Colony": msg.ColonyName}).Debug("Colony assignments resumed successfully")
	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleGetPauseStatusHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetPauseStatusMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		server.handleHTTPError(c, errors.New("Failed to get pause status, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get pause status, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Check if user is colony member (less restrictive than owner for status check)
	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, false)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	isPaused, err := server.controller.areColonyAssignmentsPaused(msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	replyMsg := rpc.CreatePauseStatusReplyMsg(msg.ColonyName, isPaused)
	jsonString, err = replyMsg.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"Colony": msg.ColonyName, "IsPaused": isPaused}).Debug("Got pause status")
	server.sendHTTPReply(c, rpc.PauseStatusReplyPayloadType, jsonString)
}
