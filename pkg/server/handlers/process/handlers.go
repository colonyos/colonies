package process

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/parsers"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/colonyos/colonies/pkg/backends"
	log "github.com/sirupsen/logrus"
)


func VerifyFunctionSpec(funcSpec *core.FunctionSpec) error {
	if funcSpec.Priority < constants.MIN_PRIORITY || funcSpec.Priority > constants.MAX_PRIORITY {
		msg := "Failed to submit function spec, priority outside range [" + strconv.Itoa(constants.MIN_PRIORITY) + ", " + strconv.Itoa(constants.MAX_PRIORITY) + "]"
		return errors.New(msg)
	}
	return nil
}

func resolveInitiator(
	colonyName string,
	recoveredID string,
	executorDB database.ExecutorDatabase,
	userDB database.UserDatabase) (string, error) {

	executor, err := executorDB.GetExecutorByID(recoveredID)
	if err != nil {
		return "", err
	}

	if executor != nil {
		return executor.Name, nil
	} else {
		user, err := userDB.GetUserByID(colonyName, recoveredID)
		if err != nil {
			return "", err
		}
		if user != nil {
			return user.Name, nil
		} else {
			return "", errors.New("Could not derive InitiatorName")
		}
	}
}

type Leader struct {
	Host    string
	APIPort int
}

type EtcdServer interface {
	CurrentCluster() Cluster
}

type Cluster interface {
	GetLeader() *Leader
}

type Controller interface {
	AddProcessToDB(process *core.Process) (*core.Process, error)
	AddProcess(process *core.Process) (*core.Process, error)
	GetProcess(processID string) (*core.Process, error)
	GetExecutor(executorID string) (*core.Executor, error)
	FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error)
	RemoveProcess(processID string) error
	RemoveAllProcesses(colonyName string, state int) error
	SetOutput(processID string, output []interface{}) error
	CloseSuccessful(processID string, executorID string, output []interface{}) error
	CloseFailed(processID string, errs []string) error
	Assign(executorID string, colonyName string, cpu int64, memory int64) (*AssignResult, error)
	UnassignExecutor(processID string) error
	PauseColonyAssignments(colonyName string) error
	ResumeColonyAssignments(colonyName string) error
	AreColonyAssignmentsPaused(colonyName string) (bool, error)
	GetEventHandler() *EventHandler
	IsLeader() bool
	GetEtcdServer() EtcdServer
}

type AssignResult struct {
	Process       *core.Process
	IsPaused      bool
	ResumeChannel <-chan bool
}

type EventHandler struct {
	realHandler backends.RealtimeEventHandler
}

func NewEventHandler(handler backends.RealtimeEventHandler) *EventHandler {
	return &EventHandler{realHandler: handler}
}

func (e *EventHandler) WaitForProcess(executorType string, state int, processID string, location string, ctx context.Context) (*core.Process, error) {
	if e.realHandler == nil {
		return nil, nil
	}
	return e.realHandler.WaitForProcess(executorType, state, processID, location, ctx)
}

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	Validator() security.Validator
	ExecutorDB() database.ExecutorDatabase
	UserDB() database.UserDatabase
	ProcessDB() database.ProcessDatabase
	BlueprintDB() database.BlueprintDatabase
	ProcessController() Controller
	ExclusiveAssign() bool
	TLS() bool
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
	if err := handlerRegistry.Register(rpc.SubmitFunctionSpecPayloadType, h.HandleSubmit); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterWithRawRequest(rpc.AssignProcessPayloadType, h.HandleAssignProcess); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.PauseAssignmentsPayloadType, h.HandlePauseAssignments); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ResumeAssignmentsPayloadType, h.HandleResumeAssignments); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetPauseStatusPayloadType, h.HandleGetPauseStatus); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetProcessHistPayloadType, h.HandleGetProcessHist); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetProcessesPayloadType, h.HandleGetProcesses); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetProcessPayloadType, h.HandleGetProcess); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveProcessPayloadType, h.HandleRemoveProcess); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveAllProcessesPayloadType, h.HandleRemoveAllProcesses); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.CloseSuccessfulPayloadType, h.HandleCloseSuccessful); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.CloseFailedPayloadType, h.HandleCloseFailed); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.SetOutputPayloadType, h.HandleSetOutput); err != nil {
		return err
	}
	return nil
}


func (h *Handlers) HandleSubmit(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSubmitFunctionSpecMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		if h.server.HandleHTTPError(c, errors.New("Failed to submit process, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to submit process spec, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.FunctionSpec == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to submit process spec, msg.FunctionSpec is nil"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = VerifyFunctionSpec(msg.FunctionSpec)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	process := core.CreateProcess(msg.FunctionSpec)

	initiatorName, err := resolveInitiator(msg.FunctionSpec.Conditions.ColonyName, recoveredID, h.server.ExecutorDB(), h.server.UserDB())
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	process.InitiatorID = recoveredID
	process.InitiatorName = initiatorName

	executor, err := h.server.ExecutorDB().GetExecutorByID(recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if executor != nil {
		process.InitiatorName = executor.Name
	} else {
		user, err := h.server.UserDB().GetUserByID(msg.FunctionSpec.Conditions.ColonyName, recoveredID)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		if user != nil {
			process.InitiatorName = user.Name
		} else {
			if h.server.HandleHTTPError(c, errors.New("Could not derive InitiatorName"), http.StatusBadRequest) {
				return
			}
		}
	}

	addedProcess, err := h.server.ProcessController().AddProcess(process)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedProcess == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to submit process spec, addedProcess is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedProcess.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": process.FunctionSpec.Conditions.ColonyName, "ProcessId": process.ID}).Debug("Submitting process")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// handleAssignProcess handles HTTP requests for process assignment to executors.
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
func (h *Handlers) HandleAssignProcess(c backends.Context, recoveredID string, payloadType string, jsonString string, originalRequest string) {
	var err error
	if h.server.ExclusiveAssign() && !h.server.ProcessController().IsLeader() {
		// Find out who is the leader
		leader := h.server.ProcessController().GetEtcdServer().CurrentCluster().GetLeader()
		leaderHost := leader.Host
		leaderPort := leader.APIPort
		insecure := !h.server.TLS()

		log.WithFields(log.Fields{"LeaderHost": leaderHost, "LeaderPort": leaderPort}).Debug("Redirecting request to leader")
		client := client.CreateColoniesClient(leaderHost, leaderPort, insecure, true)

		jsonReplyString, err := client.SendRawMessage(string(originalRequest), insecure)
		if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to send raw message to leader")
			return
		}

		c.String(http.StatusOK, jsonReplyString)
		return
	}

	msg, err := rpc.CreateAssignProcessMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to assign process, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to assign process, msg.msgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Timeout == 0 {
		h.server.HandleHTTPError(c, errors.New("Invalid timeout value, timeout cannot be zero"), http.StatusBadRequest)
		return
	}

	executor, err := h.server.ProcessController().GetExecutor(recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to assign process, executor not found"), http.StatusInternalServerError)
		return
	}

	var cpu int64
	if msg.AvailableCPU == "" {
		cpu = math.MaxInt64
	} else {
		cpu, err = parsers.ConvertCPUToInt(msg.AvailableCPU)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	}

	var memory int64
	if msg.AvailableMemory == "" {
		memory = math.MaxInt64
	} else {
		memory, err = parsers.ConvertMemoryToBytes(msg.AvailableMemory)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
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
		ctx, cancelCtx = context.WithTimeout(c.Request().Context(), time.Duration(msg.Timeout)*time.Second)
		defer cancelCtx()
	}

	for {
		result, assignErr := h.server.ProcessController().Assign(recoveredID, msg.ColonyName, cpu, memory)
		if assignErr != nil {
			h.server.HandleHTTPError(c, assignErr, http.StatusInternalServerError)
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
					h.server.HandleHTTPError(c, errors.New("Assignment timeout: colony assignments are paused"), http.StatusRequestTimeout)
					return
				}
			} else {
				// No timeout specified, return immediately
				h.server.HandleHTTPError(c, errors.New("No processes available: colony assignments are paused"), http.StatusServiceUnavailable)
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
			h.server.ProcessController().GetEventHandler().WaitForProcess(executor.Type, core.WAITING, "", executor.LocationName, ctx)
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
		h.server.HandleHTTPError(c, errors.New("No process available for assignment"), http.StatusNotFound)
		return
	}

	jsonString, err = process.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID, "ExecutorId": process.AssignedExecutorID}).Debug("Assigning process")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetProcessHist(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessHistMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get process hist, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get process history, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		return
	}

	processes, err := h.server.ProcessController().FindProcessHistory(msg.ColonyName, msg.ExecutorID, msg.Seconds, msg.State)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	jsonString, err = core.ConvertProcessArrayToJSON(processes)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName": msg.ColonyName,
		"ExecutorId": msg.ExecutorID,
		"Seconds":    msg.Seconds,
		"State":      msg.State}).
		Debug("Finding process history")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetProcesses(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessesMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get processes, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get processes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "Count": msg.Count}).Debug("Getting processes")

	if msg.Count > constants.MAX_COUNT {
		if h.server.HandleHTTPError(c, errors.New("Count is larger than MaxCount limit <"+strconv.Itoa(constants.MAX_COUNT)+">"), http.StatusBadRequest) {
			return
		}
	}

	if msg.Initiator != "" {
		users, err := h.server.UserDB().GetUsersByColonyName(msg.ColonyName)
		if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
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
			if h.server.HandleHTTPError(c, errors.New("User <"+msg.Initiator+"> does not exist"), http.StatusBadRequest) {
				return
			}
		}
	}

	switch msg.State {
	case core.WAITING:
		processes, err := h.server.ProcessDB().FindWaitingProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	case core.RUNNING:
		processes, err := h.server.ProcessDB().FindRunningProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	case core.SUCCESS:
		processes, err := h.server.ProcessDB().FindSuccessfulProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	case core.FAILED:
		processes, err := h.server.ProcessDB().FindFailedProcesses(msg.ColonyName, msg.ExecutorType, msg.Label, msg.Initiator, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	default:
		err := errors.New("Failed to get processes, invalid msg.State")
		h.server.HandleHTTPError(c, err, http.StatusBadRequest)
		return
	}
}

func (h *Handlers) HandleGetProcess(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get process, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get process, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := h.server.ProcessController().GetProcess(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get process, process is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = process.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Getting process")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRemoveProcess(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveProcessMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove process, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove process, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := h.server.ProcessController().GetProcess(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove process, process is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.ProcessGraphID != "" {
		err := errors.New("Failed to remove, cannot remove a process part of a workflow, delete the entire workflow instead")
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		}
		return
	}

	err = h.server.ProcessController().RemoveProcess(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Removing process")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleRemoveAllProcesses(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveAllProcessesMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove all processes, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove all processes, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.ProcessController().RemoveAllProcesses(msg.ColonyName, msg.State)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName}).Debug("Removing all processes")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleSetOutput(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSetOutputMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to set output, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to set output, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := h.server.ProcessController().GetProcess(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		errmsg := "Failed to set output, process is nil"
		log.Error(errmsg)
		h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to close process as successful, not allowed to close process as successful"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.State != core.RUNNING {
		errmsg := "Failed to set output, process is not running"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = h.server.ProcessController().SetOutput(process.ID, msg.Output)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to set output")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Set output")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleCloseSuccessful(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseSuccessfulMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to close successful, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to close process as successful, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := h.server.ProcessController().GetProcess(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		errmsg := "Failed to close process as successful, process is nil"
		log.Error(errmsg)
		h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if process.AssignedExecutorID == "" {
		errmsg := "Failed to close process as successful, process is not assigned"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to close process as successful, not allowed to close process as successful"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = h.server.ProcessController().CloseSuccessful(process.ID, recoveredID, msg.Output)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to close process as successful")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// If this was a reconciliation process, update the blueprint status from the output
	var blueprintID string
	var blueprintName string
	var colonyName string

	// Check for old-style reconciliation (embedded blueprint)
	if process.FunctionSpec.Reconciliation != nil {
		if process.FunctionSpec.Reconciliation.New != nil {
			blueprintID = process.FunctionSpec.Reconciliation.New.ID
		} else if process.FunctionSpec.Reconciliation.Old != nil {
			blueprintID = process.FunctionSpec.Reconciliation.Old.ID
		}
	}

	// Check for new cron-based reconciliation (fetch-based)
	if blueprintID == "" && process.FunctionSpec.FuncName == "reconcile" {
		if bpName, ok := process.FunctionSpec.KwArgs["blueprintName"].(string); ok {
			blueprintName = bpName
			colonyName = process.FunctionSpec.Conditions.ColonyName
		}
	}

	// Update blueprint if we have identification (either ID or name+colony)
	if (blueprintID != "" || (blueprintName != "" && colonyName != "")) && len(msg.Output) > 0 {
		// Get blueprint by ID or by name
		var blueprint *core.Blueprint
		var err error
		if blueprintID != "" {
			blueprint, err = h.server.BlueprintDB().GetBlueprintByID(blueprintID)
		} else {
			blueprint, err = h.server.BlueprintDB().GetBlueprintByName(colonyName, blueprintName)
		}

		if err == nil && blueprint != nil {
			// The first output entry should contain the status map and metadata
			if statusMap, ok := msg.Output[0].(map[string]interface{}); ok {
				// Update status if present
				if status, ok := statusMap["status"]; ok {
					if statusData, ok := status.(map[string]interface{}); ok {
						err = h.server.BlueprintDB().UpdateBlueprintStatus(blueprint.ID, statusData)
						if err != nil {
							log.WithFields(log.Fields{
								"Error":       err,
								"BlueprintID": blueprint.ID,
								"ProcessID":   process.ID,
							}).Warn("Failed to update blueprint status from reconciliation output")
						} else {
							log.WithFields(log.Fields{
								"BlueprintID": blueprint.ID,
								"ProcessID":   process.ID,
							}).Debug("Updated blueprint status from reconciliation output")
						}
					}
				}

				// NOTE: We do NOT update LastReconciliationProcess here to avoid race conditions.
				// The cron controller already updates it when creating the processgraph.
				// Updating it here can cause older processes to overwrite newer process IDs
				// when they complete out of order.
			}
		}
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Close successful")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleCloseFailed(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseFailedMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to close failed, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to close process as failed, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := h.server.ProcessController().GetProcess(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		errmsg := "Failed to close process as failed, process is nil"
		log.Error(errmsg)
		h.server.HandleHTTPError(c, errors.New(errmsg), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedExecutorID == "" {
		errmsg := "Failed to close process as failed, process is not assigned"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	if process.AssignedExecutorID != recoveredID {
		errmsg := "Failed to close process as failed, not allowed to close process as failed"
		log.Error(errmsg)
		err := errors.New(errmsg)
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	err = h.server.ProcessController().CloseFailed(process.ID, msg.Errors)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to close process as failed")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Close failed")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandlePauseAssignments(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreatePauseAssignmentsMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		h.server.HandleHTTPError(c, errors.New("Failed to pause assignments, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to pause assignments, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Check if user is colony owner
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.ProcessController().PauseColonyAssignments(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"Colony": msg.ColonyName}).Debug("Colony assignments paused successfully")
	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleResumeAssignments(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateResumeAssignmentsMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		h.server.HandleHTTPError(c, errors.New("Failed to resume assignments, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to resume assignments, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Check if user is colony owner
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.ProcessController().ResumeColonyAssignments(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"Colony": msg.ColonyName}).Debug("Colony assignments resumed successfully")
	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleGetPauseStatus(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetPauseStatusMsgFromJSON(jsonString)
	if err != nil {
		log.Warning(err)
		h.server.HandleHTTPError(c, errors.New("Failed to get pause status, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get pause status, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Check if user is colony member (less restrictive than owner for status check)
	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	isPaused, err := h.server.ProcessController().AreColonyAssignmentsPaused(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	replyMsg := rpc.CreatePauseStatusReplyMsg(msg.ColonyName, isPaused)
	jsonString, err = replyMsg.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"Colony": msg.ColonyName, "IsPaused": isPaused}).Debug("Got pause status")
	h.server.SendHTTPReply(c, rpc.PauseStatusReplyPayloadType, jsonString)
}
