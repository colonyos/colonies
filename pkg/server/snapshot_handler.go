package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleCreateSnapshotHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCreateSnapshotMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to create snapshot, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to create snapshot, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	snapshot, err := server.db.CreateSnapshot(msg.ColonyID, msg.Label, msg.Name)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	jsonStr, err := snapshot.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"SnapshotID": snapshot.ID}).Debug("Creating snapshot")

	server.sendHTTPReply(c, payloadType, jsonStr)
}

func (server *ColoniesServer) handleGetSnapshotHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetSnapshotMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get snapshot, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get snapshot, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	var snapshot *core.Snapshot
	if msg.SnapshotID != "" {
		snapshot, err = server.db.GetSnapshotByID(msg.ColonyID, msg.SnapshotID)
		if server.handleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else if msg.Name != "" {
		snapshot, err = server.db.GetSnapshotByName(msg.ColonyID, msg.Name)
		if server.handleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else {
		if server.handleHTTPError(c, errors.New("Failed to get snapshot, malformatted msg"), http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	}

	jsonStr, err := snapshot.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"SnapshotID": snapshot.ID}).Debug("Getting snapshot")

	server.sendHTTPReply(c, payloadType, jsonStr)
}

func (server *ColoniesServer) handleGetSnapshotsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetSnapshotsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get snapshots, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get snapshots, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	snapshots, err := server.db.GetSnapshotsByColonyID(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	jsonStr, err := core.ConvertSnapshotArrayToJSON(snapshots)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"ColonyID": msg.ColonyID}).Debug("Getting snapshots")

	server.sendHTTPReply(c, payloadType, jsonStr)
}

func (server *ColoniesServer) handleDeleteSnapshotHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteSnapshotMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete snapshot, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete snapshot, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if msg.SnapshotID != "" {
		err = server.db.DeleteSnapshotByID(msg.ColonyID, msg.SnapshotID)
		if server.handleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else if msg.Name != "" {
		err = server.db.DeleteSnapshotByName(msg.ColonyID, msg.Name)
		if server.handleHTTPError(c, err, http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	} else {
		if server.handleHTTPError(c, errors.New("Failed to delete snapsnot, malformatted msg"), http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	}

	log.WithFields(log.Fields{"SnapshotID": msg.SnapshotID}).Debug("Deleting snapshot")

	server.sendEmptyHTTPReply(c, payloadType)
}
