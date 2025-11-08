package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/handlers/process"
	"github.com/colonyos/colonies/pkg/server/registry"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	Validator() security.Validator
	ServiceDB() database.ServiceDatabase
	ProcessController() process.Controller
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{
		server: server,
	}
}

// submitReconciliationFunc submits a reconciliation function spec based on the ServiceDefinition handler
// Returns the process ID of the created reconciliation process
func (h *Handlers) submitReconciliationFunc(reconciliation *core.Reconciliation, sd *core.ServiceDefinition) (string, error) {
	if sd == nil || sd.Spec.Handler.ExecutorType == "" || sd.Spec.Handler.FunctionName == "" {
		// No handler defined, skip reconciliation
		log.WithFields(log.Fields{
			"SD_IsNil":       sd == nil,
			"ExecutorType":   func() string { if sd != nil { return sd.Spec.Handler.ExecutorType }; return "" }(),
			"FunctionName":   func() string { if sd != nil { return sd.Spec.Handler.FunctionName }; return "" }(),
		}).Info("Skipping reconciliation - no handler defined")
		return "", nil
	}

	// Skip reconciliation if action is noop (no changes detected)
	if reconciliation.Action == core.ReconciliationNoop {
		log.WithFields(log.Fields{
			"ExecutorType": sd.Spec.Handler.ExecutorType,
			"FunctionName": sd.Spec.Handler.FunctionName,
		}).Info("Skipping reconciliation - no changes detected (noop)")
		return "", nil
	}

	log.WithFields(log.Fields{
		"ExecutorType":       sd.Spec.Handler.ExecutorType,
		"FunctionName":       sd.Spec.Handler.FunctionName,
		"Action":             reconciliation.Action,
		"ColonyName":         sd.Metadata.Namespace,
		"ReconciliationNil":  reconciliation == nil,
		"ReconciliationOld":  reconciliation.Old != nil,
		"ReconciliationNew":  reconciliation.New != nil,
	}).Info("Submitting reconciliation function")

	// Create a function spec with the reconciliation data
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ColonyName = sd.Metadata.Namespace
	funcSpec.Conditions.ExecutorType = sd.Spec.Handler.ExecutorType
	funcSpec.FuncName = sd.Spec.Handler.FunctionName
	funcSpec.Reconciliation = reconciliation

	log.WithFields(log.Fields{
		"FuncSpecReconciliationNil": funcSpec.Reconciliation == nil,
	}).Info("Set reconciliation on FunctionSpec")

	// Create process from function spec
	process := core.CreateProcess(funcSpec)

	log.WithFields(log.Fields{
		"ProcessReconciliationNil": process.FunctionSpec.Reconciliation == nil,
	}).Info("Created process from FunctionSpec")

	// Submit the process
	addedProcess, err := h.server.ProcessController().AddProcess(process)
	if err != nil {
		log.WithFields(log.Fields{
			"Error":        err,
			"ExecutorType": sd.Spec.Handler.ExecutorType,
			"FunctionName": sd.Spec.Handler.FunctionName,
		}).Error("Failed to submit reconciliation function")
		return "", err
	}

	log.WithFields(log.Fields{
		"ProcessID":    addedProcess.ID,
		"ExecutorType": sd.Spec.Handler.ExecutorType,
		"FunctionName": sd.Spec.Handler.FunctionName,
		"Action":       reconciliation.Action,
	}).Info("Successfully submitted reconciliation function")

	return addedProcess.ID, nil
}

// RegisterHandlers implements the HandlerRegistrar interface
func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	// ServiceDefinition handlers
	if err := handlerRegistry.Register(rpc.AddServiceDefinitionPayloadType, h.HandleAddServiceDefinition); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetServiceDefinitionPayloadType, h.HandleGetServiceDefinition); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetServiceDefinitionsPayloadType, h.HandleGetServiceDefinitions); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveServiceDefinitionPayloadType, h.HandleRemoveServiceDefinition); err != nil {
		return err
	}

	// Service handlers
	if err := handlerRegistry.Register(rpc.AddServicePayloadType, h.HandleAddService); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetServicePayloadType, h.HandleGetService); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetServicesPayloadType, h.HandleGetServices); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.UpdateServicePayloadType, h.HandleUpdateService); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveServicePayloadType, h.HandleRemoveService); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetServiceHistoryPayloadType, h.HandleGetServiceHistory); err != nil {
		return err
	}

	return nil
}

// HandleAddServiceDefinition - Only colony owner can add ServiceDefinitions
func (h *Handlers) HandleAddServiceDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddServiceDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add service definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add service definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ServiceDefinition == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add service definition, service definition is nil"), http.StatusBadRequest)
		return
	}

	// IMPORTANT: Only colony owner can add ServiceDefinitions
	// Namespace field holds the colony name
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ServiceDefinition.Metadata.Namespace)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Generate ID if not provided
	if msg.ServiceDefinition.ID == "" {
		msg.ServiceDefinition.ID = core.GenerateRandomID()
	}

	err = h.server.ServiceDB().AddServiceDefinition(msg.ServiceDefinition)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ID":         msg.ServiceDefinition.ID,
		"Name":       msg.ServiceDefinition.Metadata.Name,
		"Kind":       msg.ServiceDefinition.Kind,
		"ColonyName": msg.ServiceDefinition.Metadata.Namespace,
	}).Debug("Adding service definition")

	jsonString, err = msg.ServiceDefinition.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetServiceDefinition retrieves a ServiceDefinition by name
func (h *Handlers) HandleGetServiceDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetServiceDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get service definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get service definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view service definitions
	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	sd, err := h.server.ServiceDB().GetServiceDefinitionByName(msg.ColonyName, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if sd == nil {
		h.server.HandleHTTPError(c, errors.New("Service definition not found"), http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"Name":       msg.Name,
		"ColonyName": msg.ColonyName,
	}).Debug("Getting service definition")

	jsonString, err = sd.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetServiceDefinitions retrieves all ServiceDefinitions in a colony
func (h *Handlers) HandleGetServiceDefinitions(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetServiceDefinitionsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get service definitions, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get service definitions, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view service definitions
	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	sds, err := h.server.ServiceDB().GetServiceDefinitionsByNamespace(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName": msg.ColonyName,
		"Count":      len(sds),
	}).Debug("Getting service definitions")

	jsonString, err = core.ConvertServiceDefinitionArrayToJSON(sds)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleRemoveServiceDefinition removes a ServiceDefinition - Only colony owner can remove
func (h *Handlers) HandleRemoveServiceDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveServiceDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove service definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove service definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Only colony owner can remove ServiceDefinitions
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Get the ServiceDefinition to find its Kind
	sd, err := h.server.ServiceDB().GetServiceDefinitionByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if sd == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ServiceDefinition '%s' not found in namespace '%s'", msg.Name, msg.Namespace), http.StatusNotFound)
		return
	}

	// Check if there are any services using this ServiceDefinition
	services, err := h.server.ServiceDB().GetServicesByNamespaceAndKind(msg.Namespace, sd.Spec.Names.Kind)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if len(services) > 0 {
		h.server.HandleHTTPError(c, fmt.Errorf("cannot remove ServiceDefinition '%s': %d service(s) of kind '%s' still exist", msg.Name, len(services), sd.Spec.Names.Kind), http.StatusConflict)
		return
	}

	err = h.server.ServiceDB().RemoveServiceDefinitionByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Removing service definition")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleAddService adds a new Service instance
func (h *Handlers) HandleAddService(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddServiceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add service, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add service, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Service == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add service, service is nil"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to add services
	err = h.server.Validator().RequireMembership(recoveredID, msg.Service.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Service.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Generate ID if not provided
	if msg.Service.ID == "" {
		msg.Service.ID = core.GenerateRandomID()
	}

	// Validate service against its ServiceDefinition schema
	// Service Kind is required
	if msg.Service.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("service kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all ServiceDefinitions in the namespace
	sds, err := h.server.ServiceDB().GetServiceDefinitionsByNamespace(msg.Service.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching ServiceDefinition - REQUIRED
	var matchedSD *core.ServiceDefinition
	for _, sd := range sds {
		if sd.Spec.Names.Kind == msg.Service.Kind {
			matchedSD = sd
			break
		}
	}

	// ServiceDefinition must exist
	if matchedSD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ServiceDefinition for kind '%s' not found in namespace '%s'", msg.Service.Kind, msg.Service.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedSD.Spec.Schema != nil {
		if err := core.ValidateServiceAgainstSchema(msg.Service, matchedSD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("service validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	err = h.server.ServiceDB().AddService(msg.Service)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Save service history for create action
	history := core.CreateServiceHistory(msg.Service, recoveredID, "create")
	if err := h.server.ServiceDB().AddServiceHistory(history); err != nil {
		log.WithFields(log.Fields{"Error": err, "ServiceID": msg.Service.ID}).Warn("Failed to save service history")
		// Don't fail the request if history saving fails
	}

	log.WithFields(log.Fields{
		"ID":        msg.Service.ID,
		"Namespace": msg.Service.Metadata.Namespace,
		"Name":      msg.Service.Metadata.Name,
		"Kind":      msg.Service.Kind,
	}).Debug("Adding service")

	// Submit reconciliation function if handler is defined
	if matchedSD != nil {
		reconciliation := core.CreateReconciliation(nil, msg.Service)
		processID, err := h.submitReconciliationFunc(reconciliation, matchedSD)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Kind":  msg.Service.Kind,
			}).Warn("Failed to submit reconciliation after service add")
			// Don't fail the request if reconciliation submission fails
		} else if processID != "" {
			// Update service with reconciliation tracking info
			msg.Service.Metadata.LastReconciliationProcess = processID
			msg.Service.Metadata.LastReconciliationTime = time.Now()
			err = h.server.ServiceDB().UpdateService(msg.Service)
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
					"ProcessID": processID,
				}).Warn("Failed to update service with reconciliation tracking info")
			} else {
				log.WithFields(log.Fields{
					"ProcessID": processID,
					"ServiceName": msg.Service.Metadata.Name,
				}).Info("Updated service with reconciliation tracking info")
			}
		}
	}

	jsonString, err = msg.Service.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetService retrieves a Service by namespace and name
func (h *Handlers) HandleGetService(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetServiceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get service, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get service, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view services
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	service, err := h.server.ServiceDB().GetServiceByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if service == nil {
		h.server.HandleHTTPError(c, errors.New("Service not found"), http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Getting service")

	jsonString, err = service.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetServices retrieves Services by namespace and/or kind
func (h *Handlers) HandleGetServices(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetServicesMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get services, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get services, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view services
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	var services []*core.Service
	if msg.Kind == "" {
		// Get all services in namespace
		services, err = h.server.ServiceDB().GetServicesByNamespace(msg.Namespace)
	} else {
		// Get services by namespace and kind
		services, err = h.server.ServiceDB().GetServicesByNamespaceAndKind(msg.Namespace, msg.Kind)
	}

	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Kind":      msg.Kind,
		"Count":     len(services),
	}).Debug("Getting services")

	jsonString, err = core.ConvertServiceArrayToJSON(services)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleUpdateService updates an existing Service
func (h *Handlers) HandleUpdateService(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateUpdateServiceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to update service, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to update service, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Service == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to update service, service is nil"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to update services
	err = h.server.Validator().RequireMembership(recoveredID, msg.Service.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Service.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Get the old service for reconciliation
	oldService, err := h.server.ServiceDB().GetServiceByName(msg.Service.Metadata.Namespace, msg.Service.Metadata.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Validate service against its ServiceDefinition schema
	// Service Kind is required
	if msg.Service.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("service kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all ServiceDefinitions in the namespace
	sds, err := h.server.ServiceDB().GetServiceDefinitionsByNamespace(msg.Service.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching ServiceDefinition - REQUIRED
	var matchedSD *core.ServiceDefinition
	for _, sd := range sds {
		if sd.Spec.Names.Kind == msg.Service.Kind {
			matchedSD = sd
			break
		}
	}

	// ServiceDefinition must exist
	if matchedSD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ServiceDefinition for kind '%s' not found in namespace '%s'", msg.Service.Kind, msg.Service.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedSD.Spec.Schema != nil {
		if err := core.ValidateServiceAgainstSchema(msg.Service, matchedSD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("service validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Check if spec changed and increment generation if it did
	specChanged := false
	if oldService != nil {
		// Preserve the ID from the existing service
		msg.Service.ID = oldService.ID

		reconciliation := core.CreateReconciliation(oldService, msg.Service)
		if reconciliation.Diff != nil && len(reconciliation.Diff.SpecChanges) > 0 {
			// Spec changed, increment generation
			msg.Service.Metadata.Generation = oldService.Metadata.Generation + 1
			specChanged = true
		} else {
			// Preserve old generation if spec didn't change
			msg.Service.Metadata.Generation = oldService.Metadata.Generation
		}
	}

	err = h.server.ServiceDB().UpdateService(msg.Service)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Save service history only if spec changed (not for status-only updates)
	if specChanged {
		history := core.CreateServiceHistory(msg.Service, recoveredID, "update")
		if err := h.server.ServiceDB().AddServiceHistory(history); err != nil {
			log.WithFields(log.Fields{"Error": err, "ServiceID": msg.Service.ID}).Warn("Failed to save service history")
			// Don't fail the request if history saving fails
		}
	}

	log.WithFields(log.Fields{
		"ID":        msg.Service.ID,
		"Namespace": msg.Service.Metadata.Namespace,
		"Name":      msg.Service.Metadata.Name,
	}).Debug("Updating service")

	// Always submit reconciliation if handler is defined
	// The reconciler will determine if actual changes are needed via the Diff
	// This ensures reconciliation happens even when setting the same value
	// (useful for recovering from inconsistent state)
	if matchedSD != nil && matchedSD.Spec.Handler.ExecutorType != "" {
		reconciliation := core.CreateReconciliation(oldService, msg.Service)
		processID, err := h.submitReconciliationFunc(reconciliation, matchedSD)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Kind":  msg.Service.Kind,
			}).Warn("Failed to submit reconciliation after service update")
			// Don't fail the request if reconciliation submission fails
		} else if processID != "" {
			// Update service with reconciliation tracking info
			msg.Service.Metadata.LastReconciliationProcess = processID
			msg.Service.Metadata.LastReconciliationTime = time.Now()
			err = h.server.ServiceDB().UpdateService(msg.Service)
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
					"ProcessID": processID,
				}).Warn("Failed to update service with reconciliation tracking info")
			} else {
				log.WithFields(log.Fields{
					"ProcessID": processID,
					"ServiceName": msg.Service.Metadata.Name,
				}).Info("Updated service with reconciliation tracking info")
			}
		}
	}

	jsonString, err = msg.Service.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleRemoveService removes a Service by namespace and name
func (h *Handlers) HandleRemoveService(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveServiceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove service, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove service, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to remove services
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	err = h.server.ServiceDB().RemoveServiceByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Removing service")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleGetServiceHistory retrieves history for a service
func (h *Handlers) HandleGetServiceHistory(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetServiceHistoryMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get service history, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get service history, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Get the service to check permissions
	service, err := h.server.ServiceDB().GetServiceByID(msg.ServiceID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}
	if service == nil {
		h.server.HandleHTTPError(c, errors.New("Service not found"), http.StatusNotFound)
		return
	}

	// Require membership or colony owner to view history
	err = h.server.Validator().RequireMembership(recoveredID, service.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, service.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	histories, err := h.server.ServiceDB().GetServiceHistory(msg.ServiceID, msg.Limit)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = core.ConvertServiceHistoryArrayToJSON(histories)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ServiceID": msg.ServiceID,
		"Limit":     msg.Limit,
		"Count":     len(histories),
	}).Debug("Getting service history")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}
