package cron

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	cronlib "github.com/colonyos/colonies/pkg/cron"
	serverutils "github.com/colonyos/colonies/pkg/server/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	Validator() security.Validator
	CronController() interface {
		AddCron(cron *core.Cron) (*core.Cron, error)
		GetCron(cronID string) (*core.Cron, error)
		GetCrons(colonyName string, count int) ([]*core.Cron, error)
		RunCron(cronID string) (*core.Cron, error)
		RemoveCron(cronID string) error
		GetCronPeriod() int
	}
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
	if err := handlerRegistry.RegisterGin(rpc.AddCronPayloadType, h.HandleAddCron); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetCronPayloadType, h.HandleGetCron); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetCronsPayloadType, h.HandleGetCrons); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RunCronPayloadType, h.HandleRunCron); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RemoveCronPayloadType, h.HandleRemoveCron); err != nil {
		return err
	}
	return nil
}

// resolveInitiator resolves the initiator name from the recoveredID
func (h *Handlers) resolveInitiator(colonyName string, recoveredID string) (string, error) {
	executor, err := h.server.ExecutorDB().GetExecutorByID(recoveredID)
	if err != nil {
		return "", err
	}

	if executor != nil {
		return executor.Name, nil
	} else {
		user, err := h.server.UserDB().GetUserByID(colonyName, recoveredID)
		if err != nil {
			return "", err
		}
		if user != nil {
			return user.Name, nil
		} else {
			return "", errors.New("Could not derive InitiatorName")
		}
	}
}

func (h *Handlers) HandleAddCron(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddCronMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Cron == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add cron, msg.Cron is nil"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.Cron.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Validate that workflow and cron expression is valid
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(msg.Cron.WorkflowSpec)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	err = serverutils.VerifyWorkflowSpec(workflowSpec)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if msg.Cron.Interval == 0 {
		if h.server.HandleHTTPError(c, errors.New("Cron interval must be -1 (disabled) or larger than 0"), http.StatusBadRequest) {
			return
		}
	}

	if msg.Cron.Interval == -1 {
		_, err = cronlib.Next(msg.Cron.CronExpression)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		if msg.Cron.Random {
			if h.server.HandleHTTPError(c, errors.New("Random cron is only supported when specifying intervals"), http.StatusBadRequest) {
				return
			}
		}
	}

	msg.Cron.ID = core.GenerateRandomID()
	msg.Cron.InitiatorID = recoveredID

	initiatorName, err := h.resolveInitiator(workflowSpec.ColonyName, recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	msg.Cron.InitiatorID = recoveredID
	msg.Cron.InitiatorName = initiatorName

	addedCron, err := h.server.CronController().AddCron(msg.Cron)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedCron == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add cron, addedCron is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedCron.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronId": addedCron.ID}).Debug("Adding cron")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetCron(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetCronMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	cron, err := h.server.CronController().GetCron(msg.CronID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if cron == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get cron, cron is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, cron.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	cron.CheckerPeriod = h.server.CronController().GetCronPeriod()

	jsonString, err = cron.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Getting cron")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetCrons(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetCronsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get crons, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get crons, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	crons, err := h.server.CronController().GetCrons(msg.ColonyName, msg.Count)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if crons == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get crons, crons is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = core.ConvertCronArrayToJSON(crons)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName, "Count": msg.Count}).Debug("Getting crons")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRunCron(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRunCronMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to run cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to run cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	cron, err := h.server.CronController().RunCron(msg.CronID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if cron == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to run cron, cron is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, cron.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = cron.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Running cron")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRemoveCron(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveCronMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove cron, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove cron, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	cron, err := h.server.CronController().GetCron(msg.CronID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if cron == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove cron, cron is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, cron.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.CronController().RemoveCron(cron.ID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Removing cron")

	h.server.SendEmptyHTTPReply(c, payloadType)
}