package core

import "encoding/json"

type WorkflowSpec struct {
	ColonyName    string         `json:"colonyname"`
	FunctionSpecs []FunctionSpec `json:"functionspecs"`
}

func CreateWorkflowSpec(colonyName string) *WorkflowSpec {
	workflowSpec := &WorkflowSpec{ColonyName: colonyName}
	return workflowSpec
}

func (workflowSpec *WorkflowSpec) AddFunctionSpec(funcSpec *FunctionSpec) {
	workflowSpec.FunctionSpecs = append(workflowSpec.FunctionSpecs, *funcSpec)
}

func ConvertJSONToWorkflowSpec(jsonString string) (*WorkflowSpec, error) {
	var workflowSpec *WorkflowSpec
	err := json.Unmarshal([]byte(jsonString), &workflowSpec)
	if err != nil {
		return nil, err
	}

	return workflowSpec, nil
}

func (workflowSpec *WorkflowSpec) Equals(workflowSpec2 *WorkflowSpec) bool {
	same := true
	if workflowSpec.ColonyName != workflowSpec2.ColonyName {
		same = false
	}

	if workflowSpec.FunctionSpecs != nil && workflowSpec2.FunctionSpecs == nil {
		same = false
	} else if workflowSpec.FunctionSpecs == nil && workflowSpec2.FunctionSpecs != nil {
		same = false
	} else {
		counter := 0
		for _, funcSpec := range workflowSpec.FunctionSpecs {
			for _, funcSpec2 := range workflowSpec2.FunctionSpecs {
				if funcSpec.Equals(&funcSpec2) {
					counter++
				}
			}
		}
		if counter != len(workflowSpec.FunctionSpecs) && counter != len(workflowSpec2.FunctionSpecs) {
			same = false
		}
	}

	return same
}

func (workflowSpec *WorkflowSpec) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(workflowSpec)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
