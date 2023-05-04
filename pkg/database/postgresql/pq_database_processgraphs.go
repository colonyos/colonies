package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
)

func (db *PQDatabase) AddProcessGraph(processGraph *core.ProcessGraph) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `PROCESSGRAPHS (PROCESSGRAPH_ID, TARGET_COLONY_ID, ROOTS, STATE, SUBMISSION_TIME, START_TIME, END_TIME) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.postgresql.Exec(sqlStatement, processGraph.ID, processGraph.ColonyID, pq.Array(processGraph.Roots), processGraph.State, time.Now(), time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseProcessGraphs(rows *sql.Rows) ([]*core.ProcessGraph, error) {
	var graphs []*core.ProcessGraph

	for rows.Next() {
		var processGraphID string
		var colonyID string
		var roots []string
		var state int
		var submissionTime time.Time
		var startTime time.Time
		var endTime time.Time
		if err := rows.Scan(&processGraphID, &colonyID, pq.Array(&roots), &state, &submissionTime, &startTime, &endTime); err != nil {
			return nil, err
		}

		graph, err := core.CreateProcessGraph(colonyID)
		graph.ID = processGraphID
		graph.ColonyID = colonyID
		graph.State = state
		graph.SubmissionTime = submissionTime
		graph.StartTime = startTime
		graph.EndTime = endTime
		if err != nil {
			return graphs, err
		}

		for _, root := range roots {
			graph.AddRoot(root)
		}

		graphs = append(graphs, graph)
	}

	return graphs, nil
}

func (db *PQDatabase) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE PROCESSGRAPH_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, processGraphID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	processGraphs, err := db.parseProcessGraphs(rows)
	if err != nil {
		return nil, err
	}

	if len(processGraphs) > 1 {
		return nil, errors.New("Expected one processgraph, processgraph id should be unique")
	}

	if len(processGraphs) == 0 {
		return nil, nil
	}

	return processGraphs[0], nil
}

func (db *PQDatabase) SetProcessGraphState(processGraphID string, state int) error {
	graph, err := db.GetProcessGraphByID(processGraphID)
	if err != nil {
		return err
	}

	if graph.State == core.WAITING && state == core.RUNNING {
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSGRAPHS SET START_TIME=$1, STATE=$2 WHERE PROCESSGRAPH_ID=$3`
		_, err := db.postgresql.Exec(sqlStatement, time.Now(), state, processGraphID)
		if err != nil {
			return err
		}
	} else if state == core.SUCCESS || state == core.FAILED {
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSGRAPHS SET END_TIME=$1, STATE=$2 WHERE PROCESSGRAPH_ID=$3`
		_, err := db.postgresql.Exec(sqlStatement, time.Now(), state, processGraphID)
		if err != nil {
			return err
		}
	} else {
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSGRAPHS SET STATE=$1 WHERE PROCESSGRAPH_ID=$2`
		_, err := db.postgresql.Exec(sqlStatement, state, processGraphID)
		if err != nil {
			return err
		}

	}

	return nil
}

func (db *PQDatabase) FindWaitingProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.WAITING, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindRunningProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.RUNNING, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindSuccessfulProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.SUCCESS, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindFailedProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.FAILED, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) countProcessGraphsByColonyID(state int, colonyID string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE STATE=$1 AND TARGET_COLONY_ID=$2`
	rows, err := db.postgresql.Query(sqlStatement, state, colonyID)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	rows.Next()
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}
func (db *PQDatabase) CountWaitingProcessGraphsByColonyID(colonyID string) (int, error) {
	return db.countProcessGraphsByColonyID(core.WAITING, colonyID)
}

func (db *PQDatabase) CountRunningProcessGraphsByColonyID(colonyID string) (int, error) {
	return db.countProcessGraphsByColonyID(core.RUNNING, colonyID)
}

func (db *PQDatabase) CountSuccessfulProcessGraphsByColonyID(colonyID string) (int, error) {
	return db.countProcessGraphsByColonyID(core.SUCCESS, colonyID)
}

func (db *PQDatabase) CountFailedProcessGraphsByColonyID(colonyID string) (int, error) {
	return db.countProcessGraphsByColonyID(core.FAILED, colonyID)
}

func (db *PQDatabase) countProcessGraphs(state int) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE STATE=$1`
	rows, err := db.postgresql.Query(sqlStatement, state)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	rows.Next()
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (db *PQDatabase) DeleteAllProcessGraphsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return db.DeleteAllProcessesInProcessGraphsByColonyID(colonyID)
}

// XXX: This function may delete all belonging processes if the graph is running.
func (db *PQDatabase) DeleteAllWaitingProcessGraphsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, core.WAITING)
	if err != nil {
		return err
	}

	err = db.DeleteAllProcessesInProcessGraphsByColonyIDWithState(colonyID, core.WAITING)
	if err != nil {
		return err
	}

	return nil
}

// XXX: This function can cause inconsisteny, for example if the processgraph is running, and all running processes
// is deleted it will no longer be possible to resolve the processgraph
func (db *PQDatabase) DeleteAllRunningProcessGraphsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, core.RUNNING)
	if err != nil {
		return err
	}

	err = db.DeleteAllProcessesInProcessGraphsByColonyIDWithState(colonyID, core.RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllSuccessfulProcessGraphsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, core.SUCCESS)
	if err != nil {
		return err
	}

	err = db.DeleteAllProcessesInProcessGraphsByColonyIDWithState(colonyID, core.SUCCESS)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllFailedProcessGraphsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, core.FAILED)
	if err != nil {
		return err
	}

	err = db.DeleteAllProcessesInProcessGraphsByColonyIDWithState(colonyID, core.FAILED)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteProcessGraphByID(processGraphID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE PROCESSGRAPH_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, processGraphID)
	if err != nil {
		return err
	}

	return db.DeleteAllProcessesByProcessGraphID(processGraphID)
}

func (db *PQDatabase) CountWaitingProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.WAITING)
}

func (db *PQDatabase) CountRunningProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.RUNNING)
}

func (db *PQDatabase) CountSuccessfulProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.SUCCESS)
}

func (db *PQDatabase) CountFailedProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.FAILED)
}
