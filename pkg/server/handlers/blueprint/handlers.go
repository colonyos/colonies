package blueprint

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
	BlueprintDB() database.BlueprintDatabase
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

// submitReconciliationFunc submits a reconciliation function spec based on the BlueprintDefinition handler
// Returns the process ID of the created reconciliation process
func (h *Handlers) submitReconciliationFunc(reconciliation *core.Reconciliation, sd *core.BlueprintDefinition) (string, error) {
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

	// Check if blueprint instance specifies executor targeting
	// Priority: 1) root handler field, 2) spec fields (backward compatibility)
	// This allows targeting specific reconciler nodes (e.g., edge-node raspberry pi)
	// while keeping blueprint definitions generic
	if reconciliation.New != nil {
		var targetExecutors []string

		// Check root-level handler first (preferred, cleaner separation)
		if reconciliation.New.Handler != nil {
			if len(reconciliation.New.Handler.ExecutorNames) > 0 {
				targetExecutors = reconciliation.New.Handler.ExecutorNames
				log.WithFields(log.Fields{
					"ExecutorNames":  targetExecutors,
					"BlueprintName":  reconciliation.New.Metadata.Name,
					"Source":         "handler.executorNames",
				}).Info("Targeting specific executors for reconciliation")
			} else if reconciliation.New.Handler.ExecutorName != "" {
				targetExecutors = []string{reconciliation.New.Handler.ExecutorName}
				log.WithFields(log.Fields{
					"ExecutorName":   reconciliation.New.Handler.ExecutorName,
					"BlueprintName":  reconciliation.New.Metadata.Name,
					"Source":         "handler.executorName",
				}).Info("Targeting specific executor for reconciliation")
			}
		}

		// Fall back to spec fields for backward compatibility
		if len(targetExecutors) == 0 && reconciliation.New.Spec != nil {
			if executorNames, ok := reconciliation.New.Spec["executorNames"].([]interface{}); ok {
				// Convert []interface{} to []string
				for _, name := range executorNames {
					if nameStr, ok := name.(string); ok {
						targetExecutors = append(targetExecutors, nameStr)
					}
				}
				if len(targetExecutors) > 0 {
					log.WithFields(log.Fields{
						"ExecutorNames":  targetExecutors,
						"BlueprintName":  reconciliation.New.Metadata.Name,
						"Source":         "spec.executorNames (deprecated)",
					}).Info("Targeting specific executors for reconciliation")
				}
			} else if executorName, ok := reconciliation.New.Spec["executorName"].(string); ok {
				targetExecutors = []string{executorName}
				log.WithFields(log.Fields{
					"ExecutorName":   executorName,
					"BlueprintName":  reconciliation.New.Metadata.Name,
					"Source":         "spec.executorName (deprecated)",
				}).Info("Targeting specific executor for reconciliation")
			}
		}

		// Apply targeting if found
		if len(targetExecutors) > 0 {
			funcSpec.Conditions.ExecutorNames = targetExecutors
		}
	}

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
	// BlueprintDefinition handlers
	if err := handlerRegistry.Register(rpc.AddBlueprintDefinitionPayloadType, h.HandleAddBlueprintDefinition); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetBlueprintDefinitionPayloadType, h.HandleGetBlueprintDefinition); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetBlueprintDefinitionsPayloadType, h.HandleGetBlueprintDefinitions); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveBlueprintDefinitionPayloadType, h.HandleRemoveBlueprintDefinition); err != nil {
		return err
	}

	// Blueprint handlers
	if err := handlerRegistry.Register(rpc.AddBlueprintPayloadType, h.HandleAddBlueprint); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetBlueprintPayloadType, h.HandleGetBlueprint); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetBlueprintsPayloadType, h.HandleGetBlueprints); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.UpdateBlueprintPayloadType, h.HandleUpdateBlueprint); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveBlueprintPayloadType, h.HandleRemoveBlueprint); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetBlueprintHistoryPayloadType, h.HandleGetBlueprintHistory); err != nil {
		return err
	}

	return nil
}

// HandleAddBlueprintDefinition - Only colony owner can add BlueprintDefinitions
func (h *Handlers) HandleAddBlueprintDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddBlueprintDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add blueprint definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add blueprint definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.BlueprintDefinition == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add blueprint definition, blueprint definition is nil"), http.StatusBadRequest)
		return
	}

	// IMPORTANT: Only colony owner can add BlueprintDefinitions
	// Namespace field holds the colony name
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.BlueprintDefinition.Metadata.Namespace)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Generate ID if not provided
	if msg.BlueprintDefinition.ID == "" {
		msg.BlueprintDefinition.ID = core.GenerateRandomID()
	}

	err = h.server.BlueprintDB().AddBlueprintDefinition(msg.BlueprintDefinition)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ID":         msg.BlueprintDefinition.ID,
		"Name":       msg.BlueprintDefinition.Metadata.Name,
		"Kind":       msg.BlueprintDefinition.Kind,
		"ColonyName": msg.BlueprintDefinition.Metadata.Namespace,
	}).Debug("Adding blueprint definition")

	jsonString, err = msg.BlueprintDefinition.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetBlueprintDefinition retrieves a BlueprintDefinition by name
func (h *Handlers) HandleGetBlueprintDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetBlueprintDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view blueprint definitions
	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	sd, err := h.server.BlueprintDB().GetBlueprintDefinitionByName(msg.ColonyName, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if sd == nil {
		h.server.HandleHTTPError(c, errors.New("Blueprint definition not found"), http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"Name":       msg.Name,
		"ColonyName": msg.ColonyName,
	}).Debug("Getting blueprint definition")

	jsonString, err = sd.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetBlueprintDefinitions retrieves all BlueprintDefinitions in a colony
func (h *Handlers) HandleGetBlueprintDefinitions(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetBlueprintDefinitionsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint definitions, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint definitions, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view blueprint definitions
	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.ColonyName)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	sds, err := h.server.BlueprintDB().GetBlueprintDefinitionsByNamespace(msg.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName": msg.ColonyName,
		"Count":      len(sds),
	}).Debug("Getting blueprint definitions")

	jsonString, err = core.ConvertBlueprintDefinitionArrayToJSON(sds)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleRemoveBlueprintDefinition removes a BlueprintDefinition - Only colony owner can remove
func (h *Handlers) HandleRemoveBlueprintDefinition(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveBlueprintDefinitionMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove blueprint definition, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove blueprint definition, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Only colony owner can remove BlueprintDefinitions
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Get the BlueprintDefinition to find its Kind
	sd, err := h.server.BlueprintDB().GetBlueprintDefinitionByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if sd == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("BlueprintDefinition '%s' not found in namespace '%s'", msg.Name, msg.Namespace), http.StatusNotFound)
		return
	}

	// Check if there are any blueprints using this BlueprintDefinition
	blueprints, err := h.server.BlueprintDB().GetBlueprintsByNamespaceAndKind(msg.Namespace, sd.Spec.Names.Kind)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if len(blueprints) > 0 {
		h.server.HandleHTTPError(c, fmt.Errorf("cannot remove BlueprintDefinition '%s': %d blueprint(s) of kind '%s' still exist", msg.Name, len(blueprints), sd.Spec.Names.Kind), http.StatusConflict)
		return
	}

	err = h.server.BlueprintDB().RemoveBlueprintDefinitionByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Removing blueprint definition")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleAddBlueprint adds a new Blueprint instance
func (h *Handlers) HandleAddBlueprint(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddBlueprintMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add blueprint, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to add blueprint, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Blueprint == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to add blueprint, blueprint is nil"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to add blueprints
	err = h.server.Validator().RequireMembership(recoveredID, msg.Blueprint.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Blueprint.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Generate ID if not provided
	if msg.Blueprint.ID == "" {
		msg.Blueprint.ID = core.GenerateRandomID()
	}

	// Validate blueprint against its BlueprintDefinition schema
	// Blueprint Kind is required
	if msg.Blueprint.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("blueprint kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all BlueprintDefinitions in the namespace
	sds, err := h.server.BlueprintDB().GetBlueprintDefinitionsByNamespace(msg.Blueprint.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching BlueprintDefinition - REQUIRED
	var matchedSD *core.BlueprintDefinition
	for _, sd := range sds {
		if sd.Spec.Names.Kind == msg.Blueprint.Kind {
			matchedSD = sd
			break
		}
	}

	// BlueprintDefinition must exist
	if matchedSD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("BlueprintDefinition for kind '%s' not found in namespace '%s'", msg.Blueprint.Kind, msg.Blueprint.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedSD.Spec.Schema != nil {
		if err := core.ValidateBlueprintAgainstSchema(msg.Blueprint, matchedSD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("blueprint validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	err = h.server.BlueprintDB().AddBlueprint(msg.Blueprint)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Save blueprint history for create action
	history := core.CreateBlueprintHistory(msg.Blueprint, recoveredID, "create")
	if err := h.server.BlueprintDB().AddBlueprintHistory(history); err != nil {
		log.WithFields(log.Fields{"Error": err, "BlueprintID": msg.Blueprint.ID}).Warn("Failed to save blueprint history")
		// Don't fail the request if history saving fails
	}

	log.WithFields(log.Fields{
		"ID":        msg.Blueprint.ID,
		"Namespace": msg.Blueprint.Metadata.Namespace,
		"Name":      msg.Blueprint.Metadata.Name,
		"Kind":      msg.Blueprint.Kind,
	}).Debug("Adding blueprint")

	// Submit reconciliation function if handler is defined
	if matchedSD != nil {
		reconciliation := core.CreateReconciliation(nil, msg.Blueprint)
		processID, err := h.submitReconciliationFunc(reconciliation, matchedSD)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Kind":  msg.Blueprint.Kind,
			}).Warn("Failed to submit reconciliation after blueprint add")
			// Don't fail the request if reconciliation submission fails
		} else if processID != "" {
			// Update blueprint with reconciliation tracking info
			msg.Blueprint.Metadata.LastReconciliationProcess = processID
			msg.Blueprint.Metadata.LastReconciliationTime = time.Now()
			err = h.server.BlueprintDB().UpdateBlueprint(msg.Blueprint)
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
					"ProcessID": processID,
				}).Warn("Failed to update blueprint with reconciliation tracking info")
			} else {
				log.WithFields(log.Fields{
					"ProcessID": processID,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Info("Updated blueprint with reconciliation tracking info")
			}
		}
	}

	jsonString, err = msg.Blueprint.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetBlueprint retrieves a Blueprint by namespace and name
func (h *Handlers) HandleGetBlueprint(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetBlueprintMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view blueprints
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	blueprint, err := h.server.BlueprintDB().GetBlueprintByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if blueprint == nil {
		h.server.HandleHTTPError(c, errors.New("Blueprint not found"), http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Getting blueprint")

	jsonString, err = blueprint.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleGetBlueprints retrieves Blueprints by namespace and/or kind
func (h *Handlers) HandleGetBlueprints(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetBlueprintsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprints, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprints, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to view blueprints
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	var blueprints []*core.Blueprint
	if msg.Kind == "" {
		// Get all blueprints in namespace
		blueprints, err = h.server.BlueprintDB().GetBlueprintsByNamespace(msg.Namespace)
	} else {
		// Get blueprints by namespace and kind
		blueprints, err = h.server.BlueprintDB().GetBlueprintsByNamespaceAndKind(msg.Namespace, msg.Kind)
	}

	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Kind":      msg.Kind,
		"Count":     len(blueprints),
	}).Debug("Getting blueprints")

	jsonString, err = core.ConvertBlueprintArrayToJSON(blueprints)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleUpdateBlueprint updates an existing Blueprint
func (h *Handlers) HandleUpdateBlueprint(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateUpdateBlueprintMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to update blueprint, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to update blueprint, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	if msg.Blueprint == nil {
		h.server.HandleHTTPError(c, errors.New("Failed to update blueprint, blueprint is nil"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to update blueprints
	err = h.server.Validator().RequireMembership(recoveredID, msg.Blueprint.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Blueprint.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Get the old blueprint for reconciliation
	oldBlueprint, err := h.server.BlueprintDB().GetBlueprintByName(msg.Blueprint.Metadata.Namespace, msg.Blueprint.Metadata.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Validate blueprint against its BlueprintDefinition schema
	// Blueprint Kind is required
	if msg.Blueprint.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("blueprint kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all BlueprintDefinitions in the namespace
	sds, err := h.server.BlueprintDB().GetBlueprintDefinitionsByNamespace(msg.Blueprint.Metadata.Namespace)
	if err != nil {
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Find matching BlueprintDefinition - REQUIRED
	var matchedSD *core.BlueprintDefinition
	for _, sd := range sds {
		if sd.Spec.Names.Kind == msg.Blueprint.Kind {
			matchedSD = sd
			break
		}
	}

	// BlueprintDefinition must exist
	if matchedSD == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("BlueprintDefinition for kind '%s' not found in namespace '%s'", msg.Blueprint.Kind, msg.Blueprint.Metadata.Namespace), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedSD.Spec.Schema != nil {
		if err := core.ValidateBlueprintAgainstSchema(msg.Blueprint, matchedSD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("blueprint validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Check if spec changed and increment generation if it did
	specChanged := false
	if oldBlueprint != nil {
		// Preserve the ID from the existing blueprint
		msg.Blueprint.ID = oldBlueprint.ID

		reconciliation := core.CreateReconciliation(oldBlueprint, msg.Blueprint)
		if reconciliation.Diff != nil && len(reconciliation.Diff.SpecChanges) > 0 {
			// Spec changed, increment generation
			msg.Blueprint.Metadata.Generation = oldBlueprint.Metadata.Generation + 1
			specChanged = true
		} else {
			// Preserve old generation if spec didn't change
			msg.Blueprint.Metadata.Generation = oldBlueprint.Metadata.Generation
		}
	}

	err = h.server.BlueprintDB().UpdateBlueprint(msg.Blueprint)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Save blueprint history only if spec changed (not for status-only updates)
	if specChanged {
		history := core.CreateBlueprintHistory(msg.Blueprint, recoveredID, "update")
		if err := h.server.BlueprintDB().AddBlueprintHistory(history); err != nil {
			log.WithFields(log.Fields{"Error": err, "BlueprintID": msg.Blueprint.ID}).Warn("Failed to save blueprint history")
			// Don't fail the request if history saving fails
		}
	}

	log.WithFields(log.Fields{
		"ID":        msg.Blueprint.ID,
		"Namespace": msg.Blueprint.Metadata.Namespace,
		"Name":      msg.Blueprint.Metadata.Name,
	}).Debug("Updating blueprint")

	// Always submit reconciliation if handler is defined
	// The reconciler will determine if actual changes are needed via the Diff
	// This ensures reconciliation happens even when setting the same value
	// (useful for recovering from inconsistent state)
	if matchedSD != nil && matchedSD.Spec.Handler.ExecutorType != "" {
		reconciliation := core.CreateReconciliation(oldBlueprint, msg.Blueprint)
		processID, err := h.submitReconciliationFunc(reconciliation, matchedSD)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Kind":  msg.Blueprint.Kind,
			}).Warn("Failed to submit reconciliation after blueprint update")
			// Don't fail the request if reconciliation submission fails
		} else if processID != "" {
			// Update blueprint with reconciliation tracking info
			msg.Blueprint.Metadata.LastReconciliationProcess = processID
			msg.Blueprint.Metadata.LastReconciliationTime = time.Now()
			err = h.server.BlueprintDB().UpdateBlueprint(msg.Blueprint)
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
					"ProcessID": processID,
				}).Warn("Failed to update blueprint with reconciliation tracking info")
			} else {
				log.WithFields(log.Fields{
					"ProcessID": processID,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Info("Updated blueprint with reconciliation tracking info")
			}
		}
	}

	jsonString, err = msg.Blueprint.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}

// HandleRemoveBlueprint removes a Blueprint by namespace and name
func (h *Handlers) HandleRemoveBlueprint(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveBlueprintMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove blueprint, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove blueprint, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership or colony owner to remove blueprints
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Get the blueprint before removing it so we can trigger delete reconciliation
	blueprint, err := h.server.BlueprintDB().GetBlueprintByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if blueprint == nil {
		h.server.HandleHTTPError(c, errors.New("Blueprint not found"), http.StatusNotFound)
		return
	}

	// Get the BlueprintDefinition to find the handler
	var matchedSD *core.BlueprintDefinition
	sds, err := h.server.BlueprintDB().GetBlueprintDefinitions()
	if err == nil {
		for _, sd := range sds {
			if sd.Spec.Names.Kind == blueprint.Kind {
				matchedSD = sd
				break
			}
		}
	}

	// Trigger delete reconciliation BEFORE removing from database
	if matchedSD != nil {
		reconciliation := core.CreateReconciliation(blueprint, nil) // old=blueprint, new=nil => delete action
		processID, err := h.submitReconciliationFunc(reconciliation, matchedSD)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Kind":  blueprint.Kind,
				"Name":  blueprint.Metadata.Name,
			}).Warn("Failed to submit delete reconciliation")
			// Don't fail the request if reconciliation submission fails
		} else if processID != "" {
			log.WithFields(log.Fields{
				"ProcessID":     processID,
				"BlueprintName": blueprint.Metadata.Name,
				"Action":        "delete",
			}).Info("Submitted delete reconciliation process")
		}
	}

	// Now remove the blueprint from database
	err = h.server.BlueprintDB().RemoveBlueprintByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace": msg.Namespace,
		"Name":      msg.Name,
	}).Debug("Removing blueprint")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleGetBlueprintHistory retrieves history for a blueprint
func (h *Handlers) HandleGetBlueprintHistory(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetBlueprintHistoryMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint history, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get blueprint history, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Get the blueprint to check permissions
	blueprint, err := h.server.BlueprintDB().GetBlueprintByID(msg.BlueprintID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}
	if blueprint == nil {
		h.server.HandleHTTPError(c, errors.New("Blueprint not found"), http.StatusNotFound)
		return
	}

	// Require membership or colony owner to view history
	err = h.server.Validator().RequireMembership(recoveredID, blueprint.Metadata.Namespace, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, blueprint.Metadata.Namespace)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	histories, err := h.server.BlueprintDB().GetBlueprintHistory(msg.BlueprintID, msg.Limit)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = core.ConvertBlueprintHistoryArrayToJSON(histories)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"BlueprintID": msg.BlueprintID,
		"Limit":     msg.Limit,
		"Count":     len(histories),
	}).Debug("Getting blueprint history")

	h.server.SendHTTPReply(c, payloadType, jsonString)
}
