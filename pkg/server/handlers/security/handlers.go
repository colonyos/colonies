package security

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	log "github.com/sirupsen/logrus"
)

// Server defines the interface that this handler needs from the server
type Server interface {
	// HTTP Response methods using backend abstraction
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	
	// Server identity
	GetServerID() (string, error)
	
	// Security and validation
	Validator() security.Validator
	
	// Database access
	UserDB() database.UserDatabase
	ExecutorDB() database.ExecutorDatabase
	ColonyDB() database.ColonyDatabase
	SecurityDB() database.SecurityDatabase
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{server: server}
}

// RegisterHandlers implements the HandlerRegistrar interface
func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	if err := handlerRegistry.Register(rpc.ChangeUserIDPayloadType, h.HandleChangeUserID); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ChangeExecutorIDPayloadType, h.HandleChangeExecutorID); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ChangeColonyIDPayloadType, h.HandleChangeColonyID); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ChangeServerIDPayloadType, h.HandleChangeServerID); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleChangeUserID(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeUserIDMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to change user Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.UserID == "" {
		h.server.HandleHTTPError(c, errors.New("Failed to change user Id, user Id is empty"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	user, err := h.server.UserDB().GetUserByID(msg.ColonyName, recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if user == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to change user Id, user not found"), http.StatusBadRequest)
		return
	}

	if len(msg.UserID) != 64 {
		h.server.HandleHTTPError(c, errors.New("Failed to change user Id, new user Id is not 64 characters"), http.StatusBadRequest)
		return
	}

	err = h.server.SecurityDB().ChangeUserID(msg.ColonyName, user.ID, msg.UserID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName": msg.ColonyName,
		"Name":       user.Name,
		"OldUserID":  user.ID,
		"NewUserId":  msg.UserID}).
		Debug("Changing user Id")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleChangeExecutorID(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeExecutorIDMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to change executor Id, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to change executor Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ExecutorID == "" {
		h.server.HandleHTTPError(c, errors.New("Failed to change executor Id, executor Id is empty"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executor, err := h.server.ExecutorDB().GetExecutorByID(recoveredID) // TODO: GetExecutorByID should take colony name
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if executor == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to change executor Id, executor not found"), http.StatusBadRequest)
		return
	}

	if len(msg.ExecutorID) != 64 {
		h.server.HandleHTTPError(c, errors.New("Failed to change executor Id, new executor Id is not 64 characters"), http.StatusBadRequest)
		return
	}

	err = h.server.SecurityDB().ChangeExecutorID(msg.ColonyName, executor.ID, msg.ExecutorID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName":    msg.ColonyName,
		"Name":          executor.Name,
		"OldExecutorID": executor.ID,
		"NewExecutorId": msg.ExecutorID}).
		Debug("Changing executor Id")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleChangeColonyID(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeColonyIDMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to change colony Id, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to change colony Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ColonyID == "" {
		h.server.HandleHTTPError(c, errors.New("Failed to change colony Id, colony Id is empty"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := h.server.ColonyDB().GetColonyByName(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if len(msg.ColonyID) != 64 {
		h.server.HandleHTTPError(c, errors.New("Failed to change colony Id, new colony Id is not 64 characters"), http.StatusBadRequest)
		return
	}

	err = h.server.SecurityDB().ChangeColonyID(msg.ColonyName, colony.ID, msg.ColonyID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName":  msg.ColonyName,
		"OldColonyID": colony.ID,
		"NewColonyId": msg.ColonyID}).
		Debug("Changing colony Id")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleChangeServerID(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeServerIDMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to change colony Id, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to change colony Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ServerID == "" {
		h.server.HandleHTTPError(c, errors.New("Failed to change colony Id, colony Id is empty"), http.StatusBadRequest)
		return
	}

	serverID, err := h.server.GetServerID()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	err = h.server.Validator().RequireServerOwner(recoveredID, serverID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.SecurityDB().SetServerID(serverID, msg.ServerID)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"OldServerID": serverID,
		"NewServerId": msg.ServerID}).
		Debug("Changing server Id")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}