package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) handleAddFileHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddFileMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to add file, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.File == nil {
		server.handleHTTPError(c, errors.New("Failed to add file, msg.File is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, msg.File.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Bypass colonies controller and use the database directly, no need to synchronize this operation since files are inmutable
	file := msg.File
	file.ID = core.GenerateRandomID()
	server.db.AddFile(msg.File)

	log.WithFields(log.Fields{"FileID": file.ID}).Debug("Adding file")

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleGetFileHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFileMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to add log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get file, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Bypass colonies controller and use the database directly, no need to synchronize this operation since files are inmutable
	var files []*core.File
	var savedError error
	if msg.FileID != "" {
		file, err := server.db.GetFileByID(msg.ColonyID, msg.FileID)
		if err != nil {
			savedError = err
		} else {
			files = []*core.File{file}
		}
	} else if msg.Prefix != "" && msg.Name != "" {
		if msg.Latest {
			files, err = server.db.GetLatestFileByName(msg.ColonyID, msg.Prefix, msg.Name)
			if err != nil {
				savedError = err
			}
		} else {
			files, err = server.db.GetFileByName(msg.ColonyID, msg.Prefix, msg.Name)
			if err != nil {
				savedError = err
			}
		}
	} else {
		if server.handleHTTPError(c, errors.New("malformat get file msg"), http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	}

	if len(files) > 0 {
		if server.handleHTTPError(c, errors.New("Failed to get file"), http.StatusNotFound) {
			log.Error(err)
			return
		}
	} else {
		for _, file := range files {
			err = server.validator.RequireExecutorMembership(recoveredID, file.ColonyID, true)
			if server.handleHTTPError(c, err, http.StatusForbidden) {
				log.Error(err)
				return
			}
		}
	}

	if server.handleHTTPError(c, savedError, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	jsonStr, err := core.ConvertFileArrayToJSON(files)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to files")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"FileID": msg.FileID, "Prefix": msg.Prefix, "Name": msg.Name, "Latest": msg.Latest}).Debug("Getting file")

	server.sendHTTPReply(c, payloadType, jsonStr)
}

func (server *ColoniesServer) handleGetFilesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFilesMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get files, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get files, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
}

func (server *ColoniesServer) handleGetFilePrefixesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
}
