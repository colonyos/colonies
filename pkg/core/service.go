package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/google/uuid"
)

// Service is a generic container for services
type Service struct {
	ID       string                 `json:"resourceid"`
	Kind     string                 `json:"kind"`
	Metadata ResourceMetadata       `json:"metadata"`
	Spec     map[string]interface{} `json:"spec"`
	Status   map[string]interface{} `json:"status,omitempty"`
}

// ResourceMetadata contains metadata for services
type ResourceMetadata struct {
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

// ResourceDefinition defines a service type
type ResourceDefinition struct {
	ID       string                 `json:"resourcedefinitionid"`
	Kind     string                 `json:"kind"`
	Metadata ResourceMetadata       `json:"metadata"`
	Spec     ResourceDefinitionSpec `json:"spec"`
}

// ResourceDefinitionSpec defines the specification for a ResourceDefinition
type ResourceDefinitionSpec struct {
	Group   string            `json:"group"`
	Version string            `json:"version"`
	Names   ResourceDefinitionNames           `json:"names"`
	Scope   string            `json:"scope"` // "Namespaced" or "Cluster"
	Schema  *ValidationSchema `json:"schema,omitempty"`
	Handler HandlerSpec       `json:"handler"`
}

// ResourceDefinitionNames defines service names
type ResourceDefinitionNames struct {
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

// Reconciliation contains the old and new state of a service with computed diff
type Reconciliation struct {
	Old    *Service            `json:"old,omitempty"`
	New    *Service            `json:"new,omitempty"`
	Diff   *ResourceDiff        `json:"diff,omitempty"`
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

// ResourceDiff contains the differences between two services
type ResourceDiff struct {
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

// CreateResource creates a new service
func CreateResource(kind, name, namespace string) *Service {
	uid := uuid.New()
	c := crypto.CreateCrypto()
	id := c.GenerateHash(uid.String())

	return &Service{
		ID:   id,
		Kind: kind,
		Metadata: ResourceMetadata{
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

// CreateResourceDefinition creates a new ResourceDefinition
func CreateResourceDefinition(name, group, version, kind, plural, scope, executorType, functionName string) *ResourceDefinition {
	uid := uuid.New()
	c := crypto.CreateCrypto()
	id := c.GenerateHash(uid.String())

	return &ResourceDefinition{
		ID:   id,
		Kind: "ResourceDefinition",
		Metadata: ResourceMetadata{
			Name:        name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Generation:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Spec: ResourceDefinitionSpec{
			Group:   group,
			Version: version,
			Names: ResourceDefinitionNames{
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
func (r *Service) SetSpec(key string, value interface{}) {
	if r.Spec == nil {
		r.Spec = make(map[string]interface{})
	}
	r.Spec[key] = value
	r.Metadata.Generation++
	r.Metadata.UpdatedAt = time.Now()
}

// GetSpec retrieves a spec value
func (r *Service) GetSpec(key string) (interface{}, bool) {
	val, ok := r.Spec[key]
	return val, ok
}

// SetStatus sets a status value
func (r *Service) SetStatus(key string, value interface{}) {
	if r.Status == nil {
		r.Status = make(map[string]interface{})
	}
	r.Status[key] = value
	r.Metadata.UpdatedAt = time.Now()
}

// GetStatus retrieves a status value
func (r *Service) GetStatus(key string) (interface{}, bool) {
	val, ok := r.Status[key]
	return val, ok
}

// Validate validates the service
func (r *Service) Validate() error {
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
func (r *Service) ToJSON() (string, error) {
	return toJSON(r)
}

// ValidateAgainstRD validates the Service against its ResourceDefinition schema
func (r *Service) ValidateAgainstRD(rd *ResourceDefinition) error {
	// Check that kind matches
	if r.Kind != rd.Spec.Names.Kind {
		return fmt.Errorf("kind mismatch: service has '%s' but ResourceDefinition defines '%s'", r.Kind, rd.Spec.Names.Kind)
	}

	// Validate against schema if one is defined
	if rd.Spec.Schema != nil {
		if err := validateAgainstSchema(r.Spec, rd.Spec.Schema); err != nil {
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

// Validate validates the ResourceDefinition
func (rd *ResourceDefinition) Validate() error {
	if rd.Kind != "ResourceDefinition" {
		return fmt.Errorf("kind must be ResourceDefinition")
	}
	if rd.Spec.Group == "" {
		return fmt.Errorf("spec.group is required")
	}
	if rd.Spec.Version == "" {
		return fmt.Errorf("spec.version is required")
	}
	if rd.Spec.Names.Kind == "" {
		return fmt.Errorf("spec.names.kind is required")
	}
	if rd.Spec.Names.Plural == "" {
		return fmt.Errorf("spec.names.plural is required")
	}
	if rd.Spec.Scope != "Namespaced" && rd.Spec.Scope != "Cluster" {
		return fmt.Errorf("spec.scope must be 'Namespaced' or 'Cluster'")
	}
	if rd.Spec.Handler.ExecutorType == "" {
		return fmt.Errorf("spec.handler.executorType is required")
	}
	if rd.Spec.Handler.FunctionName == "" {
		return fmt.Errorf("spec.handler.functionName is required")
	}
	if rd.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	return nil
}

// GetAPIVersion returns the full API version
func (rd *ResourceDefinition) GetAPIVersion() string {
	return rd.Spec.Group + "/" + rd.Spec.Version
}

// ToJSON converts to JSON string
func (rd *ResourceDefinition) ToJSON() (string, error) {
	return toJSON(rd)
}

// ConvertJSONToResource parses JSON to Service
func ConvertJSONToResource(jsonString string) (*Service, error) {
	var service Service
	if err := json.Unmarshal([]byte(jsonString), &service); err != nil {
		return nil, err
	}
	initResource(&service)
	return &service, nil
}

// ConvertJSONToResourceDefinition parses JSON to ResourceDefinition
func ConvertJSONToResourceDefinition(jsonString string) (*ResourceDefinition, error) {
	var rd ResourceDefinition
	if err := json.Unmarshal([]byte(jsonString), &rd); err != nil {
		return nil, err
	}
	initMetadata(&rd.Metadata)
	return &rd, nil
}

// Helper functions

func initResource(r *Service) {
	initMetadata(&r.Metadata)
	if r.Spec == nil {
		r.Spec = make(map[string]interface{})
	}
	if r.Status == nil {
		r.Status = make(map[string]interface{})
	}
}

func initMetadata(m *ResourceMetadata) {
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

// CreateReconciliation creates a Reconciliation from old and new services
func CreateReconciliation(old, new *Service) *Reconciliation {
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

// Diff computes the differences between this service and another
func (r *Service) Diff(other *Service) *ResourceDiff {
	if other == nil {
		return &ResourceDiff{HasChanges: false}
	}

	diff := &ResourceDiff{
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

// ResourceDiff helper methods

// HasFieldChange checks if a specific field has changed
func (rd *ResourceDiff) HasFieldChange(path string) bool {
	allChanges := append(rd.SpecChanges, rd.StatusChanges...)
	allChanges = append(allChanges, rd.MetadataChanges...)

	for _, change := range allChanges {
		if change.Path == path {
			return true
		}
	}
	return false
}

// GetFieldChange retrieves a specific field change
func (rd *ResourceDiff) GetFieldChange(path string) *FieldChange {
	allChanges := append(rd.SpecChanges, rd.StatusChanges...)
	allChanges = append(allChanges, rd.MetadataChanges...)

	for _, change := range allChanges {
		if change.Path == path {
			return &change
		}
	}
	return nil
}

// OnlyMetadataChanged returns true if only metadata changed
func (rd *ResourceDiff) OnlyMetadataChanged() bool {
	return len(rd.MetadataChanges) > 0 && len(rd.SpecChanges) == 0 && len(rd.StatusChanges) == 0
}

// OnlyStatusChanged returns true if only status changed
func (rd *ResourceDiff) OnlyStatusChanged() bool {
	return len(rd.StatusChanges) > 0 && len(rd.SpecChanges) == 0 && len(rd.MetadataChanges) == 0
}

// ConvertResourceArrayToJSON converts a slice of Services to JSON
func ConvertResourceArrayToJSON(services []*Service) (string, error) {
	jsonBytes, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ConvertJSONToResourceArray parses JSON to a slice of Services
func ConvertJSONToResourceArray(jsonString string) ([]*Service, error) {
	var services []*Service
	if err := json.Unmarshal([]byte(jsonString), &services); err != nil {
		return nil, err
	}
	for _, service := range services {
		initResource(service)
	}
	return services, nil
}

// ConvertResourceDefinitionArrayToJSON converts a slice of ResourceDefinitions to JSON
func ConvertResourceDefinitionArrayToJSON(rds []*ResourceDefinition) (string, error) {
	jsonBytes, err := json.MarshalIndent(rds, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ConvertJSONToResourceDefinitionArray parses JSON to a slice of ResourceDefinitions
func ConvertJSONToResourceDefinitionArray(jsonString string) ([]*ResourceDefinition, error) {
	var rds []*ResourceDefinition
	if err := json.Unmarshal([]byte(jsonString), &rds); err != nil {
		return nil, err
	}
	return rds, nil
}

// ValidateResourceAgainstSchema validates a Service's spec against a ResourceDefinition's schema
func ValidateResourceAgainstSchema(service *Service, schema *ValidationSchema) error {
	if schema == nil {
		return nil // No schema means no validation
	}

	// Check required fields
	for _, requiredField := range schema.Required {
		if _, ok := service.Spec[requiredField]; !ok {
			return fmt.Errorf("required field '%s' is missing", requiredField)
		}
	}

	// Validate each field in the spec
	for fieldName, fieldValue := range service.Spec {
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

// ResourceHistory represents a historical snapshot of a service
type ResourceHistory struct {
	ID         string                 `json:"historyid"`
	ResourceID string                 `json:"resourceid"`
	Kind       string                 `json:"kind"`
	Namespace  string                 `json:"namespace"`
	Name       string                 `json:"name"`
	Generation int64                  `json:"generation"`
	Spec       map[string]interface{} `json:"spec"`
	Status     map[string]interface{} `json:"status,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	ChangedBy  string                 `json:"changedby"` // Executor or User ID
	ChangeType string                 `json:"changetype"` // "create", "update", "delete"
}

// CreateResourceHistory creates a new ResourceHistory from a Service
func CreateResourceHistory(service *Service, changedBy string, changeType string) *ResourceHistory {
	return &ResourceHistory{
		ID:         uuid.New().String(),
		ResourceID: service.ID,
		Kind:       service.Kind,
		Namespace:  service.Metadata.Namespace,
		Name:       service.Metadata.Name,
		Generation: service.Metadata.Generation,
		Spec:       copyMap(service.Spec),
		Status:     copyMap(service.Status),
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

// ToJSON converts ResourceHistory to JSON
func (rh *ResourceHistory) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(rh)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ConvertJSONToResourceHistory converts JSON to ResourceHistory
func ConvertJSONToResourceHistory(jsonString string) (*ResourceHistory, error) {
	var history ResourceHistory
	err := json.Unmarshal([]byte(jsonString), &history)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ConvertJSONToResourceHistoryArray converts JSON array to ResourceHistory array
func ConvertJSONToResourceHistoryArray(jsonString string) ([]*ResourceHistory, error) {
	var histories []*ResourceHistory
	err := json.Unmarshal([]byte(jsonString), &histories)
	if err != nil {
		return nil, err
	}
	return histories, nil
}

// ConvertResourceHistoryArrayToJSON converts ResourceHistory array to JSON
func ConvertResourceHistoryArrayToJSON(histories []*ResourceHistory) (string, error) {
	jsonBytes, err := json.Marshal(histories)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
