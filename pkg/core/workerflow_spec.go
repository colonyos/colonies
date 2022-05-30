package core

import "encoding/json"

type WorkflowSpec struct {
	RuntimeGroup bool           `json:"runtimegroup"`
	ColonyID     string         `json:"colonyid"`
	ProcessSpecs []*ProcessSpec `json:"processspecs"`
}

func CreateWorkflowSpec(colonyID string, runtimeGroup bool) *WorkflowSpec {
	workflowSpec := &WorkflowSpec{ColonyID: colonyID, RuntimeGroup: runtimeGroup}
	return workflowSpec
}

func (workflowSpec *WorkflowSpec) AddProcessSpec(processSpec *ProcessSpec) {
	workflowSpec.ProcessSpecs = append(workflowSpec.ProcessSpecs, processSpec)
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
	if workflowSpec.ColonyID != workflowSpec2.ColonyID {
		same = false
	}

	if workflowSpec.RuntimeGroup != workflowSpec2.RuntimeGroup {
		same = false
	}

	if workflowSpec.ProcessSpecs != nil && workflowSpec2.ProcessSpecs == nil {
		same = false
	} else if workflowSpec.ProcessSpecs == nil && workflowSpec2.ProcessSpecs != nil {
		same = false
	} else {
		counter := 0
		for _, processSpec := range workflowSpec.ProcessSpecs {
			for _, processSpec2 := range workflowSpec2.ProcessSpecs {
				if processSpec.Equals(processSpec2) {
					counter++
				}
			}
		}
		if counter != len(workflowSpec.ProcessSpecs) && counter != len(workflowSpec2.ProcessSpecs) {
			same = false
		}
	}

	return same
}

func (workflowSpec *WorkflowSpec) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(workflowSpec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
