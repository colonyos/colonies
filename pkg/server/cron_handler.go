package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddCronHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
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
