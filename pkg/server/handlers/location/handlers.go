package location

import (
	"errors"
	"net/http"

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
	GetLocationDB() database.LocationDatabase
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

func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	if err := handlerRegistry.Register(rpc.AddLocationPayloadType, h.HandleAddLocation); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetLocationsPayloadType, h.HandleGetLocations); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetLocationPayloadType, h.HandleGetLocation); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveLocationPayloadType, h.HandleRemoveLocation); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleAddLocation(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddLocationMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to add location, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add location, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Location == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add location, location is nil"), http.StatusBadRequest)
		return
	}

	colony, err := h.server.GetColonyDB().GetColonyByName(msg.Location.ColonyName)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to resolve colony name"), http.StatusBadRequest) {
			return
		}
	}

	if colony == nil {
		if h.server.HandleHTTPError(c, errors.New("Colony with name <"+msg.Location.ColonyName+"> does not exists"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetValidator().RequireColonyOwner(recoveredID, colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	locationExist, err := h.server.GetLocationDB().GetLocationByName(msg.Location.ColonyName, msg.Location.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if locationExist != nil {
		if h.server.HandleHTTPError(c, errors.New("A location with name <"+msg.Location.Name+"> already exists in Colony with name <"+msg.Location.ColonyName+">"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetLocationDB().AddLocation(msg.Location)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	addedLocation, err := h.server.GetLocationDB().GetLocationByName(colony.Name, msg.Location.Name)
	if addedLocation == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add location, addedLocation is nil"), http.StatusInternalServerError)
		return
	}

	jsonString, err = addedLocation.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": addedLocation.ColonyName, "Name": addedLocation.Name, "LocationID": addedLocation.ID}).Debug("Adding location")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetLocations(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetLocationsMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get locations, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get locations, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	locations, err := h.server.GetLocationDB().GetLocationsByColonyName(colony.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertLocationArrayToJSON(locations)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyId": colony.ID}).Debug("Getting locations")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetLocation(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetLocationMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to get location, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get location, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	location, err := h.server.GetLocationDB().GetLocationByName(msg.ColonyName, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if location == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get location, location is nil"), http.StatusNotFound)
		return
	}

	jsonString, err = location.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "Name": msg.Name}).Debug("Getting location")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRemoveLocation(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveLocationMsgFromJSON(jsonString)
	if err != nil {
		if h.server.HandleHTTPError(c, errors.New("Failed to remove location, invalid JSON"), http.StatusBadRequest) {
			return
		}
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove location, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	location, err := h.server.GetLocationDB().GetLocationByName(msg.ColonyName, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	if location == nil {
		if h.server.HandleHTTPError(c, errors.New("Location with name <"+msg.Name+"> not found"), http.StatusBadRequest) {
			return
		}
	}

	err = h.server.GetLocationDB().RemoveLocationByName(colony.Name, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "Name": msg.Name}).Debug("Removing location")

	h.server.SendEmptyHTTPReply(c, payloadType)
}
