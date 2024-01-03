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

	serverID, err := server.getServerID()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Colony == nil {
		server.handleHTTPError(c, errors.New("Failed to add colony, colony is <nil>"), http.StatusBadRequest)
		return
	}

	if len(msg.Colony.ID) != 64 {
		server.handleHTTPError(c, errors.New("Failed to add colony, invalid colony Id length"), http.StatusBadRequest)
		return
	}

	colonyExist, err := server.db.GetColonyByName(msg.Colony.Name)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if colonyExist != nil {
		if server.handleHTTPError(c, errors.New("A Colony with name <"+msg.Colony.Name+"> already exists"), http.StatusBadRequest) {
			return
		}
	}

	addedColony, err := server.controller.addColony(msg.Colony)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if addedColony == nil {
		server.handleHTTPError(c, errors.New("Failed to add colony, addedColony is <nil>"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedColony.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.Colony.Name, "ColonyId": addedColony.ID}).Debug("Adding colony")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleRemoveColonyHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveColonyMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to remove colony, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to remove colony, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	colony, err := server.db.GetColonyByName(msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if colony == nil {
		if server.handleHTTPError(c, errors.New("Colony with name <"+msg.ColonyName+"> not found"), http.StatusBadRequest) {
			return
		}
	}

	err = server.controller.removeColony(colony.Name)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": colony.ID, "ColonyName": colony.Name}).Debug("Removing colony")

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

	serverID, err := server.getServerID()

	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, serverID)
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

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.controller.getColony(msg.ColonyName)
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

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName}).Debug("Getting colony")

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

	colony, err := server.db.GetColonyByName(msg.ColonyName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if colony == nil {
		if server.handleHTTPError(c, errors.New("Colony with name <"+msg.ColonyName+"> not found"), http.StatusBadRequest) {
			return
		}
	}

	err = server.validator.RequireMembership(recoveredID, colony.Name, true)
	if err != nil {
		return
	}

	stat, err := server.controller.getColonyStatistics(colony.Name)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = stat.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName}).Debug("Getting colony statistics")

	server.sendHTTPReply(c, payloadType, jsonString)
}
