package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddGeneratorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddGeneratorMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to add generator, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add generator, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Generator == nil {
		server.handleHTTPError(c, errors.New("Failed to add generator, msg.ProcessSpec is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.Generator.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

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

	log.WithFields(log.Fields{"GeneratorID": addedGenerator.ID}).Info("Adding generator")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetGeneratorHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetGeneratorMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get generator, failed to parse JSON"), http.StatusBadRequest)
		return
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

	err = server.validator.RequireRuntimeMembership(recoveredID, generator.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = generator.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"GeneratorID": generator.ID}).Info("Getting generator")

	server.sendHTTPReply(c, payloadType, jsonString)
}
