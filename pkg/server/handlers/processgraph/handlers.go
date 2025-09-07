package processgraph

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, initiatorID string) (*core.ProcessGraph, error)
	GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	RemoveProcessGraph(processGraphID string) error
	RemoveAllProcessGraphs(colonyName string, state int) error
	AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, initiatorID string, insert bool) (*core.Process, error)
}

type Validator interface {
	RequireMembership(recoveredID string, colonyName string, executorMayJoin bool) error
	RequireColonyOwner(recoveredID string, colonyName string) error
}

type Server interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	Validator() Validator
	Controller() Controller
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
	if err := handlerRegistry.RegisterGin(rpc.SubmitWorkflowSpecPayloadType, h.HandleSubmitWorkflow); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetProcessGraphPayloadType, h.HandleGetProcessGraph); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetProcessGraphsPayloadType, h.HandleGetProcessGraphs); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RemoveProcessGraphPayloadType, h.HandleRemoveProcessGraph); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RemoveAllProcessGraphsPayloadType, h.HandleRemoveAllProcessGraphs); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.AddChildPayloadType, h.HandleAddChild); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleSubmitWorkflow(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSubmitWorkflowSpecMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to submit workkflow, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to submit workflow, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.WorkflowSpec == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to submit workflow, msg.WorkflowSpec is nil"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.WorkflowSpec.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	processGraph, err := h.server.Controller().SubmitWorkflowSpec(msg.WorkflowSpec, recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = processGraph.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetProcessGraph(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessGraphMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get processgraph, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get processgraph, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	graph, err := h.server.Controller().GetProcessGraphByID(msg.ProcessGraphID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if graph == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get processgraph, graph is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, graph.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = graph.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessGraphId": graph.ID}).Debug("Getting processgraph")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetProcessGraphs(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessGraphsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get processgraphs, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get processgraphs, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName}).Debug("Getting processgraphs")

	switch msg.State {
	case core.WAITING:
		graphs, err := h.server.Controller().FindWaitingProcessGraphs(msg.ColonyName, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	case core.RUNNING:
		graphs, err := h.server.Controller().FindRunningProcessGraphs(msg.ColonyName, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	case core.SUCCESS:
		graphs, err := h.server.Controller().FindSuccessfulProcessGraphs(msg.ColonyName, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	case core.FAILED:
		graphs, err := h.server.Controller().FindFailedProcessGraphs(msg.ColonyName, msg.Count)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		h.server.SendHTTPReply(c, payloadType, jsonString)
	default:
		err := errors.New("invalid msg.State")
		h.server.HandleHTTPError(c, err, http.StatusBadRequest)
		return
	}
}

func (h *Handlers) HandleRemoveProcessGraph(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveProcessGraphMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove processgraph, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove processgraph, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	graph, err := h.server.Controller().GetProcessGraphByID(msg.ProcessGraphID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if graph == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove processgraph, graph is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, graph.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.Controller().RemoveProcessGraph(msg.ProcessGraphID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ProcessGraphId": graph.ID}).Debug("Removing processgraph")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleRemoveAllProcessGraphs(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveAllProcessGraphsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove all processgraphs, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove all processgraphs, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.Controller().RemoveAllProcessGraphs(msg.ColonyName, msg.State)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName}).Debug("Removing all processgraphs")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleAddChild(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddChildMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add child to processgraph, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add child to processgraph, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.FunctionSpec == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add child to processgraph, msg.FunctionSpec is nil"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	process := core.CreateProcess(msg.FunctionSpec)
	addedProcess, err := h.server.Controller().AddChild(msg.ProcessGraphID, msg.ParentProcessID, msg.ChildProcessID, process, recoveredID, msg.Insert)
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

	log.WithFields(log.Fields{
		"ProcessGraphId":  msg.ProcessGraphID,
		"ParentProcessID": msg.ParentProcessID,
		"ChildProcessID":  msg.ChildProcessID,
		"ProcessID":       process.ID}).
		Debug("Adding child process")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}