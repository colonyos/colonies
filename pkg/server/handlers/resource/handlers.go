package resource

import (
	"errors"
	"fmt"
	"net/http"

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
func (h *Handlers) submitReconciliationFunc(reconciliation *core.Reconciliation, rd *core.ResourceDefinition) error {
	if rd == nil || rd.Spec.Handler.ExecutorType == "" || rd.Spec.Handler.FunctionName == "" {
		// No handler defined, skip reconciliation
		log.WithFields(log.Fields{
			"RD_IsNil":       rd == nil,
			"ExecutorType":   func() string { if rd != nil { return rd.Spec.Handler.ExecutorType }; return "" }(),
			"FunctionName":   func() string { if rd != nil { return rd.Spec.Handler.FunctionName }; return "" }(),
		}).Info("Skipping reconciliation - no handler defined")
		return nil
	}

	// Skip reconciliation if action is noop (no changes detected)
	if reconciliation.Action == core.ReconciliationNoop {
		log.WithFields(log.Fields{
			"ExecutorType": rd.Spec.Handler.ExecutorType,
			"FunctionName": rd.Spec.Handler.FunctionName,
		}).Info("Skipping reconciliation - no changes detected (noop)")
		return nil
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
		return err
	}

	log.WithFields(log.Fields{
		"ProcessID":    addedProcess.ID,
		"ExecutorType": rd.Spec.Handler.ExecutorType,
		"FunctionName": rd.Spec.Handler.FunctionName,
		"Action":       reconciliation.Action,
	}).Info("Successfully submitted reconciliation function")

	return nil
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

	// Resource handlers
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

	return nil
}

// HandleAddResourceDefinition - Only colony owner can add ResourceDefinitions
func (h *Handlers) HandleAddResourceDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddResourceDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add resource definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add resource definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.ResourceDefinition == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add resource definition, resource definition is nil"), http.StatusBadRequest)
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
	}).Debug("Adding resource definition")

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
		h.server.HandleHTTPError(c, errors.New("Failed to get resource definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get resource definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view resource definitions
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
		h.server.HandleHTTPError(c, errors.New("Resource definition not found"), http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"Name":       msg.Name,
		"ColonyName": msg.ColonyName,
	}).Debug("Getting resource definition")

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
		h.server.HandleHTTPError(c, errors.New("Failed to get resource definitions, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get resource definitions, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view resource definitions
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
	}).Debug("Getting resource definitions")

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
		h.server.HandleHTTPError(c, errors.New("Failed to remove resource definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove resource definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
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

	// Check if there are any resources using this ResourceDefinition
	resources, err := h.server.ResourceDB().GetResourcesByNamespaceAndKind(msg.Namespace, rd.Spec.Names.Kind)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if len(resources) > 0 {
		h.server.HandleHTTPError(c, fmt.Errorf("cannot remove ResourceDefinition '%s': %d resource(s) of kind '%s' still exist", msg.Name, len(resources), rd.Spec.Names.Kind), http.StatusConflict)
		return
	}

	err = h.server.ResourceDB().RemoveResourceDefinitionByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Removing resource definition")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleAddResource adds a new Resource instance
func (h *Handlers) HandleAddResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddResourceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add resource, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add resource, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Resource == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add resource, resource is nil"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to add resources
	err = h.server.Validator().RequireMembership(recoveredID, msg.Resource.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Resource.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Generate ID if not provided
	if msg.Resource.ID == "" {
		msg.Resource.ID = core.GenerateRandomID()
	}

	// Validate resource against its ResourceDefinition schema
	// Resource Kind is required
	if msg.Resource.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("resource kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all ResourceDefinitions in the namespace
	rds, err := h.server.ResourceDB().GetResourceDefinitionsByNamespace(msg.Resource.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching ResourceDefinition - REQUIRED
	var matchedRD *core.ResourceDefinition
	for _, rd := range rds {
		if rd.Spec.Names.Kind == msg.Resource.Kind {
			matchedRD = rd
			break
		}
	}

	// ResourceDefinition must exist
	if matchedRD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ResourceDefinition for kind '%s' not found in namespace '%s'", msg.Resource.Kind, msg.Resource.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedRD.Spec.Schema != nil {
		if err := core.ValidateResourceAgainstSchema(msg.Resource, matchedRD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("resource validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	err = h.server.ResourceDB().AddResource(msg.Resource)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ID":        msg.Resource.ID,
		"Namespace": msg.Resource.Metadata.Namespace,
		"Name":      msg.Resource.Metadata.Name,
		"Kind":      msg.Resource.Kind,
	}).Debug("Adding resource")

	// Submit reconciliation function if handler is defined
	if matchedRD != nil {
		reconciliation := core.CreateReconciliation(nil, msg.Resource)
		err = h.submitReconciliationFunc(reconciliation, matchedRD)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Kind":  msg.Resource.Kind,
			}).Warn("Failed to submit reconciliation after resource add")
			// Don't fail the request if reconciliation submission fails
		}
	}

	jsonString, err = msg.Resource.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetResource retrieves a Resource by namespace and name
func (h *Handlers) HandleGetResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetResourceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get resource, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get resource, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view resources
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	resource, err := h.server.ResourceDB().GetResourceByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if resource == nil {
		h.server.HandleHTTPError(c, errors.New("Resource not found"), http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Getting resource")

	jsonString, err = resource.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetResources retrieves Resources by namespace and/or kind
func (h *Handlers) HandleGetResources(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetResourcesMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get resources, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get resources, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view resources
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	var resources []*core.Resource
	if msg.Kind == "" {
		// Get all resources in namespace
		resources, err = h.server.ResourceDB().GetResourcesByNamespace(msg.Namespace)
	} else {
		// Get resources by namespace and kind
		resources, err = h.server.ResourceDB().GetResourcesByNamespaceAndKind(msg.Namespace, msg.Kind)
	}

	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Kind":      msg.Kind,
		"Count":     len(resources),
	}).Debug("Getting resources")

	jsonString, err = core.ConvertResourceArrayToJSON(resources)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleUpdateResource updates an existing Resource
func (h *Handlers) HandleUpdateResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateUpdateResourceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to update resource, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to update resource, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Resource == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to update resource, resource is nil"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to update resources
	err = h.server.Validator().RequireMembership(recoveredID, msg.Resource.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Resource.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Get the old resource for reconciliation
	oldResource, err := h.server.ResourceDB().GetResourceByName(msg.Resource.Metadata.Namespace, msg.Resource.Metadata.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Validate resource against its ResourceDefinition schema
	// Resource Kind is required
	if msg.Resource.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("resource kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all ResourceDefinitions in the namespace
	rds, err := h.server.ResourceDB().GetResourceDefinitionsByNamespace(msg.Resource.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching ResourceDefinition - REQUIRED
	var matchedRD *core.ResourceDefinition
	for _, rd := range rds {
		if rd.Spec.Names.Kind == msg.Resource.Kind {
			matchedRD = rd
			break
		}
	}

	// ResourceDefinition must exist
	if matchedRD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("ResourceDefinition for kind '%s' not found in namespace '%s'", msg.Resource.Kind, msg.Resource.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedRD.Spec.Schema != nil {
		if err := core.ValidateResourceAgainstSchema(msg.Resource, matchedRD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("resource validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	err = h.server.ResourceDB().UpdateResource(msg.Resource)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ID":        msg.Resource.ID,
		"Namespace": msg.Resource.Metadata.Namespace,
		"Name":      msg.Resource.Metadata.Name,
	}).Debug("Updating resource")

	// Submit reconciliation function (matchedRD already validated above)
	reconciliation := core.CreateReconciliation(oldResource, msg.Resource)
	err = h.submitReconciliationFunc(reconciliation, matchedRD)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"Kind":  msg.Resource.Kind,
		}).Warn("Failed to submit reconciliation after resource update")
		// Don't fail the request if reconciliation submission fails
	}

	jsonString, err = msg.Resource.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleRemoveResource removes a Resource by namespace and name
func (h *Handlers) HandleRemoveResource(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveResourceMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove resource, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove resource, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to remove resources
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
	}).Debug("Removing resource")

	h.server.SendEmptyHTTPReply(c, payloadType)
}
