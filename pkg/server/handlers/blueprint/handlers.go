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

// createConsolidatedReconciliationWorkflowSpec creates a workflow spec for consolidated blueprint reconciliation
// This workflow creates a single process that reconciles ALL blueprints of a given Kind
// The reconciler will fetch all blueprints and reconcile them in parallel
func (h *Handlers) createConsolidatedReconciliationWorkflowSpec(colonyName string, kind string, sd *core.BlueprintDefinition) (string, error) {
	if sd == nil || sd.Spec.Handler.ExecutorType == "" {
		return "", fmt.Errorf("no handler defined for blueprint kind: %s", kind)
	}

	// Get all blueprints of this Kind to find unique handler executors
	blueprints, err := h.server.BlueprintDB().GetBlueprintsByNamespaceAndKind(colonyName, kind)
	if err != nil {
		return "", fmt.Errorf("failed to get blueprints for kind %s: %w", kind, err)
	}

	log.WithFields(log.Fields{
		"ColonyName":      colonyName,
		"Kind":            kind,
		"BlueprintCount":  len(blueprints),
	}).Info("Creating consolidated workflow spec")

	// Collect unique handler executor names
	executorNames := make(map[string]bool)
	for _, bp := range blueprints {
		if bp.Handler != nil {
			if bp.Handler.ExecutorName != "" {
				executorNames[bp.Handler.ExecutorName] = true
				log.WithFields(log.Fields{
					"BlueprintName": bp.Metadata.Name,
					"ExecutorName":  bp.Handler.ExecutorName,
				}).Info("Found handler executor")
			}
			for _, name := range bp.Handler.ExecutorNames {
				executorNames[name] = true
			}
		}
	}

	log.WithFields(log.Fields{
		"UniqueExecutors": len(executorNames),
	}).Info("Collected unique handler executors")

	// Create function specs - one for each unique handler executor (or one for all if none specified)
	var functionSpecs []core.FunctionSpec

	if len(executorNames) == 0 {
		// No specific handlers - create single function targeting executor type
		funcSpec := core.CreateEmptyFunctionSpec()
		funcSpec.NodeName = "reconcile-all"
		funcSpec.Conditions.ColonyName = colonyName
		funcSpec.Conditions.ExecutorType = sd.Spec.Handler.ExecutorType
		funcSpec.FuncName = "reconcile"
		funcSpec.KwArgs = map[string]interface{}{
			"kind": kind,
		}
		functionSpecs = append(functionSpecs, *funcSpec)
	} else {
		// Create one function spec per unique handler executor
		i := 0
		for executorName := range executorNames {
			funcSpec := core.CreateEmptyFunctionSpec()
			funcSpec.NodeName = fmt.Sprintf("reconcile-%d", i)
			funcSpec.Conditions.ColonyName = colonyName
			funcSpec.Conditions.ExecutorType = sd.Spec.Handler.ExecutorType
			funcSpec.Conditions.ExecutorNames = []string{executorName}
			funcSpec.FuncName = "reconcile"
			funcSpec.KwArgs = map[string]interface{}{
				"kind": kind,
			}
			functionSpecs = append(functionSpecs, *funcSpec)
			i++
		}
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
// This creates a single process targeting one specific handler executor
func (h *Handlers) createReconcilerCronWorkflowSpec(colonyName string, kind string, sd *core.BlueprintDefinition, handlerExecutorName string) (string, error) {
	if sd == nil || sd.Spec.Handler.ExecutorType == "" {
		return "", fmt.Errorf("no handler defined for blueprint kind: %s", kind)
	}

	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "reconcile"
	funcSpec.Conditions.ColonyName = colonyName
	funcSpec.Conditions.ExecutorType = sd.Spec.Handler.ExecutorType
	funcSpec.FuncName = "reconcile"
	funcSpec.KwArgs = map[string]interface{}{
		"kind": kind,
	}

	// Target specific executor if specified
	if handlerExecutorName != "" {
		funcSpec.Conditions.ExecutorNames = []string{handlerExecutorName}
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
	if sd == nil || sd.Spec.Handler.ExecutorType == "" {
		return nil, fmt.Errorf("no handler defined for blueprint kind: %s", blueprint.Kind)
	}

	// Create a function spec for consolidated reconciliation by Kind
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "reconcile"
	funcSpec.Conditions.ColonyName = blueprint.Metadata.ColonyName
	funcSpec.Conditions.ExecutorType = sd.Spec.Handler.ExecutorType
	funcSpec.FuncName = "reconcile"
	funcSpec.KwArgs = map[string]interface{}{
		"kind": blueprint.Kind,
	}

	// Apply executor targeting if specified
	if blueprint.Handler != nil {
		if len(blueprint.Handler.ExecutorNames) > 0 {
			funcSpec.Conditions.ExecutorNames = blueprint.Handler.ExecutorNames
		} else if blueprint.Handler.ExecutorName != "" {
			funcSpec.Conditions.ExecutorNames = []string{blueprint.Handler.ExecutorName}
		}
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
		// Get the handler executor name for this blueprint
		handlerExecutorName := ""
		if msg.Blueprint.Handler != nil {
			if msg.Blueprint.Handler.ExecutorName != "" {
				handlerExecutorName = msg.Blueprint.Handler.ExecutorName
			} else if len(msg.Blueprint.Handler.ExecutorNames) > 0 {
				handlerExecutorName = msg.Blueprint.Handler.ExecutorNames[0]
			}
		}

		// Use Kind + handler executor for cron name (one cron per reconciler)
		cronName := "reconcile-" + msg.Blueprint.Kind
		if handlerExecutorName != "" {
			cronName = cronName + "-" + handlerExecutorName
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
		crons, err := h.server.CronController().GetCrons(msg.Blueprint.Metadata.ColonyName, 1000)
		var existingCron *core.Cron
		if err == nil {
			for _, cron := range crons {
				if cron.Name == cronName {
					existingCron = cron
					break
				}
			}
		}

		if existingCron == nil {
			// Create workflow spec targeting this specific handler executor
			workflowSpec, err := h.createReconcilerCronWorkflowSpec(msg.Blueprint.Metadata.ColonyName, msg.Blueprint.Kind, matchedSD, handlerExecutorName)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":           err,
					"Kind":            msg.Blueprint.Kind,
					"HandlerExecutor": handlerExecutorName,
				}).Warn("Failed to create reconciliation workflow spec")
			} else {
				// Create cron for periodic self-healing by this specific reconciler
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
					"Kind":            msg.Blueprint.Kind,
					"CronName":        cronName,
					"CronID":          addedCron.ID,
					"HandlerExecutor": handlerExecutorName,
					"Interval":        60,
				}).Info("Auto-created reconciliation cron for handler")

				existingCron = addedCron
			}
		} else {
			log.WithFields(log.Fields{
				"BlueprintName":   msg.Blueprint.Metadata.Name,
				"Kind":            msg.Blueprint.Kind,
				"CronName":        cronName,
				"HandlerExecutor": handlerExecutorName,
			}).Debug("Cron already exists for handler")
		}

		// Submit immediate reconciliation process for this specific blueprint
		if existingCron != nil {
			immediateProcess, err := h.createImmediateReconciliationProcess(msg.Blueprint, matchedSD, recoveredID, initiatorName)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":         err,
					"BlueprintName": msg.Blueprint.Metadata.Name,
				}).Warn("Failed to create immediate reconciliation process")
			} else {
				_, err = h.server.ProcessController().AddProcess(immediateProcess)
				if err != nil {
					log.WithFields(log.Fields{
						"Error":         err,
						"BlueprintName": msg.Blueprint.Metadata.Name,
					}).Warn("Failed to submit immediate reconciliation process")
				} else {
					log.WithFields(log.Fields{
						"BlueprintName": msg.Blueprint.Metadata.Name,
						"ProcessID":     immediateProcess.ID,
					}).Info("Submitted immediate reconciliation process for new blueprint")
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

	// Trigger immediate reconciliation by running the cron for this blueprint's handler
	if matchedSD != nil && matchedSD.Spec.Handler.ExecutorType != "" {
		// Get the handler executor name for this blueprint
		handlerExecutorName := ""
		if msg.Blueprint.Handler != nil {
			if msg.Blueprint.Handler.ExecutorName != "" {
				handlerExecutorName = msg.Blueprint.Handler.ExecutorName
			} else if len(msg.Blueprint.Handler.ExecutorNames) > 0 {
				handlerExecutorName = msg.Blueprint.Handler.ExecutorNames[0]
			}
		}

		// Find the cron for this specific handler
		cronName := "reconcile-" + msg.Blueprint.Kind
		if handlerExecutorName != "" {
			cronName = cronName + "-" + handlerExecutorName
		}

		crons, err := h.server.CronController().GetCrons(msg.Blueprint.Metadata.ColonyName, 1000)
		var existingCron *core.Cron
		if err == nil {
			for _, cron := range crons {
				if cron.Name == cronName {
					existingCron = cron
					break
				}
			}
		}

		if existingCron != nil {
			// Trigger the cron immediately
			_, err := h.server.CronController().RunCron(existingCron.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":           err,
					"BlueprintName":   msg.Blueprint.Metadata.Name,
					"CronName":        cronName,
					"HandlerExecutor": handlerExecutorName,
				}).Warn("Failed to trigger reconciliation cron after blueprint update")
			} else {
				log.WithFields(log.Fields{
					"BlueprintName":   msg.Blueprint.Metadata.Name,
					"Generation":      msg.Blueprint.Metadata.Generation,
					"CronName":        cronName,
					"CronID":          existingCron.ID,
					"HandlerExecutor": handlerExecutorName,
				}).Info("Triggered reconciliation cron immediately after blueprint update")
			}
		} else {
			log.WithFields(log.Fields{
				"BlueprintName":   msg.Blueprint.Metadata.Name,
				"CronName":        cronName,
				"HandlerExecutor": handlerExecutorName,
			}).Debug("No reconciliation cron found for handler")
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

	// Use consolidated cron name based on Kind
	cronName := "reconcile-" + blueprint.Kind

	// Remove the blueprint from database first
	err = h.server.BlueprintDB().RemoveBlueprintByName(msg.Namespace, msg.Name)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Check if there are any remaining blueprints of this Kind
	remainingBlueprints, err := h.server.BlueprintDB().GetBlueprintsByNamespaceAndKind(msg.Namespace, blueprint.Kind)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"Kind":  blueprint.Kind,
		}).Warn("Failed to check remaining blueprints")
		remainingBlueprints = []*core.Blueprint{} // Assume none remaining on error
	}

	// Only remove consolidated cron if no blueprints of this Kind remain
	if len(remainingBlueprints) == 0 {
		crons, err := h.server.CronController().GetCrons(msg.Namespace, 1000)
		if err == nil {
			for _, c := range crons {
				if c.Name == cronName {
					err = h.server.CronController().RemoveCron(c.ID)
					if err != nil {
						log.WithFields(log.Fields{
							"Error":    err,
							"Kind":     blueprint.Kind,
							"CronName": cronName,
						}).Warn("Failed to remove consolidated cron")
					} else {
						log.WithFields(log.Fields{
							"Kind":     blueprint.Kind,
							"CronName": cronName,
						}).Info("Auto-removed consolidated cron (last blueprint of Kind deleted)")
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
	} else {
		log.WithFields(log.Fields{
			"BlueprintName":       msg.Name,
			"Kind":                blueprint.Kind,
			"RemainingBlueprints": len(remainingBlueprints),
		}).Debug("Consolidated cron kept (other blueprints of same Kind exist)")
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
