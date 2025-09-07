package generator

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	serverutils "github.com/colonyos/colonies/pkg/server/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	Validator() security.Validator
	GeneratorController() Controller
	GeneratorDB() database.GeneratorDatabase
	ExecutorDB() database.ExecutorDatabase
	UserDB() database.UserDatabase
}

type Controller interface {
	AddGenerator(generator *core.Generator) (*core.Generator, error)
	GetGenerator(generatorID string) (*core.Generator, error)
	ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error)
	GetGenerators(colonyName string, count int) ([]*core.Generator, error)
	PackGenerator(generatorID string, colonyName string, arg string) error
	RemoveGenerator(generatorID string) error
	GetGeneratorPeriod() int
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
	if err := handlerRegistry.RegisterGin(rpc.AddGeneratorPayloadType, h.HandleAddGenerator); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetGeneratorPayloadType, h.HandleGetGenerator); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.ResolveGeneratorPayloadType, h.HandleResolveGenerator); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetGeneratorsPayloadType, h.HandleGetGenerators); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.PackGeneratorPayloadType, h.HandlePackGenerator); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RemoveGeneratorPayloadType, h.HandleRemoveGenerator); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddGenerator(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Generator == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add generator, msg.Generator is nil"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.Generator.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Validate that workflow is valid
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(msg.Generator.WorkflowSpec)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	err = serverutils.VerifyWorkflowSpec(workflowSpec)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	msg.Generator.ID = core.GenerateRandomID()

	initiatorName, err := h.resolveInitiator(msg.Generator.ColonyName, recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	msg.Generator.InitiatorID = recoveredID
	msg.Generator.InitiatorName = initiatorName

	addedGenerator, err := h.server.GeneratorController().AddGenerator(msg.Generator)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedGenerator == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add generator, addedGenerator is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedGenerator.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": addedGenerator.ID}).Debug("Adding generator")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetGenerator(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := h.server.GeneratorController().GetGenerator(msg.GeneratorID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, generator.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	generator.CheckerPeriod = h.server.GeneratorController().GetGeneratorPeriod()
	queueSize, err := h.server.GeneratorDB().CountGeneratorArgs(generator.ID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}
	generator.QueueSize = queueSize

	jsonString, err = generator.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID}).Debug("Getting generator")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleResolveGenerator(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateResolveGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to resolve generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to resolve generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := h.server.GeneratorController().ResolveGenerator(msg.ColonyName, msg.GeneratorName)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to resolve generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, generator.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = generator.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID, "GeneratorName": generator.Name}).Debug("Resolving generator")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetGenerators(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetGeneratorsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get generators, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get generators, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	generators, err := h.server.GeneratorController().GetGenerators(msg.ColonyName, msg.Count)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertGeneratorArrayToJSON(generators)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyName, "Count": msg.Count}).Debug("Getting generators")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandlePackGenerator(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreatePackGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to inc generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to increment generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := h.server.GeneratorController().GetGenerator(msg.GeneratorID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to increment generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, generator.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.GeneratorController().PackGenerator(generator.ID, generator.ColonyName, msg.Arg)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID, "Arg": msg.Arg}).Debug("Adding arg to generator")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleRemoveGenerator(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := h.server.GeneratorController().GetGenerator(msg.GeneratorID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, generator.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.GeneratorController().RemoveGenerator(generator.ID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID}).Debug("Removing generator")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// resolveInitiator resolves the initiator name from the recoveredID
func (h *Handlers) resolveInitiator(colonyName string, recoveredID string) (string, error) {
	executor, err := h.server.ExecutorDB().GetExecutorByID(recoveredID)
	if err == nil && executor != nil {
		return executor.Name, nil
	}

	user, err := h.server.UserDB().GetUserByID(colonyName, recoveredID)
	if err == nil && user != nil {
		return user.Name, nil
	}

	return "", errors.New("Failed to resolve initiator")
}