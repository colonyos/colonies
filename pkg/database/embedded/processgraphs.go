package embedded

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/embedded/index"
)

// copyProcessGraph creates a deep copy of a ProcessGraph.
func copyProcessGraph(g *core.ProcessGraph) *core.ProcessGraph {
	cp := *g
	if g.Roots != nil {
		cp.Roots = make([]string, len(g.Roots))
		copy(cp.Roots, g.Roots)
	}
	if g.ProcessIDs != nil {
		cp.ProcessIDs = make([]string, len(g.ProcessIDs))
		copy(cp.ProcessIDs, g.ProcessIDs)
	}
	if g.Nodes != nil {
		cp.Nodes = make([]core.GraphNode, len(g.Nodes))
		copy(cp.Nodes, g.Nodes)
	}
	if g.Edges != nil {
		cp.Edges = make([]core.Edge, len(g.Edges))
		copy(cp.Edges, g.Edges)
	}
	return &cp
}

func (db *EmbeddedDatabase) AddProcessGraph(processGraph *core.ProcessGraph) error {
	processGraph.SubmissionTime = time.Now()
	processGraph.State = core.WAITING

	cp := copyProcessGraph(processGraph)
	if err := db.processGraphs.Put(cp.ID, cp); err != nil {
		return err
	}

	db.processGraphsIdx.byColony.Add(cp.ColonyName, cp.State, index.IndexEntry[string]{
		SortKey:    cp.SubmissionTime.UnixNano(),
		PrimaryKey: cp.ID,
	})

	return nil
}

func (db *EmbeddedDatabase) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	g, ok := db.processGraphs.Get(processGraphID)
	if !ok {
		return nil, nil
	}
	return copyProcessGraph(g), nil
}

func (db *EmbeddedDatabase) SetProcessGraphState(processGraphID string, state int) error {
	g, ok := db.processGraphs.Get(processGraphID)
	if !ok {
		return nil
	}

	oldState := g.State

	cp := *g
	// Update timestamps based on transition
	if cp.State == core.WAITING && state == core.RUNNING {
		cp.StartTime = time.Now()
	} else if state == core.SUCCESS || state == core.FAILED || state == core.CANCELLED {
		cp.EndTime = time.Now()
	}

	cp.State = state

	if err := db.processGraphs.Put(processGraphID, &cp); err != nil {
		return err
	}

	// Update CompoundIndex
	db.processGraphsIdx.byColony.Remove(cp.ColonyName, oldState, index.IndexEntry[string]{
		SortKey:    cp.SubmissionTime.UnixNano(),
		PrimaryKey: cp.ID,
	})
	db.processGraphsIdx.byColony.Add(cp.ColonyName, state, index.IndexEntry[string]{
		SortKey:    cp.SubmissionTime.UnixNano(),
		PrimaryKey: cp.ID,
	})

	return nil
}

func (db *EmbeddedDatabase) findProcessGraphsByState(colonyName string, state int, count int) ([]*core.ProcessGraph, error) {
	var result []*core.ProcessGraph

	// Descend for most recent first
	db.processGraphsIdx.byColony.DescendFirst(colonyName, state, count, func(entry index.IndexEntry[string]) bool {
		if g, ok := db.processGraphs.Get(entry.PrimaryKey); ok {
			result = append(result, copyProcessGraph(g))
		}
		return true
	})

	return result, nil
}

func (db *EmbeddedDatabase) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.findProcessGraphsByState(colonyName, core.WAITING, count)
}

func (db *EmbeddedDatabase) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.findProcessGraphsByState(colonyName, core.RUNNING, count)
}

func (db *EmbeddedDatabase) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.findProcessGraphsByState(colonyName, core.SUCCESS, count)
}

func (db *EmbeddedDatabase) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.findProcessGraphsByState(colonyName, core.FAILED, count)
}

func (db *EmbeddedDatabase) FindCancelledProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.findProcessGraphsByState(colonyName, core.CANCELLED, count)
}

func (db *EmbeddedDatabase) RemoveProcessGraphByID(processGraphID string) error {
	g, ok := db.processGraphs.Get(processGraphID)
	if !ok {
		return nil
	}

	db.processGraphsIdx.byColony.Remove(g.ColonyName, g.State, index.IndexEntry[string]{
		SortKey:    g.SubmissionTime.UnixNano(),
		PrimaryKey: g.ID,
	})
	db.processGraphs.Delete(processGraphID)

	return db.RemoveAllProcessesByProcessGraphID(processGraphID)
}

func (db *EmbeddedDatabase) RemoveAllProcessGraphsByColonyName(colonyName string) error {
	graphs := db.processGraphs.Filter(func(g *core.ProcessGraph) bool {
		return g.ColonyName == colonyName
	})
	for _, g := range graphs {
		db.processGraphsIdx.byColony.Remove(g.ColonyName, g.State, index.IndexEntry[string]{
			SortKey:    g.SubmissionTime.UnixNano(),
			PrimaryKey: g.ID,
		})
		db.processGraphs.Delete(g.ID)
	}
	return db.RemoveAllProcessesInProcessGraphsByColonyName(colonyName)
}

func (db *EmbeddedDatabase) removeProcessGraphsByColonyNameAndState(colonyName string, state int) error {
	graphs := db.processGraphs.Filter(func(g *core.ProcessGraph) bool {
		return g.ColonyName == colonyName && g.State == state
	})
	for _, g := range graphs {
		db.processGraphsIdx.byColony.Remove(g.ColonyName, g.State, index.IndexEntry[string]{
			SortKey:    g.SubmissionTime.UnixNano(),
			PrimaryKey: g.ID,
		})
		db.processGraphs.Delete(g.ID)
	}

	// Remove processes in process graphs with matching state
	db.removeProcessesInProcessGraphsByColonyNameWithState(colonyName, state)

	return nil
}

func (db *EmbeddedDatabase) RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error {
	return db.removeProcessGraphsByColonyNameAndState(colonyName, core.WAITING)
}

func (db *EmbeddedDatabase) RemoveAllRunningProcessGraphsByColonyName(colonyName string) error {
	return db.removeProcessGraphsByColonyNameAndState(colonyName, core.RUNNING)
}

func (db *EmbeddedDatabase) RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error {
	return db.removeProcessGraphsByColonyNameAndState(colonyName, core.SUCCESS)
}

func (db *EmbeddedDatabase) RemoveAllFailedProcessGraphsByColonyName(colonyName string) error {
	return db.removeProcessGraphsByColonyNameAndState(colonyName, core.FAILED)
}

func (db *EmbeddedDatabase) RemoveAllCancelledProcessGraphsByColonyName(colonyName string) error {
	return db.removeProcessGraphsByColonyNameAndState(colonyName, core.CANCELLED)
}

func (db *EmbeddedDatabase) CountWaitingProcessGraphs() (int, error) {
	return db.processGraphsIdx.byColony.CountBySubgroup(core.WAITING), nil
}

func (db *EmbeddedDatabase) CountRunningProcessGraphs() (int, error) {
	return db.processGraphsIdx.byColony.CountBySubgroup(core.RUNNING), nil
}

func (db *EmbeddedDatabase) CountSuccessfulProcessGraphs() (int, error) {
	return db.processGraphsIdx.byColony.CountBySubgroup(core.SUCCESS), nil
}

func (db *EmbeddedDatabase) CountFailedProcessGraphs() (int, error) {
	return db.processGraphsIdx.byColony.CountBySubgroup(core.FAILED), nil
}

func (db *EmbeddedDatabase) CountCancelledProcessGraphs() (int, error) {
	return db.processGraphsIdx.byColony.CountBySubgroup(core.CANCELLED), nil
}

func (db *EmbeddedDatabase) CountWaitingProcessGraphsByColonyName(colonyName string) (int, error) {
	return db.processGraphsIdx.byColony.Count(colonyName, core.WAITING), nil
}

func (db *EmbeddedDatabase) CountRunningProcessGraphsByColonyName(colonyName string) (int, error) {
	return db.processGraphsIdx.byColony.Count(colonyName, core.RUNNING), nil
}

func (db *EmbeddedDatabase) CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error) {
	return db.processGraphsIdx.byColony.Count(colonyName, core.SUCCESS), nil
}

func (db *EmbeddedDatabase) CountFailedProcessGraphsByColonyName(colonyName string) (int, error) {
	return db.processGraphsIdx.byColony.Count(colonyName, core.FAILED), nil
}

func (db *EmbeddedDatabase) CountCancelledProcessGraphsByColonyName(colonyName string) (int, error) {
	return db.processGraphsIdx.byColony.Count(colonyName, core.CANCELLED), nil
}

// removeProcessesInProcessGraphsByColonyNameWithState removes processes in process graphs
// with matching colony name and state. Internal helper for cascade deletes.
func (db *EmbeddedDatabase) removeProcessesInProcessGraphsByColonyNameWithState(colonyName string, state int) {
	db.RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName, state)
	processes := db.processes.Filter(func(p *core.Process) bool {
		return p.FunctionSpec.Conditions.ColonyName == colonyName &&
			p.ProcessGraphID != "" &&
			p.State == state
	})
	for _, p := range processes {
		db.removeProcessFromIndexes(p)
		db.processes.Delete(p.ID)
	}
}
