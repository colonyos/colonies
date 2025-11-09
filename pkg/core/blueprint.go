package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/google/uuid"
)

// Blueprint is a generic container for blueprints
type Blueprint struct {
	ID       string                 `json:"blueprintid"`
	Kind     string                 `json:"kind"`
	Metadata BlueprintMetadata        `json:"metadata"`
	Spec     map[string]interface{} `json:"spec"`
	Status   map[string]interface{} `json:"status,omitempty"`
}

// BlueprintMetadata contains metadata for blueprints
type BlueprintMetadata struct {
	Name                      string            `json:"name"`
	Namespace                 string            `json:"namespace"`
	Labels                    map[string]string `json:"labels,omitempty"`
	Annotations               map[string]string `json:"annotations,omitempty"`
	Generation                int64             `json:"generation,omitempty"`
	CreatedAt                 time.Time         `json:"createdAt,omitempty"`
	UpdatedAt                 time.Time         `json:"updatedAt,omitempty"`
	LastReconciliationProcess string            `json:"lastReconciliationProcess,omitempty"`
	LastReconciliationTime    time.Time         `json:"lastReconciliationTime,omitempty"`
}

// BlueprintDefinition defines a blueprint type
type BlueprintDefinition struct {
	ID       string                  `json:"blueprintdefinitionid"`
	Kind     string                  `json:"kind"`
	Metadata BlueprintMetadata         `json:"metadata"`
	Spec     BlueprintDefinitionSpec   `json:"spec"`
}

// BlueprintDefinitionSpec defines the specification for a BlueprintDefinition
type BlueprintDefinitionSpec struct {
	Group   string                   `json:"group"`
	Version string                   `json:"version"`
	Names   BlueprintDefinitionNames   `json:"names"`
	Scope   string                   `json:"scope"` // "Namespaced" or "Cluster"
	Schema  *ValidationSchema        `json:"schema,omitempty"`
	Handler HandlerSpec              `json:"handler"`
}

// BlueprintDefinitionNames defines blueprint names
type BlueprintDefinitionNames struct {
	Kind       string   `json:"kind"`
	ListKind   string   `json:"listKind"`
	Singular   string   `json:"singular"`
	Plural     string   `json:"plural"`
	ShortNames []string `json:"shortNames,omitempty"`
}

// HandlerSpec defines the executor handler
type HandlerSpec struct {
	ExecutorType      string `json:"executorType"`
	FunctionName      string `json:"functionName"`
	ReconcileInterval int    `json:"reconcileInterval,omitempty"`
}

// ValidationSchema defines JSON Schema validation
type ValidationSchema struct {
	Type       string                    `json:"type,omitempty"`
	Properties map[string]SchemaProperty `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// SchemaProperty defines a schema property
type SchemaProperty struct {
	Type        string                    `json:"type,omitempty"`
	Description string                    `json:"description,omitempty"`
	Enum        []interface{}             `json:"enum,omitempty"`
	Default     interface{}               `json:"default,omitempty"`
	Properties  map[string]SchemaProperty `json:"properties,omitempty"`
	Items       *SchemaProperty           `json:"items,omitempty"`
}

// Reconciliation contains the old and new state of a blueprint with computed diff
type Reconciliation struct {
	Old    *Blueprint             `json:"old,omitempty"`
	New    *Blueprint             `json:"new,omitempty"`
	Diff   *BlueprintDiff         `json:"diff,omitempty"`
	Action ReconciliationAction `json:"action"`
}

// ReconciliationAction defines the CRUD operation to perform
type ReconciliationAction string

const (
	ReconciliationCreate ReconciliationAction = "create"
	ReconciliationUpdate ReconciliationAction = "update"
	ReconciliationDelete ReconciliationAction = "delete"
	ReconciliationNoop   ReconciliationAction = "noop"
)

// BlueprintDiff contains the differences between two blueprints
type BlueprintDiff struct {
	SpecChanges     []FieldChange `json:"specChanges,omitempty"`
	StatusChanges   []FieldChange `json:"statusChanges,omitempty"`
	MetadataChanges []FieldChange `json:"metadataChanges,omitempty"`
	HasChanges      bool          `json:"hasChanges"`
}

// FieldChange represents a change to a specific field
type FieldChange struct {
	Path     string      `json:"path"`
	OldValue interface{} `json:"oldValue,omitempty"`
	NewValue interface{} `json:"newValue,omitempty"`
	Type     ChangeType  `json:"type"`
}

// ChangeType defines the type of change
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeModified ChangeType = "modified"
	ChangeRemoved  ChangeType = "removed"
)

// CreateBlueprint creates a new blueprint
func CreateBlueprint(kind, name, namespace string) *Blueprint {
	uid := uuid.New()
	c := crypto.CreateCrypto()
	id := c.GenerateHash(uid.String())

	return &Blueprint{
		ID:   id,
		Kind: kind,
		Metadata: BlueprintMetadata{
			Name:        name,
			Namespace:   namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Generation:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Spec:   make(map[string]interface{}),
		Status: make(map[string]interface{}),
	}
}

// CreateBlueprintDefinition creates a new BlueprintDefinition
func CreateBlueprintDefinition(name, group, version, kind, plural, scope, executorType, functionName string) *BlueprintDefinition {
	uid := uuid.New()
	c := crypto.CreateCrypto()
	id := c.GenerateHash(uid.String())

	return &BlueprintDefinition{
		ID:   id,
		Kind: "BlueprintDefinition",
		Metadata: BlueprintMetadata{
			Name:        name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Generation:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Spec: BlueprintDefinitionSpec{
			Group:   group,
			Version: version,
			Names: BlueprintDefinitionNames{
				Kind:     kind,
				ListKind: kind + "List",
				Singular: strings.ToLower(kind),
				Plural:   plural,
			},
			Scope: scope,
			Handler: HandlerSpec{
				ExecutorType: executorType,
				FunctionName: functionName,
			},
		},
	}
}

// SetSpec sets a spec value and increments generation
func (r *Blueprint) SetSpec(key string, value interface{}) {
	if r.Spec == nil {
		r.Spec = make(map[string]interface{})
	}
	r.Spec[key] = value
	r.Metadata.Generation++
	r.Metadata.UpdatedAt = time.Now()
}

// GetSpec retrieves a spec value
func (r *Blueprint) GetSpec(key string) (interface{}, bool) {
	val, ok := r.Spec[key]
	return val, ok
}

// SetStatus sets a status value
func (r *Blueprint) SetStatus(key string, value interface{}) {
	if r.Status == nil {
		r.Status = make(map[string]interface{})
	}
	r.Status[key] = value
	r.Metadata.UpdatedAt = time.Now()
}

// GetStatus retrieves a status value
func (r *Blueprint) GetStatus(key string) (interface{}, bool) {
	val, ok := r.Status[key]
	return val, ok
}

// Validate validates the blueprint
func (r *Blueprint) Validate() error {
	if r.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if r.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if r.Metadata.Namespace == "" {
		return fmt.Errorf("metadata.namespace is required")
	}
	return nil
}


// ToJSON converts to JSON string
func (r *Blueprint) ToJSON() (string, error) {
	return toJSON(r)
}

// ValidateAgainstSD validates the Blueprint against its BlueprintDefinition schema
func (r *Blueprint) ValidateAgainstSD(sd *BlueprintDefinition) error {
	// Check that kind matches
	if r.Kind != sd.Spec.Names.Kind {
		return fmt.Errorf("kind mismatch: blueprint has '%s' but BlueprintDefinition defines '%s'", r.Kind, sd.Spec.Names.Kind)
	}

	// Validate against schema if one is defined
	if sd.Spec.Schema != nil {
		if err := validateAgainstSchema(r.Spec, sd.Spec.Schema); err != nil {
			return fmt.Errorf("spec validation failed: %w", err)
		}
	}

	return nil
}

// validateAgainstSchema validates data against a JSON schema
func validateAgainstSchema(data map[string]interface{}, schema *ValidationSchema) error {
	// Check required fields
	for _, required := range schema.Required {
		if _, ok := data[required]; !ok {
			return fmt.Errorf("required field '%s' is missing", required)
		}
	}

	// Validate each property
	for key, value := range data {
		if propSchema, ok := schema.Properties[key]; ok {
			if err := validateValue(key, value, &propSchema); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateValue validates a single value against its schema property
func validateValue(key string, value interface{}, prop *SchemaProperty) error {
	// Type validation
	if prop.Type != "" {
		if err := validateType(key, value, prop.Type); err != nil {
			return err
		}
	}

	// Enum validation
	if len(prop.Enum) > 0 {
		if err := validateEnum(key, value, prop.Enum); err != nil {
			return err
		}
	}

	// Nested object validation
	if prop.Type == "object" && prop.Properties != nil {
		objMap, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field '%s' must be an object", key)
		}
		for nestedKey, nestedValue := range objMap {
			if nestedProp, ok := prop.Properties[nestedKey]; ok {
				if err := validateValue(key+"."+nestedKey, nestedValue, &nestedProp); err != nil {
					return err
				}
			}
		}
	}

	// Array validation
	if prop.Type == "array" && prop.Items != nil {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("field '%s' must be an array", key)
		}
		for i, item := range arr {
			if err := validateValue(fmt.Sprintf("%s[%d]", key, i), item, prop.Items); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateType checks if value matches the expected type
func validateType(key string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' must be a string", key)
		}
	case "integer":
		// JSON unmarshaling produces float64 for numbers
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("field '%s' must be an integer", key)
			}
		case int, int32, int64:
			// Already an integer
		default:
			return fmt.Errorf("field '%s' must be an integer", key)
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// Valid number types
		default:
			return fmt.Errorf("field '%s' must be a number", key)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be a boolean", key)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("field '%s' must be an object", key)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("field '%s' must be an array", key)
		}
	}
	return nil
}

// validateEnum checks if value is in the allowed enum values
func validateEnum(key string, value interface{}, enum []interface{}) error {
	for _, allowed := range enum {
		if value == allowed {
			return nil
		}
	}
	return fmt.Errorf("field '%s' has invalid value '%v', must be one of %v", key, value, enum)
}

// Validate validates the BlueprintDefinition
func (sd *BlueprintDefinition) Validate() error {
	if sd.Kind != "BlueprintDefinition" {
		return fmt.Errorf("kind must be BlueprintDefinition")
	}
	if sd.Spec.Group == "" {
		return fmt.Errorf("spec.group is required")
	}
	if sd.Spec.Version == "" {
		return fmt.Errorf("spec.version is required")
	}
	if sd.Spec.Names.Kind == "" {
		return fmt.Errorf("spec.names.kind is required")
	}
	if sd.Spec.Names.Plural == "" {
		return fmt.Errorf("spec.names.plural is required")
	}
	if sd.Spec.Scope != "Namespaced" && sd.Spec.Scope != "Cluster" {
		return fmt.Errorf("spec.scope must be 'Namespaced' or 'Cluster'")
	}
	if sd.Spec.Handler.ExecutorType == "" {
		return fmt.Errorf("spec.handler.executorType is required")
	}
	if sd.Spec.Handler.FunctionName == "" {
		return fmt.Errorf("spec.handler.functionName is required")
	}
	if sd.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	return nil
}

// GetAPIVersion returns the full API version
func (sd *BlueprintDefinition) GetAPIVersion() string {
	return sd.Spec.Group + "/" + sd.Spec.Version
}

// ToJSON converts to JSON string
func (sd *BlueprintDefinition) ToJSON() (string, error) {
	return toJSON(sd)
}

// ConvertJSONToBlueprint parses JSON to Blueprint
func ConvertJSONToBlueprint(jsonString string) (*Blueprint, error) {
	var blueprint Blueprint
	if err := json.Unmarshal([]byte(jsonString), &blueprint); err != nil {
		return nil, err
	}
	initBlueprint(&blueprint)
	return &blueprint, nil
}

// ConvertJSONToBlueprintDefinition parses JSON to BlueprintDefinition
func ConvertJSONToBlueprintDefinition(jsonString string) (*BlueprintDefinition, error) {
	var sd BlueprintDefinition
	if err := json.Unmarshal([]byte(jsonString), &sd); err != nil {
		return nil, err
	}
	initMetadata(&sd.Metadata)
	return &sd, nil
}

// Helper functions

func initBlueprint(r *Blueprint) {
	initMetadata(&r.Metadata)
	if r.Spec == nil {
		r.Spec = make(map[string]interface{})
	}
	if r.Status == nil {
		r.Status = make(map[string]interface{})
	}
}

func initMetadata(m *BlueprintMetadata) {
	if m.Labels == nil {
		m.Labels = make(map[string]string)
	}
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
}

func toJSON(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Reconciliation helper methods

// CreateReconciliation creates a Reconciliation from old and new blueprints
func CreateReconciliation(old, new *Blueprint) *Reconciliation {
	reconciliation := &Reconciliation{
		Old: old,
		New: new,
	}

	// Determine action
	if old == nil && new != nil {
		reconciliation.Action = ReconciliationCreate
	} else if old != nil && new == nil {
		reconciliation.Action = ReconciliationDelete
	} else if old != nil && new != nil {
		// Both exist, compute diff
		diff := old.Diff(new)
		reconciliation.Diff = diff
		if diff.HasChanges {
			reconciliation.Action = ReconciliationUpdate
		} else {
			reconciliation.Action = ReconciliationNoop
		}
	} else {
		// Both nil - noop
		reconciliation.Action = ReconciliationNoop
	}

	return reconciliation
}

// Diff computes the differences between this blueprint and another
func (r *Blueprint) Diff(other *Blueprint) *BlueprintDiff {
	if other == nil {
		return &BlueprintDiff{HasChanges: false}
	}

	diff := &BlueprintDiff{
		SpecChanges:     []FieldChange{},
		StatusChanges:   []FieldChange{},
		MetadataChanges: []FieldChange{},
		HasChanges:      false,
	}

	// Compute spec changes
	specChanges := computeMapDiff("spec", r.Spec, other.Spec)
	diff.SpecChanges = specChanges

	// Compute status changes
	statusChanges := computeMapDiff("status", r.Status, other.Status)
	diff.StatusChanges = statusChanges

	// Compute metadata changes (only labels and annotations)
	metadataChanges := []FieldChange{}
	labelChanges := computeMapDiff("metadata.labels", convertStringMap(r.Metadata.Labels), convertStringMap(other.Metadata.Labels))
	annotationChanges := computeMapDiff("metadata.annotations", convertStringMap(r.Metadata.Annotations), convertStringMap(other.Metadata.Annotations))
	metadataChanges = append(metadataChanges, labelChanges...)
	metadataChanges = append(metadataChanges, annotationChanges...)
	diff.MetadataChanges = metadataChanges

	// Set HasChanges flag
	diff.HasChanges = len(diff.SpecChanges) > 0 || len(diff.StatusChanges) > 0 || len(diff.MetadataChanges) > 0

	return diff
}

// computeMapDiff computes the differences between two maps
func computeMapDiff(prefix string, oldMap, newMap map[string]interface{}) []FieldChange {
	changes := []FieldChange{}

	// Handle nil maps
	if oldMap == nil {
		oldMap = make(map[string]interface{})
	}
	if newMap == nil {
		newMap = make(map[string]interface{})
	}

	// Find added and modified fields
	for key, newValue := range newMap {
		path := prefix + "." + key
		if oldValue, exists := oldMap[key]; exists {
			// Field exists in both - check if modified
			if !deepEqual(oldValue, newValue) {
				changes = append(changes, FieldChange{
					Path:     path,
					OldValue: oldValue,
					NewValue: newValue,
					Type:     ChangeModified,
				})
			}
		} else {
			// Field is new
			changes = append(changes, FieldChange{
				Path:     path,
				OldValue: nil,
				NewValue: newValue,
				Type:     ChangeAdded,
			})
		}
	}

	// Find removed fields
	for key, oldValue := range oldMap {
		path := prefix + "." + key
		if _, exists := newMap[key]; !exists {
			changes = append(changes, FieldChange{
				Path:     path,
				OldValue: oldValue,
				NewValue: nil,
				Type:     ChangeRemoved,
			})
		}
	}

	return changes
}

// convertStringMap converts map[string]string to map[string]interface{}
func convertStringMap(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// deepEqual performs deep equality check
func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// BlueprintDiff helper methods

// HasFieldChange checks if a specific field has changed
func (sd *BlueprintDiff) HasFieldChange(path string) bool {
	allChanges := append(sd.SpecChanges, sd.StatusChanges...)
	allChanges = append(allChanges, sd.MetadataChanges...)

	for _, change := range allChanges {
		if change.Path == path {
			return true
		}
	}
	return false
}

// GetFieldChange retrieves a specific field change
func (sd *BlueprintDiff) GetFieldChange(path string) *FieldChange {
	allChanges := append(sd.SpecChanges, sd.StatusChanges...)
	allChanges = append(allChanges, sd.MetadataChanges...)

	for _, change := range allChanges {
		if change.Path == path {
			return &change
		}
	}
	return nil
}

// OnlyMetadataChanged returns true if only metadata changed
func (sd *BlueprintDiff) OnlyMetadataChanged() bool {
	return len(sd.MetadataChanges) > 0 && len(sd.SpecChanges) == 0 && len(sd.StatusChanges) == 0
}

// OnlyStatusChanged returns true if only status changed
func (sd *BlueprintDiff) OnlyStatusChanged() bool {
	return len(sd.StatusChanges) > 0 && len(sd.SpecChanges) == 0 && len(sd.MetadataChanges) == 0
}

// ConvertBlueprintArrayToJSON converts a slice of Blueprints to JSON
func ConvertBlueprintArrayToJSON(blueprints []*Blueprint) (string, error) {
	jsonBytes, err := json.MarshalIndent(blueprints, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ConvertJSONToBlueprintArray parses JSON to a slice of Blueprints
func ConvertJSONToBlueprintArray(jsonString string) ([]*Blueprint, error) {
	var blueprints []*Blueprint
	if err := json.Unmarshal([]byte(jsonString), &blueprints); err != nil {
		return nil, err
	}
	for _, blueprint := range blueprints {
		initBlueprint(blueprint)
	}
	return blueprints, nil
}

// ConvertBlueprintDefinitionArrayToJSON converts a slice of BlueprintDefinitions to JSON
func ConvertBlueprintDefinitionArrayToJSON(sds []*BlueprintDefinition) (string, error) {
	jsonBytes, err := json.MarshalIndent(sds, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ConvertJSONToBlueprintDefinitionArray parses JSON to a slice of BlueprintDefinitions
func ConvertJSONToBlueprintDefinitionArray(jsonString string) ([]*BlueprintDefinition, error) {
	var sds []*BlueprintDefinition
	if err := json.Unmarshal([]byte(jsonString), &sds); err != nil {
		return nil, err
	}
	return sds, nil
}

// ValidateBlueprintAgainstSchema validates a Blueprint's spec against a BlueprintDefinition's schema
func ValidateBlueprintAgainstSchema(blueprint *Blueprint, schema *ValidationSchema) error {
	if schema == nil {
		return nil // No schema means no validation
	}

	// Check required fields
	for _, requiredField := range schema.Required {
		if _, ok := blueprint.Spec[requiredField]; !ok {
			return fmt.Errorf("required field '%s' is missing", requiredField)
		}
	}

	// Validate each field in the spec
	for fieldName, fieldValue := range blueprint.Spec {
		if schemaProp, ok := schema.Properties[fieldName]; ok {
			if err := validateField(fieldName, fieldValue, &schemaProp); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateField validates a single field against its schema property
func validateField(fieldName string, value interface{}, prop *SchemaProperty) error {
	if prop == nil {
		return nil
	}

	// Type validation
	switch prop.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' must be a string, got %T", fieldName, value)
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// Valid number types
		default:
			return fmt.Errorf("field '%s' must be a number, got %T", fieldName, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be a boolean, got %T", fieldName, value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("field '%s' must be an object, got %T", fieldName, value)
		}
	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("field '%s' must be an array, got %T", fieldName, value)
		}
		// Validate array items if schema is provided
		if prop.Items != nil {
			for i, item := range arr {
				if err := validateField(fmt.Sprintf("%s[%d]", fieldName, i), item, prop.Items); err != nil {
					return err
				}
			}
		}
	}

	// Enum validation
	if len(prop.Enum) > 0 {
		found := false
		for _, enumValue := range prop.Enum {
			if value == enumValue {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field '%s' must be one of %v, got %v", fieldName, prop.Enum, value)
		}
	}

	// Validate nested object properties
	if prop.Type == "object" && len(prop.Properties) > 0 {
		objValue, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field '%s' must be an object", fieldName)
		}
		for nestedName, nestedValue := range objValue {
			if nestedProp, ok := prop.Properties[nestedName]; ok {
				if err := validateField(fmt.Sprintf("%s.%s", fieldName, nestedName), nestedValue, &nestedProp); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// BlueprintHistory represents a historical snapshot of a blueprint
type BlueprintHistory struct {
	ID        string                 `json:"historyid"`
	BlueprintID string                 `json:"blueprintid"`
	Kind      string                 `json:"kind"`
	Namespace string                 `json:"namespace"`
	Name      string                 `json:"name"`
	Generation int64                 `json:"generation"`
	Spec       map[string]interface{} `json:"spec"`
	Status     map[string]interface{} `json:"status,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	ChangedBy  string                 `json:"changedby"` // Executor or User ID
	ChangeType string                 `json:"changetype"` // "create", "update", "delete"
}

// CreateBlueprintHistory creates a new BlueprintHistory from a Blueprint
func CreateBlueprintHistory(blueprint *Blueprint, changedBy string, changeType string) *BlueprintHistory {
	return &BlueprintHistory{
		ID:         uuid.New().String(),
		BlueprintID:  blueprint.ID,
		Kind:       blueprint.Kind,
		Namespace:  blueprint.Metadata.Namespace,
		Name:       blueprint.Metadata.Name,
		Generation: blueprint.Metadata.Generation,
		Spec:       copyMap(blueprint.Spec),
		Status:     copyMap(blueprint.Status),
		Timestamp:  time.Now(),
		ChangedBy:  changedBy,
		ChangeType: changeType,
	}
}

// copyMap creates a deep copy of a map[string]interface{}
func copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = copyMap(val)
		case []interface{}:
			result[k] = copySlice(val)
		default:
			result[k] = v
		}
	}
	return result
}

// copySlice creates a deep copy of a []interface{}
func copySlice(s []interface{}) []interface{} {
	if s == nil {
		return nil
	}
	result := make([]interface{}, len(s))
	for i, v := range s {
		switch val := v.(type) {
		case map[string]interface{}:
			result[i] = copyMap(val)
		case []interface{}:
			result[i] = copySlice(val)
		default:
			result[i] = v
		}
	}
	return result
}

// ToJSON converts BlueprintHistory to JSON
func (sh *BlueprintHistory) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(sh)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ConvertJSONToBlueprintHistory converts JSON to BlueprintHistory
func ConvertJSONToBlueprintHistory(jsonString string) (*BlueprintHistory, error) {
	var history BlueprintHistory
	err := json.Unmarshal([]byte(jsonString), &history)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ConvertJSONToBlueprintHistoryArray converts JSON array to BlueprintHistory array
func ConvertJSONToBlueprintHistoryArray(jsonString string) ([]*BlueprintHistory, error) {
	var histories []*BlueprintHistory
	err := json.Unmarshal([]byte(jsonString), &histories)
	if err != nil {
		return nil, err
	}
	return histories, nil
}

// ConvertBlueprintHistoryArrayToJSON converts BlueprintHistory array to JSON
func ConvertBlueprintHistoryArrayToJSON(histories []*BlueprintHistory) (string, error) {
	jsonBytes, err := json.Marshal(histories)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
