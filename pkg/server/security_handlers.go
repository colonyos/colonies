package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleChangeUserIDHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeUserIDMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to change user Id, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to change user Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.UserID == "" {
		server.handleHTTPError(c, errors.New("Failed to change user Id, user Id is empty"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, false)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	user, err := server.db.GetUserByID(msg.ColonyName, recoveredID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if user == nil {
		server.handleHTTPError(c, errors.New("Failed to change user Id, user not found"), http.StatusBadRequest)
		return
	}

	if len(msg.UserID) != 64 {
		server.handleHTTPError(c, errors.New("Failed to change user Id, new user Id is not 64 characters"), http.StatusBadRequest)
		return
	}

	err = server.db.ChangeUserID(msg.ColonyName, user.ID, msg.UserID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName": msg.ColonyName,
		"Name":       user.Name,
		"OldUserID":  user.ID,
		"NewUserId":  msg.UserID}).
		Debug("Changing user Id")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleChangeExecutorIDHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeExecutorIDMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to change executor Id, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to change executor Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ExecutorID == "" {
		server.handleHTTPError(c, errors.New("Failed to change executor Id, executor Id is empty"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, false)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	executor, err := server.db.GetExecutorByID(recoveredID) // TODO: GetExecutorByID should take colony name
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if executor == nil {
		server.handleHTTPError(c, errors.New("Failed to change executor Id, executor not found"), http.StatusBadRequest)
		return
	}

	if len(msg.ExecutorID) != 64 {
		server.handleHTTPError(c, errors.New("Failed to change executor Id, new executor Id is not 64 characters"), http.StatusBadRequest)
		return
	}

	err = server.db.ChangeExecutorID(msg.ColonyName, executor.ID, msg.ExecutorID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName":    msg.ColonyName,
		"Name":          executor.Name,
		"OldExecutorID": executor.ID,
		"NewExecutorId": msg.ExecutorID}).
		Debug("Changing executor Id")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleChangeColonyIDHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeColonyIDMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to change colony Id, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to change colony Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ColonyID == "" {
		server.handleHTTPError(c, errors.New("Failed to change colony Id, colony Id is empty"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.db.GetColonyByName(msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if len(msg.ColonyID) != 64 {
		server.handleHTTPError(c, errors.New("Failed to change colony Id, new colony Id is not 64 characters"), http.StatusBadRequest)
		return
	}

	err = server.db.ChangeColonyID(msg.ColonyName, colony.ID, msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName":  msg.ColonyName,
		"OldColonyID": colony.ID,
		"NewColonyId": msg.ColonyID}).
		Debug("Changing colony Id")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleChangeServerIDHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChangeServerIDMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to change colony Id, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to change colony Id, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ServerID == "" {
		server.handleHTTPError(c, errors.New("Failed to change colony Id, colony Id is empty"), http.StatusBadRequest)
		return
	}

	serverID, err := server.getServerID()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.db.SetServerID(serverID, msg.ServerID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	log.WithFields(log.Fields{
		"OldServerID": serverID,
		"NewServerId": msg.ServerID}).
		Debug("Changing server Id")

	server.sendHTTPReply(c, payloadType, jsonString)
}
