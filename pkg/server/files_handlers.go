package server

import (
	"encoding/json"
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

	addedFile, err := server.db.GetFileByID(msg.File.ColonyID, file.ID)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	jsonStr, err := addedFile.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"FileID": file.ID}).Debug("Adding file")

	server.sendHTTPReply(c, payloadType, jsonStr)
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

	err = server.validator.RequireExecutorMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Bypass colonies controller and use the database directly, no need to synchronize this operation since files are inmutable
	var files []*core.File
	if msg.FileID != "" {
		file, err := server.db.GetFileByID(msg.ColonyID, msg.FileID)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
			server.handleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
		if file == nil {
			if server.handleHTTPError(c, errors.New("Failed to get file"), http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				server.handleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		}
		files = []*core.File{file}
	} else if msg.Prefix != "" && msg.Name != "" {
		if msg.Latest {
			files, err = server.db.GetLatestFileByName(msg.ColonyID, msg.Prefix, msg.Name)
			if server.handleHTTPError(c, err, http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				server.handleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		} else {
			files, err = server.db.GetFileByName(msg.ColonyID, msg.Prefix, msg.Name)
			if server.handleHTTPError(c, err, http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				server.handleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		}
	} else {
		if server.handleHTTPError(c, errors.New("malformat get file msg"), http.StatusInternalServerError) {
			log.Error(err)
			return
		}
	}

	if len(files) == 0 {
		if server.handleHTTPError(c, errors.New("Failed to get file"), http.StatusNotFound) {
			log.Error(err)
			return
		}
	} else {
		// This may not be strictly needed as the database lookup includes ColonyID
		// The reason is to prevent a user to correctly authenticate, but then obtain a file part of another colony
		for _, file := range files {
			if msg.ColonyID != file.ColonyID {
				if server.handleHTTPError(c, errors.New("msg.ColonyID missmatches file.ColonyID"), http.StatusForbidden) {
					log.Error(err)
					return
				}
			}
		}
	}

	jsonStr, err := core.ConvertFileArrayToJSON(files)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to converts files to json")
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

	err = server.validator.RequireExecutorMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	fileNames, err := server.db.GetFileNamesByPrefix(msg.ColonyID, msg.Prefix)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.Error(err)
		return
	}

	jsonBytes, err := json.Marshal(fileNames)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	server.sendHTTPReply(c, payloadType, string(jsonBytes))
}

func (server *ColoniesServer) handleGetFilePrefixesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
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

	err = server.validator.RequireExecutorMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	prefixes, err := server.db.GetFilePrefixes(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.Error(err)
		return
	}

	jsonBytes, err := json.Marshal(prefixes)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	server.sendHTTPReply(c, payloadType, string(jsonBytes))
}

func (server *ColoniesServer) handleDeleteFileHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteFileMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to delete file, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to delete file, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireExecutorMembership(recoveredID, msg.ColonyID, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	if msg.FileID != "" {
		err = server.db.DeleteFileByID(msg.ColonyID, msg.FileID)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			log.Error(err)
			return
		}
	} else if msg.Prefix != "" && msg.Name != "" {
		err = server.db.DeleteFileByName(msg.ColonyID, msg.Prefix, msg.Name)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			log.Error(err)
			return
		}
	} else {
		if server.handleHTTPError(c, errors.New("malformated delete file msg"), http.StatusBadRequest) {
			log.Error(err)
			return
		}
	}

	server.sendEmptyHTTPReply(c, payloadType)
}
