package file

import (
	"errors"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	Validator() security.Validator
	FileDB() database.FileDatabase
	// FileEventBus is the realtime publish/subscribe primitive used by
	// SubscribeFiles. Returning nil disables publishing — useful for
	// tests and for older deployments that haven't enabled the bus.
	// All publish call-sites in this file no-op when this returns nil.
	FileEventBus() backends.FileEventBus
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
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

func (h *Handlers) HandleAddFile(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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

	// Discriminate added vs updated by checking whether a previous
	// revision of (colony, label, name) already exists. We do this BEFORE
	// AddFile so the lookup doesn't see the new revision we're about to
	// add. A read error here is non-fatal: we fall back to "added" rather
	// than blocking the write on a transient lookup failure.
	priorRevisions, _ := h.server.FileDB().GetLatestFileByName(msg.File.ColonyName, msg.File.Label, msg.File.Name)
	isUpdate := len(priorRevisions) > 0

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

	// Publish realtime event after the DB write succeeded. The bus is
	// non-blocking and best-effort — a slow subscriber drops events
	// rather than stalling this handler.
	if bus := h.server.FileEventBus(); bus != nil {
		now := time.Now()
		var ev *core.FileEvent
		if isUpdate {
			ev = core.CreateFileUpdatedEvent(addedFile, now)
		} else {
			ev = core.CreateFileAddedEvent(addedFile, now)
		}
		bus.Publish(ev)
	}

	h.server.SendHTTPReply(c, payloadType, jsonStr)
}

func (h *Handlers) HandleGetFile(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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

func (h *Handlers) HandleGetFiles(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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

func (h *Handlers) HandleGetFileLabels(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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

func (h *Handlers) HandleRemoveFile(c backends.Context, recoveredID string, payloadType string, jsonString string) {
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

	// Resolve the file's identity BEFORE removing so the realtime event
	// has the correct (colony, label, name) coordinates. RemoveFileByName
	// already gives us those; RemoveFileByID needs a lookup first.
	var evLabel, evName string
	if msg.FileID != "" {
		// Best-effort lookup for the publish event. If the file is gone
		// or the lookup fails, we skip publishing rather than blocking
		// the remove.
		if file, lookupErr := h.server.FileDB().GetFileByID(msg.ColonyName, msg.FileID); lookupErr == nil && file != nil {
			evLabel = file.Label
			evName = file.Name
		}
		err = h.server.FileDB().RemoveFileByID(msg.ColonyName, msg.FileID)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	} else if msg.Label != "" && msg.Name != "" {
		evLabel = msg.Label
		evName = msg.Name
		err = h.server.FileDB().RemoveFileByName(msg.ColonyName, msg.Label, msg.Name)
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
	} else {
		if h.server.HandleHTTPError(c, errors.New("malformatted remove file msg"), http.StatusBadRequest) {
			return
		}
	}

	// Publish realtime event after the DB delete succeeded. Skip when
	// we couldn't resolve label+name (RemoveFileByID with stale lookup).
	if bus := h.server.FileEventBus(); bus != nil && evLabel != "" && evName != "" {
		bus.Publish(core.CreateFileRemovedEvent(msg.ColonyName, evLabel, evName, time.Now()))
	}

	h.server.SendEmptyHTTPReply(c, payloadType)
}