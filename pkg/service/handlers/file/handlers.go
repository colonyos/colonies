package file

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/service/registry"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ColoniesServer interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	Validator() security.Validator
	FileDB() database.FileDatabase
}

type Handlers struct {
	server ColoniesServer
}

func NewHandlers(server ColoniesServer) *Handlers {
	return &Handlers{server: server}
}

// RegisterHandlers implements the HandlerRegistrar interface
func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	if err := handlerRegistry.Register(rpc.AddFilePayloadType, h.HandleAddFile); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetFilePayloadType, h.HandleGetFile); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetFilesPayloadType, h.HandleGetFiles); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetFileLabelsPayloadType, h.HandleGetFileLabels); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveFilePayloadType, h.HandleRemoveFile); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddFile(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddFileMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add file, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.File == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add file, msg.File is nil"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.File.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Bypass colonies controller and use the database directly, no need to synchronize this operation since files are immutable
	file := msg.File
	file.ID = core.GenerateRandomID()
	h.server.FileDB().AddFile(msg.File)

	addedFile, err := h.server.FileDB().GetFileByID(msg.File.ColonyName, file.ID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	jsonStr, err := addedFile.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	log.WithFields(log.Fields{"FileID": file.ID}).Debug("Adding file")

	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleGetFile(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFileMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add log, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get file, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Bypass colonies controller and use the database directly, no need to synchronize this operation since files are immutable
	var files []*core.File
	if msg.FileID != "" {
		file, err := h.server.FileDB().GetFileByID(msg.ColonyName, msg.FileID)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
			h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
		if file == nil {
			if h.server.HandleHTTPError(c, errors.New("Failed to get file"), http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		}
		files = []*core.File{file}
	} else if msg.Label != "" && msg.Name != "" {
		if msg.Latest {
			files, err = h.server.FileDB().GetLatestFileByName(msg.ColonyName, msg.Label, msg.Name)
			if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		} else {
			files, err = h.server.FileDB().GetFileByName(msg.ColonyName, msg.Label, msg.Name)
			if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
				log.WithFields(log.Fields{"Error": err}).Debug("Failed to get file")
				h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
				return
			}
		}
	} else {
		if h.server.HandleHTTPError(c, errors.New("malformatted get file msg"), http.StatusInternalServerError) {
			log.WithFields(log.Fields{"Error": err}).Debug("Malformatted get file msg")
			return
		}
	}

	if len(files) == 0 {
		if h.server.HandleHTTPError(c, errors.New("Failed to get file"), http.StatusNotFound) {
			log.WithFields(log.Fields{"Error": err}).Debug("Failed to get files, len files is 0")
			return
		}
	} else {
		// This may not be strictly needed as the database lookup includes ColonyName
		// The reason is to prevent a user to correctly authenticate, but then obtain a file part of another colony
		for _, file := range files {
			if msg.ColonyName != file.ColonyName {
				if h.server.HandleHTTPError(c, errors.New("msg.ColonyName mismatches file.ColonyName"), http.StatusForbidden) {
					log.Error(err)
					return
				}
			}
		}
	}

	jsonStr, err := core.ConvertFileArrayToJSON(files)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to converts files to json")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{"FileID": msg.FileID, "Label": msg.Label, "Name": msg.Name, "Latest": msg.Latest}).Debug("Getting file")

	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleGetFiles(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFilesMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get files, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get files, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	fileDataArr, err := h.server.FileDB().GetFileDataByLabel(msg.ColonyName, msg.Label)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.Error(err)
		return
	}

	jsonBytes, err := core.ConvertFileDataArrayToJSON(fileDataArr)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.Error(err)
		return
	}

	h.server.SendHTTPReply(c, payloadType, string(jsonBytes))
}

func (h *Handlers) HandleGetFileLabels(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetFileLabelsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get files, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get files, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	var labels []*core.Label
	if msg.Name == "" {
		labels, err = h.server.FileDB().GetFileLabels(msg.ColonyName)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			log.Error(err)
			return
		}
	} else {
		labels, err = h.server.FileDB().GetFileLabelsByName(msg.ColonyName, msg.Name, msg.Exact)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			log.Error(err)
			return
		}
	}

	jsonStr, err := core.ConvertLabelArrayToJSON(labels)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Debug("Failed to converts files to json")
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleRemoveFile(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveFileMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove file, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove file, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.FileID != "" {
		err = h.server.FileDB().RemoveFileByID(msg.ColonyName, msg.FileID)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	} else if msg.Label != "" && msg.Name != "" {
		err = h.server.FileDB().RemoveFileByName(msg.ColonyName, msg.Label, msg.Name)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	} else {
		if h.server.HandleHTTPError(c, errors.New("malformatted remove file msg"), http.StatusBadRequest) {
			return
		}
	}

	h.server.SendEmptyHTTPReply(c, payloadType)
}