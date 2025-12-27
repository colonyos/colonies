package blueprint

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
	BlueprintDB() database.BlueprintDatabase
	LocationDB() database.LocationDatabase
	ProcessController() process.Controller
	CronController() interface {
		AddCron(cron *core.Cron) (*core.Cron, error)
		GetCron(cronID string) (*core.Cron, error)
		GetCrons(colonyName string, count int) ([]*core.Cron, error)
		GetCronByName(colonyName string, cronName string) (*core.Cron, error)
		RunCron(cronID string) (*core.Cron, error)
		RemoveCron(cronID string) error
		GetCronPeriod() int
	}
	ExecutorDB() database.ExecutorDatabase
	UserDB() database.UserDatabase
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{
		server: server,
	}
}

// resolveInitiator resolves the initiator name from the recoveredID
func (h *Handlers) resolveInitiator(colonyName string, recoveredID string) (string, error) {
	executor, err := h.server.ExecutorDB().GetExecutorByID(recoveredID)
	if err != nil {
		return "", err
	}

	if executor != nil {
		return executor.Name, nil
	} else {
		user, err := h.server.UserDB().GetUserByID(colonyName, recoveredID)
		if err != nil {
			return "", err
		}
		if user != nil {
			return user.Name, nil
		} else {
			return "", errors.New("Could not derive InitiatorName")
		}
	}
}

// createConsolidatedReconciliationWorkflowSpec creates a workflow spec for consolidated blueprint reconciliation
// This workflow creates processes for each unique executor type, routing blueprints to their handlers
// Blueprints can specify handler.executorType to target location-specific reconcilers
func (h *Handlers) createConsolidatedReconciliationWorkflowSpec(colonyName string, kind string, sd *core.BlueprintDefinition) (string, error) {
	if sd == nil || sd.Spec.Handler.ExecutorType == "" {
		return "", fmt.Errorf("no handler defined for blueprint kind: %s", kind)
	}

	defaultExecutorType := sd.Spec.Handler.ExecutorType

	// Get all blueprints of this Kind to find unique handler configurations
	blueprints, err := h.server.BlueprintDB().GetBlueprintsByNamespaceAndKind(colonyName, kind)
	if err != nil {
		return "", fmt.Errorf("failed to get blueprints for kind %s: %w", kind, err)
	}

	log.WithFields(log.Fields{
		"ColonyName":          colonyName,
		"Kind":                kind,
		"BlueprintCount":      len(blueprints),
		"DefaultExecutorType": defaultExecutorType,
	}).Info("Creating consolidated workflow spec")

	// Group blueprints by their handler configuration (executor type only)
	// Reconciliation is based on executor type + location, not specific executor names
	handlers := make(map[string]bool)

	for _, bp := range blueprints {
		executorType := defaultExecutorType

		if bp.Handler != nil {
			// Use blueprint's executor type if specified, otherwise use default
			if bp.Handler.ExecutorType != "" {
				executorType = bp.Handler.ExecutorType
			}
		}

		handlers[executorType] = true
		log.WithFields(log.Fields{
			"BlueprintName": bp.Metadata.Name,
			"ExecutorType":  executorType,
		}).Debug("Blueprint handler configuration")
	}

	log.WithFields(log.Fields{
		"UniqueHandlers": len(handlers),
	}).Info("Collected unique handler configurations")

	// Create function specs - one for each unique executor type
	var functionSpecs []core.FunctionSpec
	i := 0
	for executorType := range handlers {
		funcSpec := core.CreateEmptyFunctionSpec()
		funcSpec.NodeName = fmt.Sprintf("reconcile-%d", i)
		funcSpec.Conditions.ColonyName = colonyName
		funcSpec.Conditions.ExecutorType = executorType
		funcSpec.FuncName = "reconcile"
		funcSpec.KwArgs = map[string]interface{}{
			"kind": kind,
		}

		log.WithFields(log.Fields{
			"NodeName":     funcSpec.NodeName,
			"ExecutorType": executorType,
		}).Info("Created function spec for handler")

		functionSpecs = append(functionSpecs, *funcSpec)
		i++
	}

	// Create workflow with functions for each handler
	workflowSpec := &core.WorkflowSpec{
		ColonyName:    colonyName,
		FunctionSpecs: functionSpecs,
	}

	workflowJSON, err := workflowSpec.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to create workflow spec: %w", err)
	}

	return workflowJSON, nil
}

// createReconcilerCronWorkflowSpec creates a workflow spec for a specific reconciler
// This creates a single process targeting handler executors by type and location
func (h *Handlers) createReconcilerCronWorkflowSpec(colonyName string, kind string, executorType string, locationName string) (string, error) {
	if executorType == "" {
		return "", fmt.Errorf("no handler executorType defined for blueprint kind: %s", kind)
	}

	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "reconcile"
	funcSpec.Conditions.ColonyName = colonyName
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = locationName
	funcSpec.FuncName = "reconcile"
	funcSpec.KwArgs = map[string]interface{}{
		"kind": kind,
	}

	workflowSpec := &core.WorkflowSpec{
		ColonyName:    colonyName,
		FunctionSpecs: []core.FunctionSpec{*funcSpec},
	}

	workflowJSON, err := workflowSpec.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to create workflow spec: %w", err)
	}

	return workflowJSON, nil
}

// createImmediateReconciliationProcess creates a process for immediate reconciliation of a blueprint's Kind
// This is used when a blueprint is added to trigger immediate reconciliation of all blueprints of that Kind
func (h *Handlers) createImmediateReconciliationProcess(blueprint *core.Blueprint, sd *core.BlueprintDefinition, recoveredID string, initiatorName string) (*core.Process, error) {
	// Determine executor type - prefer BlueprintDefinition, fall back to Blueprint handler
	executorType := ""
	if sd != nil && sd.Spec.Handler.ExecutorType != "" {
		executorType = sd.Spec.Handler.ExecutorType
	} else if blueprint.Handler != nil && blueprint.Handler.ExecutorType != "" {
		executorType = blueprint.Handler.ExecutorType
	}

	if executorType == "" {
		return nil, fmt.Errorf("no handler defined for blueprint kind: %s", blueprint.Kind)
	}

	// Use blueprint's handler type if specified (override definition default)
	if blueprint.Handler != nil && blueprint.Handler.ExecutorType != "" {
		executorType = blueprint.Handler.ExecutorType
	}

	// Create a function spec for reconciliation of this specific blueprint
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "reconcile"
	funcSpec.Conditions.ColonyName = blueprint.Metadata.ColonyName
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = blueprint.Metadata.LocationName
	funcSpec.FuncName = "reconcile"
	funcSpec.KwArgs = map[string]interface{}{
		"kind":          blueprint.Kind,
		"blueprintName": blueprint.Metadata.Name,
	}

	// Create and return the process
	process := core.CreateProcess(funcSpec)
	process.InitiatorID = recoveredID
	process.InitiatorName = initiatorName

	return process, nil
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
	if err := handlerRegistry.Register(rpc.UpdateBlueprintStatusPayloadType, h.HandleUpdateBlueprintStatus); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.ReconcileBlueprintPayloadType, h.HandleReconcileBlueprint); err != nil {
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
	err = h.server.Validator().RequireColonyOwner(recoveredID, msg.BlueprintDefinition.Metadata.ColonyName)
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
		"ColonyName": msg.BlueprintDefinition.Metadata.ColonyName,
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

	// If sds is nil, convert to empty array (similar to how crons/executors are handled)
	if sds == nil {
		sds = []*core.BlueprintDefinition{}
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
	err = h.server.Validator().RequireMembership(recoveredID, msg.Blueprint.Metadata.ColonyName, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Blueprint.Metadata.ColonyName)
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
	sds, err := h.server.BlueprintDB().GetBlueprintDefinitionsByNamespace(msg.Blueprint.Metadata.ColonyName)
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
		h.server.HandleHTTPError(c, fmt.Errorf("BlueprintDefinition for kind '%s' not found in namespace '%s'", msg.Blueprint.Kind, msg.Blueprint.Metadata.ColonyName), http.StatusBadRequest)
		return
	}

	// Validate against schema if defined
	if matchedSD.Spec.Schema != nil {
		if err := core.ValidateBlueprintAgainstSchema(msg.Blueprint, matchedSD.Spec.Schema); err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("blueprint validation failed: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Check if blueprint with same name already exists
	existingBlueprint, err := h.server.BlueprintDB().GetBlueprintByName(msg.Blueprint.Metadata.ColonyName, msg.Blueprint.Metadata.Name)
	if err != nil {
		h.server.HandleHTTPError(c, fmt.Errorf("failed to check for existing blueprint: %w", err), http.StatusInternalServerError)
		return
	}
	if existingBlueprint != nil {
		h.server.HandleHTTPError(c, fmt.Errorf("blueprint with name '%s' already exists in colony '%s'", msg.Blueprint.Metadata.Name, msg.Blueprint.Metadata.ColonyName), http.StatusConflict)
		return
	}

	// Auto-create location if specified and doesn't exist
	var locationWasAutoCreated bool
	if msg.Blueprint.Metadata.LocationName != "" {
		existingLocation, err := h.server.LocationDB().GetLocationByName(msg.Blueprint.Metadata.ColonyName, msg.Blueprint.Metadata.LocationName)
		if err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("failed to check location: %w", err), http.StatusInternalServerError)
			return
		}

		if existingLocation == nil {
			// Create new location
			newLocation := core.CreateLocation(
				core.GenerateRandomID(),
				msg.Blueprint.Metadata.LocationName,
				msg.Blueprint.Metadata.ColonyName,
				"Auto-created from blueprint "+msg.Blueprint.Metadata.Name,
				0.0, // Default longitude
				0.0, // Default latitude
			)
			err = h.server.LocationDB().AddLocation(newLocation)
			if err != nil {
				h.server.HandleHTTPError(c, fmt.Errorf("failed to create location: %w", err), http.StatusInternalServerError)
				return
			}
			locationWasAutoCreated = true

			log.WithFields(log.Fields{
				"LocationName":  msg.Blueprint.Metadata.LocationName,
				"ColonyName":    msg.Blueprint.Metadata.ColonyName,
				"BlueprintName": msg.Blueprint.Metadata.Name,
			}).Info("Auto-created location for blueprint")
		}
	}

	err = h.server.BlueprintDB().AddBlueprint(msg.Blueprint)
	if err != nil {
		// Clean up auto-created location if blueprint creation fails
		if locationWasAutoCreated {
			if removeErr := h.server.LocationDB().RemoveLocationByName(msg.Blueprint.Metadata.ColonyName, msg.Blueprint.Metadata.LocationName); removeErr != nil {
				log.WithFields(log.Fields{
					"Error":        removeErr,
					"LocationName": msg.Blueprint.Metadata.LocationName,
					"ColonyName":   msg.Blueprint.Metadata.ColonyName,
				}).Warn("Failed to cleanup auto-created location after blueprint creation failure")
			} else {
				log.WithFields(log.Fields{
					"LocationName":  msg.Blueprint.Metadata.LocationName,
					"ColonyName":    msg.Blueprint.Metadata.ColonyName,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Info("Cleaned up auto-created location after blueprint creation failure")
			}
		}
		h.server.HandleHTTPError(c, err, http.StatusInternalServerError)
		return
	}

	// Save blueprint history for create action
	history := core.CreateBlueprintHistory(msg.Blueprint, recoveredID, "create")
	if err := h.server.BlueprintDB().AddBlueprintHistory(history); err != nil {
		log.WithFields(log.Fields{"Error": err, "BlueprintID": msg.Blueprint.ID}).Error("Failed to save blueprint history")
		h.server.HandleHTTPError(c, fmt.Errorf("blueprint created but failed to save audit history: %w", err), http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"ID":        msg.Blueprint.ID,
		"Namespace": msg.Blueprint.Metadata.ColonyName,
		"Name":      msg.Blueprint.Metadata.Name,
		"Kind":      msg.Blueprint.Kind,
	}).Debug("Adding blueprint")

	// Auto-create reconciliation cron if handler is defined
	// Start with BlueprintDefinition handler as default, then override with Blueprint handler if specified
	executorType := ""
	if matchedSD != nil && matchedSD.Spec.Handler.ExecutorType != "" {
		executorType = matchedSD.Spec.Handler.ExecutorType
	}
	// Override with blueprint's handler type if specified (blueprint handler takes precedence)
	if msg.Blueprint.Handler != nil && msg.Blueprint.Handler.ExecutorType != "" {
		executorType = msg.Blueprint.Handler.ExecutorType
	}

	if executorType != "" {
		// Use Kind + location for cron name (one cron per reconciler per location)
		locationName := msg.Blueprint.Metadata.LocationName
		cronName := "reconcile-" + msg.Blueprint.Kind
		if locationName != "" {
			cronName = cronName + "-" + locationName
		}

		// Resolve initiator name
		initiatorName, err := h.resolveInitiator(msg.Blueprint.Metadata.ColonyName, recoveredID)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":         err,
				"RecoveredID":   recoveredID,
				"BlueprintName": msg.Blueprint.Metadata.Name,
			}).Warn("Failed to resolve initiator name for cron")
			initiatorName = ""
		}

		// Check if cron for this handler already exists
		existingCron, err := h.server.CronController().GetCronByName(msg.Blueprint.Metadata.ColonyName, cronName)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":    err,
				"CronName": cronName,
			}).Warn("Failed to check for existing cron")
		}

		if existingCron == nil {
			// Create workflow spec targeting handler executors by type and location
			workflowSpec, err := h.createReconcilerCronWorkflowSpec(msg.Blueprint.Metadata.ColonyName, msg.Blueprint.Kind, executorType, locationName)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":        err,
					"Kind":         msg.Blueprint.Kind,
					"LocationName": locationName,
				}).Warn("Failed to create reconciliation workflow spec")
			} else {
				// Create cron for periodic self-healing by reconcilers at this location
				cron := &core.Cron{
					ID:                      core.GenerateRandomID(),
					ColonyName:              msg.Blueprint.Metadata.ColonyName,
					Name:                    cronName,
					Interval:                60, // 60 seconds
					WaitForPrevProcessGraph: true,
					WorkflowSpec:            workflowSpec,
					InitiatorID:             recoveredID,
					InitiatorName:           initiatorName,
				}

				addedCron, err := h.server.CronController().AddCron(cron)
				if err != nil {
					// Rollback: remove blueprint if cron creation fails
					h.server.BlueprintDB().RemoveBlueprintByID(msg.Blueprint.ID)
					h.server.HandleHTTPError(c, fmt.Errorf("failed to create reconciliation cron: %w", err), http.StatusInternalServerError)
					return
				}

				log.WithFields(log.Fields{
					"Kind":         msg.Blueprint.Kind,
					"CronName":     cronName,
					"CronID":       addedCron.ID,
					"LocationName": locationName,
					"Interval":     60,
				}).Info("Auto-created reconciliation cron for handler")

				existingCron = addedCron
			}
		} else {
			log.WithFields(log.Fields{
				"BlueprintName": msg.Blueprint.Metadata.Name,
				"Kind":          msg.Blueprint.Kind,
				"CronName":      cronName,
				"LocationName":  locationName,
			}).Debug("Cron already exists for handler")
		}

		// Submit immediate reconciliation process for this specific blueprint
		if existingCron != nil {
			immediateProcess, err := h.createImmediateReconciliationProcess(msg.Blueprint, matchedSD, recoveredID, initiatorName)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":         err,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Error("Failed to create immediate reconciliation process")
				h.server.HandleHTTPError(c, fmt.Errorf("blueprint created but failed to create reconciliation process: %w", err), http.StatusInternalServerError)
				return
			}
			_, err = h.server.ProcessController().AddProcess(immediateProcess)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":         err,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Error("Failed to submit immediate reconciliation process")
				h.server.HandleHTTPError(c, fmt.Errorf("blueprint created but failed to submit reconciliation process: %w", err), http.StatusInternalServerError)
				return
			}
			log.WithFields(log.Fields{
				"BlueprintName": msg.Blueprint.Metadata.Name,
				"ProcessID":     immediateProcess.ID,
			}).Info("Submitted immediate reconciliation process for new blueprint")
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
	} else if msg.LocationName != "" {
		// Get blueprints by namespace, kind, and location
		blueprints, err = h.server.BlueprintDB().GetBlueprintsByNamespaceKindAndLocation(msg.Namespace, msg.Kind, msg.LocationName)
	} else {
		// Get blueprints by namespace and kind
		blueprints, err = h.server.BlueprintDB().GetBlueprintsByNamespaceAndKind(msg.Namespace, msg.Kind)
	}

	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"Namespace":    msg.Namespace,
		"Kind":         msg.Kind,
		"LocationName": msg.LocationName,
		"Count":        len(blueprints),
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
	err = h.server.Validator().RequireMembership(recoveredID, msg.Blueprint.Metadata.ColonyName, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, msg.Blueprint.Metadata.ColonyName)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}
	}

	// Get the old blueprint for reconciliation
	oldBlueprint, err := h.server.BlueprintDB().GetBlueprintByName(msg.Blueprint.Metadata.ColonyName, msg.Blueprint.Metadata.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Blueprint must exist for update - return 404 if not found
	if oldBlueprint == nil {
		h.server.HandleHTTPError(c, fmt.Errorf("blueprint '%s' not found in namespace '%s'", msg.Blueprint.Metadata.Name, msg.Blueprint.Metadata.ColonyName), http.StatusNotFound)
		return
	}

	// Validate blueprint against its BlueprintDefinition schema
	// Blueprint Kind is required
	if msg.Blueprint.Kind == "" {
		h.server.HandleHTTPError(c, errors.New("blueprint kind is required"), http.StatusBadRequest)
		return
	}

	// Fetch all BlueprintDefinitions in the namespace
	sds, err := h.server.BlueprintDB().GetBlueprintDefinitionsByNamespace(msg.Blueprint.Metadata.ColonyName)
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
		h.server.HandleHTTPError(c, fmt.Errorf("BlueprintDefinition for kind '%s' not found in namespace '%s'", msg.Blueprint.Kind, msg.Blueprint.Metadata.ColonyName), http.StatusBadRequest)
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
		} else if msg.ForceGeneration {
			// Force generation bump requested (for force reconciliation)
			msg.Blueprint.Metadata.Generation = oldBlueprint.Metadata.Generation + 1
			specChanged = true
			log.WithFields(log.Fields{
				"BlueprintName": msg.Blueprint.Metadata.Name,
				"OldGeneration": oldBlueprint.Metadata.Generation,
				"NewGeneration": msg.Blueprint.Metadata.Generation,
			}).Info("Force generation bump requested")
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
			log.WithFields(log.Fields{"Error": err, "BlueprintID": msg.Blueprint.ID}).Error("Failed to save blueprint history")
			h.server.HandleHTTPError(c, fmt.Errorf("blueprint updated but failed to save audit history: %w", err), http.StatusInternalServerError)
			return
		}
	}

	log.WithFields(log.Fields{
		"ID":        msg.Blueprint.ID,
		"Namespace": msg.Blueprint.Metadata.ColonyName,
		"Name":      msg.Blueprint.Metadata.Name,
		"Generation": msg.Blueprint.Metadata.Generation,
	}).Debug("Updating blueprint")

	// Trigger immediate reconciliation for this specific blueprint
	// This creates a process with blueprintName so the reconciler knows exactly which blueprint changed
	if specChanged {
		initiatorName, err := h.resolveInitiator(msg.Blueprint.Metadata.ColonyName, recoveredID)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":         err,
				"BlueprintName": msg.Blueprint.Metadata.Name,
			}).Warn("Failed to resolve initiator for reconciliation")
		} else {
			immediateProcess, err := h.createImmediateReconciliationProcess(msg.Blueprint, matchedSD, recoveredID, initiatorName)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":         err,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Warn("Failed to create immediate reconciliation process after blueprint update")
			} else {
				_, err = h.server.ProcessController().AddProcess(immediateProcess)
				if err != nil {
					log.WithFields(log.Fields{
						"Error":         err,
						"BlueprintName": msg.Blueprint.Metadata.Name,
					}).Warn("Failed to submit immediate reconciliation process after blueprint update")
				} else {
					log.WithFields(log.Fields{
						"BlueprintName": msg.Blueprint.Metadata.Name,
						"Generation":    msg.Blueprint.Metadata.Generation,
						"ProcessID":     immediateProcess.ID,
					}).Info("Submitted immediate reconciliation process for updated blueprint")
				}
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

	// Cron naming must match AddBlueprint: reconcile-{Kind}-{locationName}
	locationName := blueprint.Metadata.LocationName
	cronName := "reconcile-" + blueprint.Kind
	if locationName != "" {
		cronName = cronName + "-" + locationName
	}

	// Remove the blueprint from database first
	err = h.server.BlueprintDB().RemoveBlueprintByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Check if there are any remaining blueprints of this Kind at this location
	remainingBlueprints, err := h.server.BlueprintDB().GetBlueprintsByNamespaceAndKind(msg.Namespace, blueprint.Kind)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"Kind":  blueprint.Kind,
		}).Warn("Failed to check remaining blueprints")
		remainingBlueprints = []*core.Blueprint{} // Assume none remaining on error
	}

	// Filter to only blueprints at the same location
	remainingAtLocation := 0
	for _, bp := range remainingBlueprints {
		if bp.Metadata.LocationName == locationName {
			remainingAtLocation++
		}
	}

	// Only remove cron if no blueprints of this Kind remain at this location
	if remainingAtLocation == 0 {
		existingCron, err := h.server.CronController().GetCronByName(msg.Namespace, cronName)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":    err,
				"CronName": cronName,
			}).Warn("Failed to get cron for deletion")
		} else if existingCron != nil {
			err = h.server.CronController().RemoveCron(existingCron.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":        err,
					"Kind":         blueprint.Kind,
					"CronName":     cronName,
					"LocationName": locationName,
				}).Warn("Failed to remove cron")
			} else {
				log.WithFields(log.Fields{
					"Kind":         blueprint.Kind,
					"CronName":     cronName,
					"LocationName": locationName,
				}).Info("Auto-removed cron (last blueprint of Kind at location deleted)")
			}
		}
	} else {
		log.WithFields(log.Fields{
			"BlueprintName":       msg.Name,
			"Kind":                blueprint.Kind,
			"LocationName":        locationName,
			"RemainingBlueprints": remainingAtLocation,
		}).Debug("Cron kept (other blueprints of same Kind at location exist)")
	}

	// Submit cleanup process to reconciler (best-effort)
	// Look up the BlueprintDefinition by Kind to get the ExecutorType
	blueprintDef, err := h.server.BlueprintDB().GetBlueprintDefinitionByKind(blueprint.Kind)
	if blueprintDef != nil && blueprintDef.Spec.Handler.ExecutorType != "" {
		// Create cleanup function spec
		funcSpec := core.CreateEmptyFunctionSpec()
		funcSpec.NodeName = "cleanup"
		funcSpec.Conditions.ColonyName = msg.Namespace
		funcSpec.Conditions.ExecutorType = blueprintDef.Spec.Handler.ExecutorType
		funcSpec.FuncName = "cleanup"
		funcSpec.KwArgs = map[string]interface{}{
			"blueprintName": msg.Name,
		}

		// Resolve initiator name for the process
		initiatorName, _ := h.resolveInitiator(msg.Namespace, recoveredID)

		// Create and submit the process
		cleanupProcess := core.CreateProcess(funcSpec)
		cleanupProcess.InitiatorID = recoveredID
		cleanupProcess.InitiatorName = initiatorName

		_, err = h.server.ProcessController().AddProcess(cleanupProcess)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":         err,
				"BlueprintName": msg.Name,
			}).Warn("Failed to submit cleanup process")
		} else {
			log.WithFields(log.Fields{
				"BlueprintName": msg.Name,
				"ExecutorType":  blueprintDef.Spec.Handler.ExecutorType,
			}).Info("Submitted cleanup process for deleted blueprint")
		}
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
	err = h.server.Validator().RequireMembership(recoveredID, blueprint.Metadata.ColonyName, true)
	if err != nil {
		// If not a member, check if colony owner
		err = h.server.Validator().RequireColonyOwner(recoveredID, blueprint.Metadata.ColonyName)
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

// HandleUpdateBlueprintStatus updates only the status field of a blueprint
// This is used by reconcilers to report status without triggering a full update
func (h *Handlers) HandleUpdateBlueprintStatus(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateUpdateBlueprintStatusMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to update blueprint status, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to update blueprint status, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership to update blueprint status (typically done by executors/reconcilers)
	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Get the blueprint by name
	blueprint, err := h.server.BlueprintDB().GetBlueprintByName(msg.ColonyName, msg.BlueprintName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if blueprint == nil {
		h.server.HandleHTTPError(c, errors.New("Blueprint not found"), http.StatusNotFound)
		return
	}

	// Update only the status
	err = h.server.BlueprintDB().UpdateBlueprintStatus(blueprint.ID, msg.Status)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"ColonyName":    msg.ColonyName,
		"BlueprintName": msg.BlueprintName,
	}).Debug("Updated blueprint status")

	h.server.SendEmptyHTTPReply(c, payloadType)
}

// HandleReconcileBlueprint triggers immediate reconciliation of a specific blueprint
// The server looks up the executor type from the blueprint's handler configuration
func (h *Handlers) HandleReconcileBlueprint(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateReconcileBlueprintMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to reconcile blueprint, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to reconcile blueprint, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	// Require membership to trigger reconciliation
	err = h.server.Validator().RequireMembership(recoveredID, msg.Namespace, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Get the blueprint by name
	blueprint, err := h.server.BlueprintDB().GetBlueprintByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	if blueprint == nil {
		h.server.HandleHTTPError(c, errors.New("Blueprint not found"), http.StatusNotFound)
		return
	}

	// If force flag is set, bump the generation first
	if msg.Force {
		blueprint.Metadata.Generation++
		err = h.server.BlueprintDB().UpdateBlueprint(blueprint)
		if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
			return
		}

		log.WithFields(log.Fields{
			"BlueprintName": blueprint.Metadata.Name,
			"Generation":    blueprint.Metadata.Generation,
		}).Debug("Bumped blueprint generation for force reconciliation")
	}

	// Resolve initiator name for the process
	initiatorName, err := h.resolveInitiator(msg.Namespace, recoveredID)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Determine executor type - use blueprint's handler first, then fall back to definition
	var executorType string
	if blueprint.Handler != nil && blueprint.Handler.ExecutorType != "" {
		executorType = blueprint.Handler.ExecutorType
	} else {
		// Try to get from blueprint definition
		sd, err := h.server.BlueprintDB().GetBlueprintDefinitionByName(msg.Namespace, blueprint.Kind)
		if err == nil && sd != nil && sd.Spec.Handler.ExecutorType != "" {
			executorType = sd.Spec.Handler.ExecutorType
		}
	}

	if executorType == "" {
		h.server.HandleHTTPError(c, errors.New("Blueprint has no handler with executor type defined"), http.StatusBadRequest)
		return
	}

	// Create the reconciliation process
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "reconcile"
	funcSpec.Conditions.ColonyName = msg.Namespace
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = blueprint.Metadata.LocationName
	funcSpec.FuncName = "reconcile"
	funcSpec.KwArgs = map[string]interface{}{
		"kind":          blueprint.Kind,
		"blueprintName": blueprint.Metadata.Name,
	}

	// Pass force flag to reconciler
	if msg.Force {
		funcSpec.KwArgs["force"] = true
	}

	// Create and submit the process
	process := core.CreateProcess(funcSpec)
	process.InitiatorID = recoveredID
	process.InitiatorName = initiatorName

	addedProcess, err := h.server.ProcessController().AddProcess(process)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{
		"BlueprintName": blueprint.Metadata.Name,
		"Kind":          blueprint.Kind,
		"ExecutorType":  executorType,
		"ProcessID":     addedProcess.ID,
		"Force":         msg.Force,
	}).Debug("Submitted reconciliation process")

	// Return the process so client can track it
	jsonString, err = addedProcess.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	h.server.SendHTTPReply(c, payloadType, jsonString)
}
