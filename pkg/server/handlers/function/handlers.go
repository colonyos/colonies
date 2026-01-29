package function

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

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	Validator() security.Validator
	FunctionDB() database.FunctionDatabase
	ExecutorDB() database.ExecutorDatabase
	UserDB() database.UserDatabase
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{server: server}
}

// RegisterHandlers implements the HandlerRegistrar interface
func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	if err := handlerRegistry.Register(rpc.AddFunctionPayloadType, h.HandleAddFunction); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetFunctionsPayloadType, h.HandleGetFunctions); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveFunctionPayloadType, h.HandleRemoveFunction); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddFunction(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddFunctionMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add function, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add function, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Function == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add function, msg.Function is nil"), http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if msg.Function.FunctionID == "" {
		msg.Function.FunctionID = core.GenerateRandomID()
		log.WithField("GeneratedID", msg.Function.FunctionID).Debug("Generated function ID")
	} else {
		log.WithField("ExistingID", msg.Function.FunctionID).Debug("Using existing function ID")
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.Function.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByName(msg.Function.ColonyName, msg.Function.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add function, Executor with name <"+msg.Function.ExecutorName+"> does not exist"), http.StatusBadRequest)
		return
	}

	if executor.ID != recoveredID {
		if h.server.HandleHTTPError(c, errors.New("Not allowed to add a function to another executor"), http.StatusForbidden) {
			return
		}
	}

	// Auto-populate LocationName from the executor
	msg.Function.LocationName = executor.LocationName

	err = h.server.FunctionDB().AddFunction(msg.Function)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	addedFunction, err := h.server.FunctionDB().GetFunctionByID(msg.Function.FunctionID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"FunctionId": addedFunction.FunctionID, "ExecutorName": addedFunction.ExecutorName, "ColonyName": addedFunction.ColonyName, "FuncName": addedFunction.FuncName}).Debug("Adding function")

	jsonString, err = addedFunction.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetFunctions(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFunctionsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get function, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get function, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusForbidden)
		return
	}

	var functions []*core.Function
	if msg.ExecutorName == "" {
		// Get all functions in the colony
		functions, err = h.server.FunctionDB().GetFunctionsByColonyName(msg.ColonyName)
	} else {
		// Get functions for specific executor
		functions, err = h.server.FunctionDB().GetFunctionsByExecutorName(msg.ColonyName, msg.ExecutorName)
	}
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	jsonString, err = core.ConvertFunctionArrayToJSON(functions)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRemoveFunction(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveFunctionMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove function, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove function, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if len(msg.FunctionID) == 0 {
		h.server.HandleHTTPError(c, errors.New("Failed to remove function, msg.FunctionID is empty"), http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{"FunctionID": msg.FunctionID, "Length": len(msg.FunctionID)}).Debug("Looking up function by ID")
	function, err := h.server.FunctionDB().GetFunctionByID(msg.FunctionID)
	if err != nil {
		log.WithFields(log.Fields{"FunctionID": msg.FunctionID, "Error": err.Error()}).Error("Failed to get function by ID")
		h.server.HandleHTTPError(c, err, http.StatusBadRequest)
		return
	}

	if function == nil {
		log.WithField("FunctionID", msg.FunctionID).Error("Function not found")
		h.server.HandleHTTPError(c, errors.New("Failed to remove function, function does not exist"), http.StatusNotFound)
		return
	}
	log.WithField("Function", function).Debug("Found function for removal")

	executor, err := h.server.ExecutorDB().GetExecutorByName(function.ColonyName, function.ExecutorName)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, function.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if executor.ID != recoveredID {
		h.server.HandleHTTPError(c, errors.New("Not allowed to remove a function from another executor"), http.StatusForbidden)
		return
	}

	err = h.server.FunctionDB().RemoveFunctionByID(msg.FunctionID)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"FunctionId": msg.FunctionID}).Debug("Removing function")

	h.server.SendEmptyHTTPReply(c, payloadType)
}