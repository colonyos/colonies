package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddCronHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddCronMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Cron == nil {
		server.handleHTTPError(c, errors.New("Failed to add cron, msg.Cron is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.Cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	addedCron, err := server.controller.addCron(msg.Cron)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedCron == nil {
		server.handleHTTPError(c, errors.New("Failed to add cron, addedCron is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedCron.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronID": addedCron.ID}).Info("Adding cron")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetCronHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetCronMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	cron, err := server.controller.getCron(msg.CronID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if cron == nil {
		server.handleHTTPError(c, errors.New("Failed to get cron, cron is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = cron.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronID": cron.ID}).Info("Getting cron")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetCronsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetCronsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get crons, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get crons, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	crons, err := server.controller.getCrons(msg.ColonyID, msg.Count)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if crons == nil {
		server.handleHTTPError(c, errors.New("Failed to get crons, crons is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = core.ConvertCronArrayToJSON(crons)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyID": msg.ColonyID, "Count": msg.Count}).Info("Getting crons")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleRunCronHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRunCronMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to run cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to run cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	cron, err := server.controller.runCron(msg.CronID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if cron == nil {
		server.handleHTTPError(c, errors.New("Failed to run cron, cron is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = cron.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronID": cron.ID}).Info("Running cron")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteCronHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteCronMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	cron, err := server.controller.getCron(msg.CronID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if cron == nil {
		server.handleHTTPError(c, errors.New("Failed to delete cron, cron is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteCron(cron.ID)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronID": cron.ID}).Info("Deleting cron")

	server.sendEmptyHTTPReply(c, payloadType)
}
