package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddProcessGraph(processGraph *core.ProcessGraph) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `PROCESSGRAPHS (PROCESSGRAPH_ID, ROOT, STATE, SUBMISSION_TIME, END_TIME, RUNTIME_GROUP) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.postgresql.Exec(sqlStatement, processGraph.ID, processGraph.Root, processGraph.State, time.Now(), processGraph.EndTime, processGraph.RuntimeGroup)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseProcessGraphs(processGraphStorage core.ProcessGraphStorage, rows *sql.Rows) ([]*core.ProcessGraph, error) {
	var graphs []*core.ProcessGraph

	for rows.Next() {
		var processGraphID string
		var root string
		var state int
		var submissionTime time.Time
		var endTime time.Time
		var runtimeGroup string
		if err := rows.Scan(&processGraphID, &root, &state, &submissionTime, &endTime, &runtimeGroup); err != nil {
			return nil, err
		}

		graph, err := core.CreateProcessGraph(processGraphStorage, root)
		graph.ID = processGraphID
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
