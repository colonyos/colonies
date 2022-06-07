package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleSubmitWorkflowHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSubmitWorkflowSpecMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to submit workflow, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to submit workflow, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.WorkflowSpec == nil {
		server.handleHTTPError(c, errors.New("Failed to submit workflow, msg.WorkflowSpec is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.WorkflowSpec.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	graph, err := core.CreateProcessGraph(msg.WorkflowSpec.ColonyID)
	_, err = server.controller.addProcessGraph(graph)
	if err != nil {
		server.handleHTTPError(c, errors.New("Failed to submit workflow, failed to add process graph"), http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"WorkflowID": graph.ID}).Info("Submitting workflow")

	// Create all processes
	processMap := make(map[string]*core.Process)
	var rootProcesses []*core.Process
	for _, processSpec := range msg.WorkflowSpec.ProcessSpecs {
		if processSpec.MaxExecTime == 0 {
			log.WithFields(log.Fields{"Name": processSpec.Name}).Warning("MaxExecTime was set to 0, resetting to -1")
			processSpec.MaxExecTime = -1
		}
		process := core.CreateProcess(processSpec)
		log.WithFields(log.Fields{"ProcessID": process.ID, "MaxExecTime": process.ProcessSpec.MaxExecTime, "MaxRetries": process.ProcessSpec.MaxRetries}).Info("Creating new process")
		if len(processSpec.Conditions.Dependencies) == 0 {
			// The process is a root process, let it start immediately
			process.WaitForParents = false
			rootProcesses = append(rootProcesses, process)
			graph.AddRoot(process.ID)
		} else {
			// The process has to wait for its parents
			process.WaitForParents = true
		}
		process.ProcessGraphID = graph.ID
		process.ProcessSpec.Conditions.ColonyID = msg.WorkflowSpec.ColonyID
		processMap[process.ProcessSpec.Name] = process
	}

	// Create dependencies
	for _, process := range processMap {
		for _, dependsOn := range process.ProcessSpec.Conditions.Dependencies {
			parentProcess := processMap[dependsOn]
			if parentProcess == nil {
				server.handleHTTPError(c, errors.New("Failed to submit workflow, invalid dependencies, are you depending on a process spec name that does not exits?"), http.StatusBadRequest)
				return
			}
			process.AddParent(parentProcess.ID)
			parentProcess.AddChild(process.ID)
		}
	}

	// Now, start all processes
	for _, process := range processMap {
		_, err := server.controller.addProcess(process)
		log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Submitting process")

		if err != nil {
			server.handleHTTPError(c, errors.New("Failed to submit workflow, failed to add process"), http.StatusInternalServerError)
			return
		}
	}

	jsonString, err = graph.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessGraphHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessGraphMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get process graph, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get process graph, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	graph, err := server.controller.getProcessGraphByID(msg.ProcessGraphID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if graph == nil {
		server.handleHTTPError(c, errors.New("Failed to get processgraph, graph is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, graph.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = graph.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"WorkflowID": graph.ID}).Info("Getting process graph")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessGraphsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessGraphsMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get process graphs, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get process grpahs, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID, true)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyID)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	log.WithFields(log.Fields{"ColonyID": msg.ColonyID}).Info("Getting process graphs")

	switch msg.State {
	case core.WAITING:
		graphs, err := server.controller.findWaitingProcessGraphs(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.RUNNING:
		graphs, err := server.controller.findRunningProcessGraphs(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.SUCCESS:
		graphs, err := server.controller.findSuccessfulProcessGraphs(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.FAILED:
		graphs, err := server.controller.findFailedProcessGraphs(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	default:
		err := errors.New("invalid msg.State")
		server.handleHTTPError(c, err, http.StatusBadRequest)
		return
	}
}
