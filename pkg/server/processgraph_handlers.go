package server

import (
	"errors"
	"fmt"
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
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.WorkflowSpec == nil {
		server.handleHTTPError(c, errors.New("msg.WorkflowSpec is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.WorkflowSpec.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	graph, err := core.CreateProcessGraph(msg.WorkflowSpec.ColonyID)

	// Create all processes
	processMap := make(map[string]*core.Process)
	var rootProcesses []*core.Process
	for _, processSpec := range msg.WorkflowSpec.ProcessSpecs {
		process := core.CreateProcess(processSpec)
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
				server.handleHTTPError(c, errors.New("invalid dependencies, are you depending on a process spec name that does not exits?"), http.StatusBadRequest)
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
		jsonStr, _ := process.ToJSON()
		fmt.Println(jsonStr)

		if err != nil {
			server.handleHTTPError(c, errors.New("failed to add process"), http.StatusInternalServerError)
			return
		}
	}

	// Finally, submit process graph
	_, err = server.controller.addProcessGraph(graph)
	if err != nil {
		server.handleHTTPError(c, errors.New("failed to add process graph"), http.StatusInternalServerError)
		return
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
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	processGraph, err := server.controller.getProcessGraphByID(msg.ProcessGraphID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if processGraph == nil {
		server.handleHTTPError(c, errors.New("processGraph is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, processGraph.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = processGraph.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessGraphsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessGraphsMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID, true)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyID)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

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
