package snapshot

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
	Validator() security.Validator
	SnapshotDB() database.SnapshotDatabase
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
	if err := handlerRegistry.RegisterGin(rpc.CreateSnapshotPayloadType, h.HandleCreateSnapshot); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetSnapshotPayloadType, h.HandleGetSnapshot); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.GetSnapshotsPayloadType, h.HandleGetSnapshots); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RemoveSnapshotPayloadType, h.HandleRemoveSnapshot); err != nil {
		return err
	}
	if err := handlerRegistry.RegisterGin(rpc.RemoveAllSnapshotsPayloadType, h.HandleRemoveAllSnapshots); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleCreateSnapshot(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCreateSnapshotMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to create snapshot, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to create snapshot, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	snapshot, err := h.server.SnapshotDB().CreateSnapshot(msg.ColonyName, msg.Label, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	jsonStr, err := snapshot.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "SnapshotID": snapshot.ID}).Debug("Creating snapshot")

	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleGetSnapshot(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetSnapshotMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get snapshot, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get snapshot, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	var snapshot *core.Snapshot
	if msg.SnapshotID != "" {
		snapshot, err = h.server.SnapshotDB().GetSnapshotByID(msg.ColonyName, msg.SnapshotID)
		if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else if msg.Name != "" {
		snapshot, err = h.server.SnapshotDB().GetSnapshotByName(msg.ColonyName, msg.Name)
		if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else {
		if h.server.HandleHTTPError(c, errors.New("Failed to get snapshot, malformatted msg"), http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	}

	jsonStr, err := snapshot.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "SnapshotID": snapshot.ID}).Debug("Getting snapshot")

	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleGetSnapshots(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetSnapshotsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get snapshots, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get snapshots, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	snapshots, err := h.server.SnapshotDB().GetSnapshotsByColonyName(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	jsonStr, err := core.ConvertSnapshotArrayToJSON(snapshots)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName}).Debug("Getting snapshots")

	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleRemoveSnapshot(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveSnapshotMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove snapshot, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove snapshot, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if msg.SnapshotID != "" {
		err = h.server.SnapshotDB().RemoveSnapshotByID(msg.ColonyName, msg.SnapshotID)
		if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else if msg.Name != "" {
		err = h.server.SnapshotDB().RemoveSnapshotByName(msg.ColonyName, msg.Name)
		if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove snapsnot, malformatted msg"), http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	}

	log.WithFields(log.Fields{"SnapshotID": msg.SnapshotID, "ColonyName": msg.ColonyName}).Debug("Removing snapshot")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleRemoveAllSnapshots(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveAllSnapshotsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove snapshot, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove snapshot, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	err = h.server.SnapshotDB().RemoveSnapshotsByColonyName(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName}).Debug("Removing all snapshots")

	h.server.SendEmptyHTTPReply(c, payloadType)
}