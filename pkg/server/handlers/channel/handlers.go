package channel

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/channel"
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
	ProcessDB() database.ProcessDatabase
	ChannelRouter() *channel.Router
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
	if err := handlerRegistry.Register(rpc.ChannelAppendPayloadType, h.HandleChannelAppend); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ChannelReadPayloadType, h.HandleChannelRead); err != nil {
		return err
	}
	return nil
}

// HandleChannelAppend handles appending a message to a channel
func (h *Handlers) HandleChannelAppend(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChannelAppendMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to append to channel, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to append to channel, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Get the process to verify membership
	process, err := h.server.ProcessDB().GetProcessByID(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Process not found"), http.StatusNotFound)
		return
	}

	// Verify colony membership
	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Get channel by process and name, creating on demand if necessary for cluster scenarios
	ch, err := h.server.ChannelRouter().GetByProcessAndName(msg.ProcessID, msg.Name)
	if err != nil {
		if err == channel.ErrChannelNotFound {
			// Try lazy creation - channel may not have replicated to this server yet
			ch, err = h.ensureChannelExists(process, msg.Name)
			if err != nil {
				if err == channel.ErrChannelNotFound {
					h.server.HandleHTTPError(c, errors.New("Channel not found"), http.StatusNotFound)
				} else {
					h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
				}
				return
			}
			log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name}).Info("Created channel on demand (cluster lazy creation)")
		} else {
			h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
	}

	// Determine caller ID - either submitter or executor
	callerID := getCallerID(recoveredID, process)

	log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name, "Sequence": msg.Sequence, "PayloadLen": len(msg.Payload), "CallerID": callerID}).Info("Appending to channel")

	// Append to channel with client-assigned sequence
	err = h.server.ChannelRouter().Append(ch.ID, callerID, msg.Sequence, msg.InReplyTo, msg.Payload)
	if err != nil {
		log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name, "Error": err}).Error("Failed to append to channel")
		if err == channel.ErrUnauthorized {
			h.server.HandleHTTPError(c, errors.New("Not authorized to write to channel"), http.StatusForbidden)
		} else {
			h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		}
		return
	}

	log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name, "Sequence": msg.Sequence}).Info("Successfully appended to channel")
	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleChannelRead handles reading messages from a channel
func (h *Handlers) HandleChannelRead(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateChannelReadMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to read from channel, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to read from channel, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Get the process to verify membership
	process, err := h.server.ProcessDB().GetProcessByID(msg.ProcessID)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		h.server.HandleHTTPError(c, errors.New("Process not found"), http.StatusNotFound)
		return
	}

	// Verify colony membership
	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		log.Error(err)
		return
	}

	// Get channel by process and name, creating on demand if necessary for cluster scenarios
	ch, err := h.server.ChannelRouter().GetByProcessAndName(msg.ProcessID, msg.Name)
	if err != nil {
		if err == channel.ErrChannelNotFound {
			// Try lazy creation - channel may not have replicated to this server yet
			ch, err = h.ensureChannelExists(process, msg.Name)
			if err != nil {
				if err == channel.ErrChannelNotFound {
					h.server.HandleHTTPError(c, errors.New("Channel not found"), http.StatusNotFound)
				} else {
					h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
				}
				return
			}
			log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name}).Info("Created channel on demand (cluster lazy creation)")
		} else {
			h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
			return
		}
	}

	// Determine caller ID - either submitter or executor
	callerID := getCallerID(recoveredID, process)

	// Read from channel
	entries, err := h.server.ChannelRouter().ReadAfter(ch.ID, callerID, msg.AfterSeq, msg.Limit)
	if err != nil {
		if err == channel.ErrUnauthorized {
			h.server.HandleHTTPError(c, errors.New("Not authorized to read from channel"), http.StatusForbidden)
		} else {
			h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return the entries
	jsonBytes, err := json.Marshal(entries)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name, "EntryCount": len(entries)}).Debug("Read from channel")
	h.server.SendHTTPReply(c, payloadType, string(jsonBytes))
}

// getCallerID determines the caller ID based on the recovered ID
// If the recovered ID matches the initiator, use submitter ID
// If it matches the assigned executor, use executor ID
func getCallerID(recoveredID string, process *core.Process) string {
	if recoveredID == process.InitiatorID {
		return process.InitiatorID
	}
	if recoveredID == process.AssignedExecutorID {
		return process.AssignedExecutorID
	}
	// Default to recovered ID (for user-based access)
	return recoveredID
}

// ensureChannelExists creates a channel on demand if it's defined in the process spec
// but doesn't exist locally. This handles cluster scenarios where a client connects
// to a different server than where the process was originally submitted.
func (h *Handlers) ensureChannelExists(process *core.Process, channelName string) (*channel.Channel, error) {
	// Check if this channel is defined in the process spec
	channelDefined := false
	for _, ch := range process.FunctionSpec.Channels {
		if ch == channelName {
			channelDefined = true
			break
		}
	}

	if !channelDefined {
		return nil, channel.ErrChannelNotFound
	}

	// Create the channel on demand
	ch := &channel.Channel{
		ID:          process.ID + "_" + channelName, // Deterministic ID
		ProcessID:   process.ID,
		Name:        channelName,
		SubmitterID: process.InitiatorID,
		ExecutorID:  process.AssignedExecutorID,
	}

	// Use CreateIfNotExists to handle concurrent creation
	if err := h.server.ChannelRouter().CreateIfNotExists(ch); err != nil {
		return nil, err
	}

	// Return the channel (might have been created by another goroutine)
	return h.server.ChannelRouter().GetByProcessAndName(process.ID, channelName)
}
