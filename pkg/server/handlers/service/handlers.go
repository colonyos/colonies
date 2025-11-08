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
	ResourceDB() database.ResourceDatabase
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

// submitReconciliationFunc submits a reconciliation function spec based on the ResourceDefinition handler
// Returns the process ID of the created reconciliation process
func (h *Handlers) submitReconciliationFunc(reconciliation *core.Reconciliation, rd *core.ResourceDefinition) (string, error) {
	if rd == nil || rd.Spec.Handler.ExecutorType == "" || rd.Spec.Handler.FunctionName == "" {
		// No handler defined, skip reconciliation
		log.WithFields(log.Fields{
			"RD_IsNil":       rd == nil,
			"ExecutorType":   func() string { if rd != nil { return rd.Spec.Handler.ExecutorType }; return "" }(),
			"FunctionName":   func() string { if rd != nil { return rd.Spec.Handler.FunctionName }; return "" }(),
		}).Info("Skipping reconciliation - no handler defined")
		return "", nil
	}

	// Skip reconciliation if action is noop (no changes detected)
	if reconciliation.Action == core.ReconciliationNoop {
		log.WithFields(log.Fields{
			"ExecutorType": rd.Spec.Handler.ExecutorType,
			"FunctionName": rd.Spec.Handler.FunctionName,
		}).Info("Skipping reconciliation - no changes detected (noop)")
		return "", nil
	}

	log.WithFields(log.Fields{
		"ExecutorType":       rd.Spec.Handler.ExecutorType,
		"FunctionName":       rd.Spec.Handler.FunctionName,
		"Action":             reconciliation.Action,
		"ColonyName":         rd.Metadata.Namespace,
		"ReconciliationNil":  reconciliation == nil,
		"ReconciliationOld":  reconciliation.Old != nil,
		"ReconciliationNew":  reconciliation.New != nil,
	}).Info("Submitting reconciliation function")

	// Create a function spec with the reconciliation data
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ColonyName = rd.Metadata.Namespace
	funcSpec.Conditions.ExecutorType = rd.Spec.Handler.ExecutorType
	funcSpec.FuncName = rd.Spec.Handler.FunctionName
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
			"ExecutorType": rd.Spec.Handler.ExecutorType,
			"FunctionName": rd.Spec.Handler.FunctionName,
		}).Error("Failed to submit reconciliation function")
		return "", err
	}

	log.WithFields(log.Fields{
		"ProcessID":    addedProcess.ID,
		"ExecutorType": rd.Spec.Handler.ExecutorType,
		"FunctionName": rd.Spec.Handler.FunctionName,
		"Action":       reconciliation.Action,
	}).Info("Successfully submitted reconciliation function")

	return addedProcess.ID, nil
}

// RegisterHandlers implements the HandlerRegistrar interface
func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	// ResourceDefinition handlers
	if err := handlerRegistry.Register(rpc.AddResourceDefinitionPayloadType, h.HandleAddResourceDefinition); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetResourceDefinitionPayloadType, h.HandleGetResourceDefinition); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetResourceDefinitionsPayloadType, h.HandleGetResourceDefinitions); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveResourceDefinitionPayloadType, h.HandleRemoveResourceDefinition); err != nil {
		return err
	}

	// Service handlers
	if err := handlerRegistry.Register(rpc.AddResourcePayloadType, h.HandleAddResource); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetResourcePayloadType, h.HandleGetResource); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetResourcesPayloadType, h.HandleGetResources); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.UpdateResourcePayloadType, h.HandleUpdateResource); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveResourcePayloadType, h.HandleRemoveResource); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetResourceHistoryPayloadType, h.HandleGetResourceHistory); err != nil {
		return err
	}

	return nil
}

// HandleAddResourceDefinition - Only colony owner can add ResourceDefinitions
func (h *Handlers) HandleAddResourceDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddResourceDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add service definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add service definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ResourceDefinition == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add service definition, service definition is nil"), http.StatusBadRequest)
		return
	}

	// IMPORTANT: Only colony owner can add ResourceDefinitions
	// Namespace field holds the colony name
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ResourceDefinition.Metadata.Namespace)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Generate ID if not provided
	if msg.ResourceDefinition.ID == "" {
		msg.ResourceDefinition.ID = core.GenerateRandomID()
	}

	err = h.server.ResourceDB().AddResourceDefinition(msg.ResourceDefinition)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ID":         msg.ResourceDefinition.ID,
		"Name":       msg.ResourceDefinition.Metadata.Name,
		"Kind":       msg.ResourceDefinition.Kind,
		"ColonyName": msg.ResourceDefinition.Metadata.Namespace,
	}).Debug("Adding service definition")

	jsonString, err = msg.ResourceDefinition.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetResourceDefinition retrieves a ResourceDefinition by name
func (h *Handlers) HandleGetResourceDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetResourceDefinitionMsgFromJSON(jsonString)
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

	rd, err := h.server.ResourceDB().GetResourceDefinitionByName(msg.ColonyName, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if rd == nil {
		h.server.HandleHTTPError(c, errors.New("Service definition not found"), http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"Name":       msg.Name,
		"ColonyName": msg.ColonyName,
	}).Debug("Getting service definition")

	jsonString, err = rd.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetResourceDefinitions retrieves all ResourceDefinitions in a colony
func (h *Handlers) HandleGetResourceDefinitions(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetResourceDefinitionsMsgFromJSON(jsonString)
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

	rds, err := h.server.ResourceDB().GetResourceDefinitionsByNamespace(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName": msg.ColonyName,
		"Count":      len(rds),
	}).Debug("Getting service definitions")

	jsonString, err = core.ConvertResourceDefinitionArrayToJSON(rds)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleRemoveResourceDefinition removes a ResourceDefinition - Only colony owner can remove
func (h *Handlers) HandleRemoveResourceDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveResourceDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove service definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove service definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Only colony owner can remove ResourceDefinitions
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Get the ResourceDefinition to find its Kind
	rd, err := h.server.ResourceDB().GetResourceDefinitionByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if rd == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ResourceDefinition '%s' not found in namespace '%s'", msg.Name, msg.Namespace), http.StatusNotFound)
		return
	}

	// Check if there are any services using this ResourceDefinition
	services, err := h.server.ResourceDB().GetResourcesByNamespaceAndKind(msg.Namespace, rd.Spec.Names.Kind)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if len(services) > 0 {
		h.server.HandleHTTPError(c, fmt.Errorf("cannot remove ResourceDefinition '%s': %d service(s) of kind '%s' still exist", msg.Name, len(services), rd.Spec.Names.Kind), http.StatusConflict)
		return
	}

	err = h.server.ResourceDB().RemoveResourceDefinitionByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Removing service definition")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleAddResource adds a new Service instance
func (h *Handlers) HandleAddResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddResourceMsgFromJSON(jsonString)
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

	// Validate service against its ResourceDefinition schema
	// Service Kind is required
	if msg.Service.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("service kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all ResourceDefinitions in the namespace
	rds, err := h.server.ResourceDB().GetResourceDefinitionsByNamespace(msg.Service.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching ResourceDefinition - REQUIRED
	var matchedRD *core.ResourceDefinition
	for _, rd := range rds {
		if rd.Spec.Names.Kind == msg.Service.Kind {
			matchedRD = rd
			break
		}
	}

	// ResourceDefinition must exist
	if matchedRD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ResourceDefinition for kind '%s' not found in namespace '%s'", msg.Service.Kind, msg.Service.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedRD.Spec.Schema != nil {
		if err := core.ValidateResourceAgainstSchema(msg.Service, matchedRD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("service validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	err = h.server.ResourceDB().AddResource(msg.Service)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Save service history for create action
	history := core.CreateResourceHistory(msg.Service, recoveredID, "create")
	if err := h.server.ResourceDB().AddResourceHistory(history); err != nil {
		log.WithFields(log.Fields{"Error": err, "ResourceID": msg.Service.ID}).Warn("Failed to save service history")
		// Don't fail the request if history saving fails
	}

	log.WithFields(log.Fields{
		"ID":        msg.Service.ID,
		"Namespace": msg.Service.Metadata.Namespace,
		"Name":      msg.Service.Metadata.Name,
		"Kind":      msg.Service.Kind,
	}).Debug("Adding service")

	// Submit reconciliation function if handler is defined
	if matchedRD != nil {
		reconciliation := core.CreateReconciliation(nil, msg.Service)
		processID, err := h.submitReconciliationFunc(reconciliation, matchedRD)
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
			err = h.server.ResourceDB().UpdateResource(msg.Service)
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
					"ProcessID": processID,
				}).Warn("Failed to update service with reconciliation tracking info")
			} else {
				log.WithFields(log.Fields{
					"ProcessID": processID,
					"ResourceName": msg.Service.Metadata.Name,
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

// HandleGetResource retrieves a Service by namespace and name
func (h *Handlers) HandleGetResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetResourceMsgFromJSON(jsonString)
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

	service, err := h.server.ResourceDB().GetResourceByName(msg.Namespace, msg.Name)
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

// HandleGetResources retrieves Services by namespace and/or kind
func (h *Handlers) HandleGetResources(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetResourcesMsgFromJSON(jsonString)
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
		services, err = h.server.ResourceDB().GetResourcesByNamespace(msg.Namespace)
	} else {
		// Get services by namespace and kind
		services, err = h.server.ResourceDB().GetResourcesByNamespaceAndKind(msg.Namespace, msg.Kind)
	}

	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Kind":      msg.Kind,
		"Count":     len(services),
	}).Debug("Getting services")

	jsonString, err = core.ConvertResourceArrayToJSON(services)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleUpdateResource updates an existing Service
func (h *Handlers) HandleUpdateResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateUpdateResourceMsgFromJSON(jsonString)
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
	oldResource, err := h.server.ResourceDB().GetResourceByName(msg.Service.Metadata.Namespace, msg.Service.Metadata.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Validate service against its ResourceDefinition schema
	// Service Kind is required
	if msg.Service.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("service kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all ResourceDefinitions in the namespace
	rds, err := h.server.ResourceDB().GetResourceDefinitionsByNamespace(msg.Service.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching ResourceDefinition - REQUIRED
	var matchedRD *core.ResourceDefinition
	for _, rd := range rds {
		if rd.Spec.Names.Kind == msg.Service.Kind {
			matchedRD = rd
			break
		}
	}

	// ResourceDefinition must exist
	if matchedRD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ResourceDefinition for kind '%s' not found in namespace '%s'", msg.Service.Kind, msg.Service.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedRD.Spec.Schema != nil {
		if err := core.ValidateResourceAgainstSchema(msg.Service, matchedRD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("service validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Check if spec changed and increment generation if it did
	specChanged := false
	if oldResource != nil {
		reconciliation := core.CreateReconciliation(oldResource, msg.Service)
		if reconciliation.Diff != nil && len(reconciliation.Diff.SpecChanges) > 0 {
			// Spec changed, increment generation
			msg.Service.Metadata.Generation = oldResource.Metadata.Generation + 1
			specChanged = true
		} else {
			// Preserve old generation if spec didn't change
			msg.Service.Metadata.Generation = oldResource.Metadata.Generation
		}
	}

	err = h.server.ResourceDB().UpdateResource(msg.Service)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Save service history only if spec changed (not for status-only updates)
	if specChanged {
		history := core.CreateResourceHistory(msg.Service, recoveredID, "update")
		if err := h.server.ResourceDB().AddResourceHistory(history); err != nil {
			log.WithFields(log.Fields{"Error": err, "ResourceID": msg.Service.ID}).Warn("Failed to save service history")
			// Don't fail the request if history saving fails
		}
	}

	log.WithFields(log.Fields{
		"ID":        msg.Service.ID,
		"Namespace": msg.Service.Metadata.Namespace,
		"Name":      msg.Service.Metadata.Name,
	}).Debug("Updating service")

	// Submit reconciliation function only if spec changed (matchedRD already validated above)
	if specChanged {
		reconciliation := core.CreateReconciliation(oldResource, msg.Service)
		processID, err := h.submitReconciliationFunc(reconciliation, matchedRD)
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
			err = h.server.ResourceDB().UpdateResource(msg.Service)
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
					"ProcessID": processID,
				}).Warn("Failed to update service with reconciliation tracking info")
			} else {
				log.WithFields(log.Fields{
					"ProcessID": processID,
					"ResourceName": msg.Service.Metadata.Name,
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

// HandleRemoveResource removes a Service by namespace and name
func (h *Handlers) HandleRemoveResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveResourceMsgFromJSON(jsonString)
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

	err = h.server.ResourceDB().RemoveResourceByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Removing service")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleGetResourceHistory retrieves history for a service
func (h *Handlers) HandleGetResourceHistory(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetResourceHistoryMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get service history, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get service history, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Get the service to check permissions
	service, err := h.server.ResourceDB().GetResourceByID(msg.ResourceID)
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

	histories, err := h.server.ResourceDB().GetResourceHistory(msg.ResourceID, msg.Limit)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = core.ConvertResourceHistoryArrayToJSON(histories)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ResourceID": msg.ResourceID,
		"Limit":      msg.Limit,
		"Count":      len(histories),
	}).Debug("Getting service history")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}
