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

	err = server.validator.RequireMembership(recoveredID, msg.File.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Bypass colonies controller and use the database directly, no need to synchronize this operation since files are inmutable
	file := msg.File
	file.ID = core.GenerateRandomID()
	server.db.AddFile(msg.File)

	addedFile, err := server.db.GetFileByID(msg.File.ColonyName, file.ID)
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

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Bypass colonies controller and use the database directly, no need to synchronize this operation since files are inmutable
	var files []*core.File
	if msg.FileID != "" {
		file, err := server.db.GetFileByID(msg.ColonyName, msg.FileID)
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
	} else if msg.Label != "" && msg.Name != "" {
		if msg.Latest {
			files, err = server.db.GetLatestFileByName(msg.ColonyName, msg.Label, msg.Name)
			if server.handleHTTPError(c, err, http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				server.handleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		} else {
			files, err = server.db.GetFileByName(msg.ColonyName, msg.Label, msg.Name)
			if server.handleHTTPError(c, err, http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				server.handleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		}
	} else {
		if server.handleHTTPError(c, errors.New("malformatted get file msg"), http.StatusInternalServerError) {
			log.WithFields(log.Fields{"Error": err}).Debug("Malformatted get file msg")
			return
		}
	}

	if len(files) == 0 {
		if server.handleHTTPError(c, errors.New("Failed to get file"), http.StatusNotFound) {
			log.WithFields(log.Fields{"Error": err}).Debug("Failed to get files, len files is 0")
			return
		}
	} else {
		// This may not be strictly needed as the database lookup includes ColonyName
		// The reason is to prevent a user to correctly authenticate, but then obtain a file part of another colony
		for _, file := range files {
			if msg.ColonyName != file.ColonyName {
				if server.handleHTTPError(c, errors.New("msg.ColonyName missmatches file.ColonyName"), http.StatusForbidden) {
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

	log.WithFields(log.Fields{"FileID": msg.FileID, "Label": msg.Label, "Name": msg.Name, "Latest": msg.Latest}).Debug("Getting file")

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

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	fileDataArr, err := server.db.GetFileDataByLabel(msg.ColonyName, msg.Label)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.Error(err)
		return
	}

	jsonBytes, err := core.ConvertFileDataArrayToJSON(fileDataArr)
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	server.sendHTTPReply(c, payloadType, string(jsonBytes))
}

func (server *ColoniesServer) handleGetFileLabelsHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFileLabelsMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to get files, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to get files, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	var labels []*core.Label
	if msg.Name == "" {
		labels, err = server.db.GetFileLabels(msg.ColonyName)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			log.Error(err)
			return
		}
	} else {
		labels, err = server.db.GetFileLabelsByName(msg.ColonyName, msg.Name)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			log.Error(err)
			return
		}

	}

	jsonStr, err := core.ConvertLabelArrayToJSON(labels)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to converts files to json")
		server.handleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	server.sendHTTPReply(c, payloadType, jsonStr)
}

func (server *ColoniesServer) handleRemoveFileHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveFileMsgFromJSON(jsonString)
	if err != nil {
		if server.handleHTTPError(c, errors.New("Failed to remove file, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("Failed to remove file, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireMembership(recoveredID, msg.ColonyName, true)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.FileID != "" {
		err = server.db.RemoveFileByID(msg.ColonyName, msg.FileID)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	} else if msg.Label != "" && msg.Name != "" {
		err = server.db.RemoveFileByName(msg.ColonyName, msg.Label, msg.Name)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	} else {
		if server.handleHTTPError(c, errors.New("malformated remove file msg"), http.StatusBadRequest) {
			return
		}
	}

	server.sendEmptyHTTPReply(c, payloadType)
}
