package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddProcessGraph(processGraph *core.ProcessGraph) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `PROCESSGRAPHS (PROCESSGRAPH_ID, TARGET_COLONY_ID, ROOT, STATE, SUBMISSION_TIME, END_TIME, RUNTIME_GROUP) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.postgresql.Exec(sqlStatement, processGraph.ID, processGraph.ColonyID, processGraph.Root, processGraph.State, time.Now(), processGraph.EndTime, processGraph.RuntimeGroup)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseProcessGraphs(processGraphStorage core.ProcessGraphStorage, rows *sql.Rows) ([]*core.ProcessGraph, error) {
	var graphs []*core.ProcessGraph

	for rows.Next() {
		var processGraphID string
		var colonyID string
		var root string
		var state int
		var submissionTime time.Time
		var endTime time.Time
		var runtimeGroup string
		if err := rows.Scan(&processGraphID, &colonyID, &root, &state, &submissionTime, &endTime, &runtimeGroup); err != nil {
			return nil, err
		}

		graph, err := core.CreateProcessGraph(processGraphStorage, colonyID, root)
		graph.ID = processGraphID
		graph.ColonyID = colonyID
		graph.State = state
		graph.SubmissionTime = submissionTime
		graph.EndTime = endTime
		graph.RuntimeGroup = runtimeGroup
		if err != nil {
			return graphs, err
		}
		graphs = append(graphs, graph)
	}

	return graphs, nil
}

func (db *PQDatabase) GetProcessGraphByID(processGraphStorage core.ProcessGraphStorage, processGraphID string) (*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE PROCESSGRAPH_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, processGraphID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	processGraphs, err := db.parseProcessGraphs(processGraphStorage, rows)
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
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSGRAPHS SET STATE=$1 WHERE PROCESSGRAPH_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, state, processGraphID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) FindWaitingProcessGraphs(processGraphStorage core.ProcessGraphStorage, colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.WAITING, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(processGraphStorage, rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindRunningProcessGraphs(processGraphStorage core.ProcessGraphStorage, colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.RUNNING, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(processGraphStorage, rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindSuccessfulProcessGraphs(processGraphStorage core.ProcessGraphStorage, colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.SUCCESS, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(processGraphStorage, rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindFailedProcessGraphs(processGraphStorage core.ProcessGraphStorage, colonyID string, count int) ([]*core.ProcessGraph, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.FAILED, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcessGraphs(processGraphStorage, rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) countProcessGraphsForColony(state int, colonyID string) (int, error) {
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

func (db *PQDatabase) NrOfWaitingProcessGraphsForColony(colonyID string) (int, error) {
	return db.countProcessGraphsForColony(core.WAITING, colonyID)
}

func (db *PQDatabase) NrOfRunningProcessGraphsForColony(colonyID string) (int, error) {
	return db.countProcessGraphsForColony(core.RUNNING, colonyID)
}

func (db *PQDatabase) NrOfSuccessfulProcessGraphsForColony(colonyID string) (int, error) {
	return db.countProcessGraphsForColony(core.SUCCESS, colonyID)
}

func (db *PQDatabase) NrOfFailedProcessGraphsForColony(colonyID string) (int, error) {
	return db.countProcessGraphsForColony(core.FAILED, colonyID)
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

func (db *PQDatabase) NrOfWaitingProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.WAITING)
}

func (db *PQDatabase) NrOfRunningProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.RUNNING)
}

func (db *PQDatabase) NrOfSuccessfulProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.SUCCESS)
}

func (db *PQDatabase) NrOfFailedProcessGraphs() (int, error) {
	return db.countProcessGraphs(core.FAILED)
}
