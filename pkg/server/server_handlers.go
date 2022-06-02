package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
)

func (server *ColoniesServer) handleVersionHTTPRequest(c *gin.Context, payloadType string, jsonString string) {
	msg, err := rpc.CreateVersionMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	versionMsg := rpc.CreateVersionMsg(build.BuildVersion, build.BuildTime)

	jsonString, err = versionMsg.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleStatisticsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetStatisticsMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
		return
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	stat, err := server.controller.getStatistics()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = stat.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}
