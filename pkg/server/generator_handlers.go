package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddGeneratorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Generator == nil {
		server.handleHTTPError(c, errors.New("Failed to add generator, msg.Generator is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, msg.Generator.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Validate that workflow is valid
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(msg.Generator.WorkflowSpec)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	err = VerifyWorkflowSpec(workflowSpec)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	msg.Generator.ID = core.GenerateRandomID()
	addedGenerator, err := server.controller.addGenerator(msg.Generator)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedGenerator == nil {
		server.handleHTTPError(c, errors.New("Failed to add generator, addedGenerator is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedGenerator.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": addedGenerator.ID}).Debug("Adding generator")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetGeneratorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := server.controller.getGenerator(msg.GeneratorID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		server.handleHTTPError(c, errors.New("Failed to get generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, generator.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	generator.CheckerPeriod = server.controller.getGeneratorPeriod()
	queueSize, err := server.db.CountGeneratorArgs(generator.ID)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}
	generator.QueueSize = queueSize

	jsonString, err = generator.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID}).Debug("Getting generator")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleResolveGeneratorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateResolveGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to resolve generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to resolve generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := server.controller.resolveGenerator(msg.GeneratorName)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		server.handleHTTPError(c, errors.New("Failed to resolve generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, generator.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = generator.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID, "GeneratorName": generator.Name}).Debug("Resolving generator")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetGeneratorsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetGeneratorsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get generators, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get generators, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	generators, err := server.controller.getGenerators(msg.ColonyID, msg.Count)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generators == nil {
		server.handleHTTPError(c, errors.New("Failed to get generators, generators is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = core.ConvertGeneratorArrayToJSON(generators)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": msg.ColonyID, "Count": msg.Count}).Debug("Getting generators")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handlePackGeneratorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreatePackGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to inc generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to increment generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := server.controller.getGenerator(msg.GeneratorID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		server.handleHTTPError(c, errors.New("Failed to increment generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, generator.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.packGenerator(generator.ID, generator.ColonyID, msg.Arg)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID, "Arg": msg.Arg}).Debug("Adding arg to generator")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleDeleteGeneratorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteGeneratorMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete generator, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	generator, err := server.controller.getGenerator(msg.GeneratorID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if generator == nil {
		server.handleHTTPError(c, errors.New("Failed to delete generator, generator is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, generator.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteGenerator(generator.ID)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorId": generator.ID}).Debug("Deleting generator")

	server.sendEmptyHTTPReply(c, payloadType)
}
