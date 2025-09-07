package adapter

import (
	"context"
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// ProcessGraphDatabase interface implementation

func (a *ColonyOSAdapter) AddProcessGraph(processgraph *core.ProcessGraph) error {
	doc := &memdb.VelocityDocument{
		ID:     processgraph.ID,
		Fields: a.processGraphToFields(processgraph),
	}
	
	return a.db.Insert(context.Background(), ProcessGraphsCollection, doc)
}

func (a *ColonyOSAdapter) GetProcessGraphs(colonyName string) ([]*core.ProcessGraph, error) {
	result, err := a.db.List(context.Background(), ProcessGraphsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var graphs []*core.ProcessGraph
	for _, doc := range result {
		graph, err := a.fieldsToProcessGraph(doc.Fields)
		if err == nil && graph.ColonyName == colonyName {
			graphs = append(graphs, graph)
		}
	}
	
	return graphs, nil
}

func (a *ColonyOSAdapter) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	doc, err := a.db.Get(context.Background(), ProcessGraphsCollection, processGraphID)
	if err != nil {
		return nil, err
	}
	
	return a.fieldsToProcessGraph(doc.Fields)
}

func (a *ColonyOSAdapter) FindProcessGraphsByColonyName(colonyName string, seconds int, state int) ([]*core.ProcessGraph, error) {
	graphs, err := a.GetProcessGraphs(colonyName)
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.ProcessGraph
	for _, graph := range graphs {
		if graph.State == state {
			filtered = append(filtered, graph)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) FindAllWaitingProcessGraphs() ([]*core.ProcessGraph, error) {
	return a.findProcessGraphsByState(core.WAITING)
}

func (a *ColonyOSAdapter) FindAllRunningProcessGraphs() ([]*core.ProcessGraph, error) {
	return a.findProcessGraphsByState(core.RUNNING)
}

func (a *ColonyOSAdapter) findProcessGraphsByState(state int) ([]*core.ProcessGraph, error) {
	result, err := a.db.List(context.Background(), ProcessGraphsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.ProcessGraph
	for _, doc := range result {
		graph, err := a.fieldsToProcessGraph(doc.Fields)
		if err == nil && graph.State == state {
			filtered = append(filtered, graph)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) RemoveProcessGraphByID(processGraphID string) error {
	return a.db.Delete(context.Background(), ProcessGraphsCollection, processGraphID)
}

func (a *ColonyOSAdapter) RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error {
	return a.removeProcessGraphsByColonyAndState(colonyName, core.WAITING)
}

func (a *ColonyOSAdapter) RemoveAllRunningProcessGraphsByColonyName(colonyName string) error {
	return a.removeProcessGraphsByColonyAndState(colonyName, core.RUNNING)
}

func (a *ColonyOSAdapter) RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error {
	return a.removeProcessGraphsByColonyAndState(colonyName, core.SUCCESS)
}

func (a *ColonyOSAdapter) RemoveAllFailedProcessGraphsByColonyName(colonyName string) error {
	return a.removeProcessGraphsByColonyAndState(colonyName, core.FAILED)
}

func (a *ColonyOSAdapter) RemoveAllProcessGraphsByColonyName(colonyName string) error {
	graphs, err := a.GetProcessGraphs(colonyName)
	if err != nil {
		return err
	}
	
	for _, graph := range graphs {
		if err := a.db.Delete(context.Background(), ProcessGraphsCollection, graph.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) removeProcessGraphsByColonyAndState(colonyName string, state int) error {
	graphs, err := a.GetProcessGraphs(colonyName)
	if err != nil {
		return err
	}
	
	for _, graph := range graphs {
		if graph.State == state {
			if err := a.db.Delete(context.Background(), ProcessGraphsCollection, graph.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) SetProcessGraphState(processGraphID string, state int) error {
	fields := map[string]interface{}{
		"state": state,
	}
	
	_, err := a.db.Update(context.Background(), ProcessGraphsCollection, processGraphID, fields)
	return err
}

func (a *ColonyOSAdapter) CountProcessGraphs() (int, error) {
	result, err := a.db.List(context.Background(), ProcessGraphsCollection, 10000, 0)
	if err != nil {
		return 0, err
	}
	return len(result), nil
}

func (a *ColonyOSAdapter) CountWaitingProcessGraphs() (int, error) {
	return a.countProcessGraphsByState(core.WAITING)
}

func (a *ColonyOSAdapter) CountRunningProcessGraphs() (int, error) {
	return a.countProcessGraphsByState(core.RUNNING)
}

func (a *ColonyOSAdapter) CountSuccessfulProcessGraphs() (int, error) {
	return a.countProcessGraphsByState(core.SUCCESS)
}

func (a *ColonyOSAdapter) CountFailedProcessGraphs() (int, error) {
	return a.countProcessGraphsByState(core.FAILED)
}

func (a *ColonyOSAdapter) CountWaitingProcessGraphsByColonyName(colonyName string) (int, error) {
	return a.countProcessGraphsByColonyAndState(colonyName, core.WAITING)
}

func (a *ColonyOSAdapter) CountRunningProcessGraphsByColonyName(colonyName string) (int, error) {
	return a.countProcessGraphsByColonyAndState(colonyName, core.RUNNING)
}

func (a *ColonyOSAdapter) CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error) {
	return a.countProcessGraphsByColonyAndState(colonyName, core.SUCCESS)
}

func (a *ColonyOSAdapter) CountFailedProcessGraphsByColonyName(colonyName string) (int, error) {
	return a.countProcessGraphsByColonyAndState(colonyName, core.FAILED)
}

func (a *ColonyOSAdapter) countProcessGraphsByState(state int) (int, error) {
	graphs, err := a.findProcessGraphsByState(state)
	if err != nil {
		return 0, err
	}
	return len(graphs), nil
}

func (a *ColonyOSAdapter) countProcessGraphsByColonyAndState(colonyName string, state int) (int, error) {
	graphs, err := a.FindProcessGraphsByColonyName(colonyName, 0, state)
	if err != nil {
		return 0, err
	}
	return len(graphs), nil
}

func (a *ColonyOSAdapter) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return a.findProcessGraphsByColonyAndState(colonyName, core.WAITING, count)
}

func (a *ColonyOSAdapter) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return a.findProcessGraphsByColonyAndState(colonyName, core.RUNNING, count)
}

func (a *ColonyOSAdapter) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return a.findProcessGraphsByColonyAndState(colonyName, core.SUCCESS, count)
}

func (a *ColonyOSAdapter) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return a.findProcessGraphsByColonyAndState(colonyName, core.FAILED, count)
}

func (a *ColonyOSAdapter) findProcessGraphsByColonyAndState(colonyName string, state int, count int) ([]*core.ProcessGraph, error) {
	graphs, err := a.GetProcessGraphs(colonyName)
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.ProcessGraph
	for _, graph := range graphs {
		if len(filtered) >= count && count > 0 {
			break
		}
		if graph.State == state {
			filtered = append(filtered, graph)
		}
	}
	
	return filtered, nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) processGraphToFields(graph *core.ProcessGraph) map[string]interface{} {
	fields := map[string]interface{}{
		"id":             graph.ID,
		"initiator_id":   graph.InitiatorID,
		"initiator_name": graph.InitiatorName,
		"colony_name":    graph.ColonyName,
		"state":          graph.State,
		"submission_time": graph.SubmissionTime,
		"start_time":     graph.StartTime,
		"end_time":       graph.EndTime,
	}
	
	// Serialize complex fields
	if rootsData, err := json.Marshal(graph.Roots); err == nil {
		fields["roots"] = string(rootsData)
	}
	if processIDsData, err := json.Marshal(graph.ProcessIDs); err == nil {
		fields["process_ids"] = string(processIDsData)
	}
	if nodesData, err := json.Marshal(graph.Nodes); err == nil {
		fields["nodes"] = string(nodesData)
	}
	if edgesData, err := json.Marshal(graph.Edges); err == nil {
		fields["edges"] = string(edgesData)
	}
	
	return fields
}

func (a *ColonyOSAdapter) fieldsToProcessGraph(fields map[string]interface{}) (*core.ProcessGraph, error) {
	graph := &core.ProcessGraph{}
	
	if id, ok := fields["id"].(string); ok {
		graph.ID = id
	}
	if initiatorID, ok := fields["initiator_id"].(string); ok {
		graph.InitiatorID = initiatorID
	}
	if initiatorName, ok := fields["initiator_name"].(string); ok {
		graph.InitiatorName = initiatorName
	}
	if colonyName, ok := fields["colony_name"].(string); ok {
		graph.ColonyName = colonyName
	}
	if state, ok := fields["state"].(int); ok {
		graph.State = state
	} else if state, ok := fields["state"].(float64); ok {
		graph.State = int(state)
	}
	
	// Deserialize complex fields
	if rootsStr, ok := fields["roots"].(string); ok {
		var roots []string
		if err := json.Unmarshal([]byte(rootsStr), &roots); err == nil {
			graph.Roots = roots
		}
	}
	if processIDsStr, ok := fields["process_ids"].(string); ok {
		var processIDs []string
		if err := json.Unmarshal([]byte(processIDsStr), &processIDs); err == nil {
			graph.ProcessIDs = processIDs
		}
	}
	
	return graph, nil
}