package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddAttributeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddAttributeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to add attribute, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add attribute, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Attribute == nil {
		server.handleHTTPError(c, errors.New("Failed to add attribute, msg.Attribute is nil"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.Attribute.TargetID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to add attribute, process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("Failed to add attribute, only runtime with id <" + process.AssignedRuntimeID + "> is allowed to set attributes")
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	msg.Attribute.GenerateID()

	addedAttribute, err := server.controller.addAttribute(msg.Attribute)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedAttribute == nil {
		server.handleHTTPError(c, errors.New("Failed to add attribute, addedAttribute is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedAttribute.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"AttributeID": msg.Attribute.ID}).Info("Adding attribute")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetAttributeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetAttributeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("Failed to get attribute, failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get attribute, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	attribute, err := server.controller.getAttribute(msg.AttributeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if attribute == nil {
		server.handleHTTPError(c, errors.New("Failed to get attribute, attribute is nil"), http.StatusInternalServerError)
		return
	}

	process, err := server.controller.getProcessByID(attribute.TargetID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("Failed to get attribute, process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = attribute.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"AttributeID": msg.AttributeID}).Info("Getting attribute")

	server.sendHTTPReply(c, payloadType, jsonString)
}
