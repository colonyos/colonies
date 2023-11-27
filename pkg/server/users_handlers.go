package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddUserHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddUserMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add user, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add user, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.User == nil {
		server.handleHTTPError(c, errors.New("Failed to add user, executor is nil"), http.StatusBadRequest)
		return
	}

	colony, err := server.db.GetColonyByName(msg.User.ColonyName)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if server.handleHTTPError(c, errors.New("Colony with name "+msg.User.ColonyName+" does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = server.validator.RequireColonyOwner(recoveredID, colony.ID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.db.AddUser(msg.User)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	addedUser, err := server.db.GetUserByName(colony.Name, msg.User.Name)
	if addedUser == nil {
		server.handleHTTPError(c, errors.New("Failed to add user, addedUser is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedUser.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": addedUser.ColonyName, "ColonyID": colony.ID, "Name": addedUser.Name, "UserID": addedUser.ID}).Debug("Adding user")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetUsersHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetUsersMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get users, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get users, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	colony, err := server.db.GetColonyByName(msg.ColonyName)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if server.handleHTTPError(c, errors.New("Colony with name "+msg.ColonyName+" does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = server.validator.RequireMembership(recoveredID, colony.ID, false)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	users, err := server.db.GetUsersByColonyName(colony.Name)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertUserArrayToJSON(users)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": colony.ID}).Debug("Getting users")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetUserHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetUserMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get user, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get user, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	colony, err := server.db.GetColonyByName(msg.ColonyName)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if server.handleHTTPError(c, errors.New("Colony with name "+msg.ColonyName+" does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = server.validator.RequireMembership(recoveredID, colony.ID, false)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	user, err := server.db.GetUserByName(msg.ColonyName, msg.Name)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = user.ToJSON()
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": colony.ID, "ColonyName": msg.ColonyName, "Name": msg.Name}).Debug("Getting user")

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteUserHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteUserMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to remove user, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to remove user, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	colony, err := server.db.GetColonyByName(msg.ColonyName)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if server.handleHTTPError(c, errors.New("Colony with name "+msg.ColonyName+" does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = server.validator.RequireColonyOwner(recoveredID, colony.ID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.db.DeleteUserByName(colony.Name, msg.Name)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "ColonyID": colony.ID, "UserName": msg.Name}).Debug("Removing user")

	server.sendEmptyHTTPReply(c, payloadType)
}
