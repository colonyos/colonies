package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddColonyHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddColonyMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add colony, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add colony, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Colony == nil {
		server.handleHTTPError(c, errors.New("Failed to add colony, colony is nil"), http.StatusBadRequest)
		return
	}

	if len(msg.Colony.ID) != 64 {
		server.handleHTTPError(c, errors.New("Failed to add colony, invalid colony id length"), http.StatusBadRequest)
		return
	}

	addedColony, err := server.controller.addColony(msg.Colony)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if addedColony == nil {
		server.handleHTTPError(c, errors.New("Failed to add colony, addedColony is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedColony.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": addedColony.ID}).Debug("Adding colony")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteColonyHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteColonyMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete colony, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete colony, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	log.WithFields(log.Fields{"ColonyId": msg.ColonyID}).Debug("Deleting colony")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleGetColoniesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColoniesMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get colonies, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get colonies, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	log.Debug("Getting colonies")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetColonyHTTPRequest(c *gin.Context, recoveredID, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColonyMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get colony, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get colony, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.controller.getColony(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if colony == nil {
		server.handleHTTPError(c, errors.New("Failed to get colony, colony is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = colony.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyID}).Debug("Getting colony")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleColonyStatisticsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColonyStatisticsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get colony statistics, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get colony statistics, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, msg.ColonyID, true)
	if err != nil {
		err = server.validator.RequireColonyOwner(recoveredID, msg.ColonyID)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	stat, err := server.controller.getColonyStatistics(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = stat.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyID}).Debug("Getting colony statistics")

	server.sendHTTPReply(c, payloadType, jsonString)
}
