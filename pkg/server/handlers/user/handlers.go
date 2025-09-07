package user

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	GetUserDB() database.UserDatabase
	GetColonyDB() database.ColonyDatabase
	GetValidator() security.Validator
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
	if err := handlerRegistry.RegisterGin(rpc.AddUserPayloadType, h.HandleAddUser); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetUsersPayloadType, h.HandleGetUsers); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetUserPayloadType, h.HandleGetUser); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RemoveUserPayloadType, h.HandleRemoveUser); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddUser(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddUserMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add user, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add user, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.User == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add user, executor is nil"), http.StatusBadRequest)
		return
	}

	colony, err := h.server.GetColonyDB().GetColonyByName(msg.User.ColonyName)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if h.server.HandleHTTPError(c, errors.New("Colony with name <"+msg.User.ColonyName+"> does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetValidator().RequireColonyOwner(recoveredID, colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	userExist, err := h.server.GetUserDB().GetUserByName(msg.User.ColonyName, msg.User.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if userExist != nil {
		if h.server.HandleHTTPError(c, errors.New("A user with name <"+msg.User.Name+"> already exists in Colony with name <"+msg.User.ColonyName+">"), http.StatusBadRequest) {
			return
		}
	}

	userExist, err = h.server.GetUserDB().GetUserByID(msg.User.ID, msg.User.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if userExist != nil {
		if h.server.HandleHTTPError(c, errors.New("A user with Id <"+msg.User.ID+"> already exists in Colony with name <"+msg.User.ColonyName+">"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetUserDB().AddUser(msg.User)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	addedUser, err := h.server.GetUserDB().GetUserByName(colony.Name, msg.User.Name)
	if addedUser == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add user, addedUser is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedUser.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": addedUser.ColonyName, "Name": addedUser.Name, "UserID": addedUser.ID}).Debug("Adding user")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetUsers(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetUsersMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get users, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get users, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	colony, err := h.server.GetColonyDB().GetColonyByName(msg.ColonyName)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if h.server.HandleHTTPError(c, errors.New("Colony with name <"+msg.ColonyName+"> does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetValidator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	users, err := h.server.GetUserDB().GetUsersByColonyName(colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertUserArrayToJSON(users)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": colony.ID}).Debug("Getting users")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetUser(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetUserMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get user, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get user, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	colony, err := h.server.GetColonyDB().GetColonyByName(msg.ColonyName)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if h.server.HandleHTTPError(c, errors.New("Colony with name <"+msg.ColonyName+"> does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetValidator().RequireMembership(recoveredID, msg.ColonyName, false)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	user, err := h.server.GetUserDB().GetUserByName(msg.ColonyName, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if user == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get user, user is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = user.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "Name": msg.Name}).Debug("Getting user")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRemoveUser(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveUserMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove user, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove user, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	colony, err := h.server.GetColonyDB().GetColonyByName(msg.ColonyName)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if h.server.HandleHTTPError(c, errors.New("Colony with name <"+msg.ColonyName+"> does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetValidator().RequireColonyOwner(recoveredID, colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	user, err := h.server.GetUserDB().GetUserByName(msg.ColonyName, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if user == nil {
		if h.server.HandleHTTPError(c, errors.New("User with name <"+msg.Name+"> not found"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetUserDB().RemoveUserByName(colony.Name, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "Name": msg.Name}).Debug("Removing user")

	h.server.SendEmptyHTTPReply(c, payloadType)
}