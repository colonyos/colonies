package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	cronlib "github.com/colonyos/colonies/pkg/cron"
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

	err = server.validator.RequireMembership(recoveredID, msg.Cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Validate that workflow and cron expression is valid
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(msg.Cron.WorkflowSpec)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	err = VerifyWorkflowSpec(workflowSpec)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if msg.Cron.Interval == 0 {
		if server.handleHTTPError(c, errors.New("Cron interval must be -1 (disabled) or larger than 0"), http.StatusBadRequest) {
			return
		}
	}

	if msg.Cron.Interval == -1 {
		_, err = cronlib.Next(msg.Cron.CronExpression)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		if msg.Cron.Random {
			if server.handleHTTPError(c, errors.New("Random cron is only supported when specifying intervals"), http.StatusBadRequest) {
				return
			}
		}
	}

	msg.Cron.ID = core.GenerateRandomID()
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

	log.WithFields(log.Fields{"CronId": addedCron.ID}).Debug("Adding cron")

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

	err = server.validator.RequireMembership(recoveredID, cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	cron.CheckerPeriod = server.controller.getCronPeriod()

	jsonString, err = cron.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Getting cron")

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

	err = server.validator.RequireMembership(recoveredID, msg.ColonyID, true)
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

	log.WithFields(log.Fields{"ColonyId": msg.ColonyID, "Count": msg.Count}).Debug("Getting crons")

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

	err = server.validator.RequireMembership(recoveredID, cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = cron.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Running cron")

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

	err = server.validator.RequireMembership(recoveredID, cron.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteCron(cron.ID)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Deleting cron")

	server.sendEmptyHTTPReply(c, payloadType)
}
