package executor

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/colonyos/colonies/pkg/backends"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	AddExecutor(executor *core.Executor, allowReregister bool) (*core.Executor, error)
	GetExecutor(executorID string) (*core.Executor, error)
	GetExecutorByColonyName(colonyName string) ([]*core.Executor, error)
}

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	Validator() security.Validator
	ExecutorDB() database.ExecutorDatabase
	ExecutorController() Controller
	AllowExecutorReregister() bool
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
	if err := handlerRegistry.Register(rpc.AddExecutorPayloadType, h.HandleAddExecutor); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetExecutorsPayloadType, h.HandleGetExecutors); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetExecutorPayloadType, h.HandleGetExecutor); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ApproveExecutorPayloadType, h.HandleApproveExecutor); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RejectExecutorPayloadType, h.HandleRejectExecutor); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveExecutorPayloadType, h.HandleRemoveExecutor); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ReportAllocationsPayloadType, h.HandleReportAllocations); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddExecutor(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddExecutorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add executor, executor is nil"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Executor.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	addedExecutor, err := h.server.ExecutorController().AddExecutor(msg.Executor, h.server.AllowExecutorReregister())
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if addedExecutor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add executor, addedExecutor is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedExecutor.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyId":     msg.Executor.ColonyName,
		"ExecutorName": addedExecutor.Name,
		"ExecutorType": addedExecutor.Type,
		"ExecutorId":   addedExecutor.ID}).
		Debug("Adding executor")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetExecutors(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetExecutorsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get executors, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get executors, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executors, err := h.server.ExecutorController().GetExecutorByColonyName(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertExecutorArrayToJSON(executors)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName}).Debug("Getting executors")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetExecutor(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetExecutorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, executor.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = executor.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Getting executor")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleApproveExecutor(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateApproveExecutorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to approve executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to approve executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to approve executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireColonyOwner(recoveredID, executor.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.ExecutorDB().ApproveExecutor(executor)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Approving executor")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleRejectExecutor(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRejectExecutorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to reject executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to reject executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to reject executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireColonyOwner(recoveredID, executor.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.ExecutorDB().RejectExecutor(executor)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Rejecting executor")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleRemoveExecutor(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveExecutorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove executor, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove executor, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove executor, executor is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireColonyOwner(recoveredID, executor.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.ExecutorDB().RemoveExecutorByName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ExecutorId": executor.ID}).Debug("Removing executor")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleReportAllocations(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateReportAllocationsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to report allocation, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to report allocation, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}
	if executor == nil {
		if h.server.HandleHTTPError(c, errors.New("Executor with name <"+msg.ExecutorName+"> does not exist"), http.StatusBadRequest) {
			return
		}
	}

	if executor.ID != recoveredID {
		if h.server.HandleHTTPError(c, errors.New("Only an executor can report allocations to itself"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.ExecutorDB().SetAllocations(msg.ColonyName, executor.Name, msg.Allocations)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ExecutorName": executor.Name, "ColonyName": msg.ColonyName}).Debug("Reporting allocations")

	h.server.SendEmptyHTTPReply(c, payloadType)
}
