package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
)

func (server *ColoniesServer) handleAddColonyHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddColonyMsgFromJSON(jsonString)
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

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Colony == nil {
		server.handleHTTPError(c, errors.New("colony is nil"), http.StatusBadRequest)
		return
	}

	if len(msg.Colony.ID) != 64 {
		server.handleHTTPError(c, errors.New("invalid colony id length"), http.StatusBadRequest)
		return
	}

	addedColony, err := server.controller.addColony(msg.Colony)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedColony == nil {
		server.handleHTTPError(c, errors.New("addedColony is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedColony.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteColonyHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteColonyMsgFromJSON(jsonString)
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

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteColony(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleGetColoniesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColoniesMsgFromJSON(jsonString)
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

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colonies, err := server.controller.getColonies()
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertColonyArrayToJSON(colonies)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetColonyHTTPRequest(c *gin.Context, recoveredID, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColonyMsgFromJSON(jsonString)
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
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.controller.getColonyByID(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if colony == nil {
		server.handleHTTPError(c, errors.New("colony is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = colony.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}
