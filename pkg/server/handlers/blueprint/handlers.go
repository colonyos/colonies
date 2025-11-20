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
	ProcessController() process.Controller
	CronController() interface {
		AddCron(cron *core.Cron) (*core.Cron, error)
		GetCron(cronID string) (*core.Cron, error)
		GetCrons(colonyName string, count int) ([]*core.Cron, error)
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

// createReconciliationWorkflowSpec creates a workflow spec for blueprint reconciliation
// This workflow contains a single process that fetches and reconciles the blueprint
func (h *Handlers) createReconciliationWorkflowSpec(blueprint *core.Blueprint, sd *core.BlueprintDefinition) (string, error) {
	if sd == nil || sd.Spec.Handler.ExecutorType == "" {
		return "", fmt.Errorf("no handler defined for blueprint kind: %s", blueprint.Kind)
	}

	// Create a function spec that tells the reconciler to fetch and reconcile this blueprint
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "reconcile"
	funcSpec.Conditions.ColonyName = blueprint.Metadata.ColonyName
	funcSpec.Conditions.ExecutorType = sd.Spec.Handler.ExecutorType
	funcSpec.FuncName = "reconcile"
	funcSpec.KwArgs = map[string]interface{}{
		"blueprintName": blueprint.Metadata.Name,
	}

	// Apply executor targeting if specified
	if blueprint.Handler != nil {
		if len(blueprint.Handler.ExecutorNames) > 0 {
			funcSpec.Conditions.ExecutorNames = blueprint.Handler.ExecutorNames
			log.WithFields(log.Fields{
				"BlueprintName":  blueprint.Metadata.Name,
				"ExecutorNames":  blueprint.Handler.ExecutorNames,
			}).Debug("Applied executor targeting from ExecutorNames")
		} else if blueprint.Handler.ExecutorName != "" {
			funcSpec.Conditions.ExecutorNames = []string{blueprint.Handler.ExecutorName}
			log.WithFields(log.Fields{
				"BlueprintName": blueprint.Metadata.Name,
				"ExecutorName":  blueprint.Handler.ExecutorName,
			}).Debug("Applied executor targeting from ExecutorName")
		}
	} else {
		log.WithFields(log.Fields{
			"BlueprintName": blueprint.Metadata.Name,
		}).Debug("No handler specified for blueprint, no executor targeting applied")
	}

	// Create a simple workflow with one function
	workflowSpec := &core.WorkflowSpec{
		ColonyName:    blueprint.Metadata.ColonyName,
		FunctionSpecs: []core.FunctionSpec{*funcSpec},
	}

	workflowJSON, err := workflowSpec.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to create workflow spec: %w", err)
	}

	return workflowJSON, nil
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
		"Namespace": msg.Blueprint.Metadata.ColonyName,
		"Name":      msg.Blueprint.Metadata.Name,
		"Kind":      msg.Blueprint.Kind,
	}).Debug("Adding blueprint")

	// Auto-create reconciliation cron if handler is defined
	if matchedSD != nil && matchedSD.Spec.Handler.ExecutorType != "" {
		// Use blueprint name for cron name
		cronName := "reconcile-" + msg.Blueprint.Metadata.Name

		// Create workflow spec for reconciliation
		workflowSpec, err := h.createReconciliationWorkflowSpec(msg.Blueprint, matchedSD)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":         err,
				"BlueprintName": msg.Blueprint.Metadata.Name,
			}).Warn("Failed to create reconciliation workflow spec")
			// Don't fail the request if workflow spec creation fails
		} else {
			// Resolve initiator name
			initiatorName, err := h.resolveInitiator(msg.Blueprint.Metadata.ColonyName, recoveredID)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":         err,
					"RecoveredID":   recoveredID,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Warn("Failed to resolve initiator name for cron")
				// Don't fail the request, use empty string
				initiatorName = ""
			}

			// Create cron for periodic self-healing
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

			// Store cron reference in blueprint annotations
			if msg.Blueprint.Metadata.Annotations == nil {
				msg.Blueprint.Metadata.Annotations = make(map[string]string)
			}
			msg.Blueprint.Metadata.Annotations["reconciliation.cron.id"] = addedCron.ID
			msg.Blueprint.Metadata.Annotations["reconciliation.cron.name"] = addedCron.Name

			err = h.server.BlueprintDB().UpdateBlueprint(msg.Blueprint)
			if err != nil {
				// Rollback: remove cron if annotation update fails
				h.server.CronController().RemoveCron(addedCron.ID)
				h.server.BlueprintDB().RemoveBlueprintByID(msg.Blueprint.ID)
				h.server.HandleHTTPError(c, fmt.Errorf("failed to update blueprint with cron reference: %w", err), http.StatusInternalServerError)
				return
			}

			log.WithFields(log.Fields{
				"BlueprintName": msg.Blueprint.Metadata.Name,
				"CronName":      cronName,
				"CronID":        addedCron.ID,
				"Interval":      60,
			}).Info("Auto-created reconciliation cron for blueprint")

			// Trigger immediate reconciliation for initial create
			_, err = h.server.CronController().RunCron(addedCron.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":    err,
					"CronName": cronName,
				}).Warn("Failed to trigger initial reconciliation")
				// Don't fail the request if immediate trigger fails
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
		"Namespace": msg.Blueprint.Metadata.ColonyName,
		"Name":      msg.Blueprint.Metadata.Name,
		"Generation": msg.Blueprint.Metadata.Generation,
	}).Debug("Updating blueprint")

	// Trigger immediate reconciliation by running the cron
	cronName := msg.Blueprint.Metadata.Annotations["reconciliation.cron.name"]
	if cronName != "" {
		cron, err := h.server.CronController().GetCrons(msg.Blueprint.Metadata.ColonyName, 1000)
		if err == nil {
			// Find the cron by name
			for _, c := range cron {
				if c.Name == cronName {
					log.WithFields(log.Fields{
						"BlueprintName": msg.Blueprint.Metadata.Name,
						"Generation":    msg.Blueprint.Metadata.Generation,
						"CronName":      cronName,
					}).Info("Triggering immediate reconciliation after blueprint update")

					_, err = h.server.CronController().RunCron(c.ID)
					if err != nil {
						log.WithFields(log.Fields{
							"Error":    err,
							"CronName": cronName,
						}).Warn("Failed to trigger reconciliation cron - reconciliation will happen on next periodic trigger")
						// Don't fail the request if cron trigger fails
					}
					break
				}
			}
		} else {
			log.WithFields(log.Fields{
				"Error":         err,
				"BlueprintName": msg.Blueprint.Metadata.Name,
				"CronName":      cronName,
			}).Warn("Failed to get crons for reconciliation trigger")
		}
	} else {
		log.WithFields(log.Fields{
			"BlueprintName": msg.Blueprint.Metadata.Name,
		}).Debug("No reconciliation cron found for blueprint (may be old blueprint without cron)")
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

	// Get cron reference before deleting blueprint
	cronName := blueprint.Metadata.Annotations["reconciliation.cron.name"]

	// Remove the blueprint from database first
	err = h.server.BlueprintDB().RemoveBlueprintByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Auto-remove the reconciliation cron (best-effort, don't fail if cron missing)
	if cronName != "" {
		crons, err := h.server.CronController().GetCrons(msg.Namespace, 1000)
		if err == nil {
			for _, c := range crons {
				if c.Name == cronName {
					err = h.server.CronController().RemoveCron(c.ID)
					if err != nil {
						log.WithFields(log.Fields{
							"Error":         err,
							"BlueprintName": msg.Name,
							"CronName":      cronName,
						}).Warn("Failed to remove reconciliation cron (may already be deleted)")
					} else {
						log.WithFields(log.Fields{
							"BlueprintName": msg.Name,
							"CronName":      cronName,
						}).Info("Auto-removed reconciliation cron for deleted blueprint")
					}
					break
				}
			}
		} else {
			log.WithFields(log.Fields{
				"Error":    err,
				"CronName": cronName,
			}).Warn("Failed to get crons for deletion")
		}
	}

	// Submit cleanup process to reconciler (best-effort)
	// Look up the BlueprintDefinition by Kind to get the ExecutorType
	var blueprintDef *core.BlueprintDefinition
	blueprintDefs, err := h.server.BlueprintDB().GetBlueprintDefinitions()
	if err == nil {
		for _, def := range blueprintDefs {
			if def.Spec.Names.Kind == blueprint.Kind {
				blueprintDef = def
				break
			}
		}
	}
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

		// Apply executor targeting if specified
		if blueprint.Handler != nil {
			if len(blueprint.Handler.ExecutorNames) > 0 {
				funcSpec.Conditions.ExecutorNames = blueprint.Handler.ExecutorNames
			} else if blueprint.Handler.ExecutorName != "" {
				funcSpec.Conditions.ExecutorNames = []string{blueprint.Handler.ExecutorName}
			}
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
