package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddCron(cron *core.Cron) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `CRONS (CRON_ID, COLONY_ID, NAME, CRON_EXPR, INTERVALL, RANDOM, NEXT_RUN, LAST_RUN, WORKFLOW_SPEC, LAST_PROCESSGRAPH_ID, SUCCESSFUL_RUNS, FAILED_RUNS) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := db.postgresql.Exec(sqlStatement, cron.ID, cron.ColonyID, cron.Name, cron.CronExpression, cron.Intervall, cron.Random, cron.NextRun, cron.LastRun, cron.WorkflowSpec, cron.LastProcessGraphID, cron.SuccessfulRuns, cron.FailedRuns)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string, successfulRuns int, failedRuns int) error {
	sqlStatement := `UPDATE  ` + db.dbPrefix + `CRONS SET NEXT_RUN=$1, LAST_RUN=$2, LAST_PROCESSGRAPH_ID=$3, SUCCESSFUL_RUNS=$4, FAILED_RUNS=$5 WHERE CRON_ID=$6`
	_, err := db.postgresql.Exec(sqlStatement, nextRun, lastRun, lastProcessGraphID, successfulRuns, failedRuns, cronID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseCrons(rows *sql.Rows) ([]*core.Cron, error) {
	var crons []*core.Cron

	for rows.Next() {
		var cronID string
		var colonyID string
		var name string
		var cronExpr string
		var intervall int
		var random bool
		var nextRun time.Time
		var lastRun time.Time
		var workflowSpec string
		var lastProcessGraphID string
		var successfulRuns int
		var failedRuns int

		if err := rows.Scan(&cronID, &colonyID, &name, &cronExpr, &intervall, &random, &nextRun, &lastRun, &workflowSpec, &lastProcessGraphID, &successfulRuns, &failedRuns); err != nil {
			return nil, err
		}

		cron := &core.Cron{ID: cronID, ColonyID: colonyID, Name: name, CronExpression: cronExpr, Intervall: intervall, Random: random, NextRun: nextRun, LastRun: lastRun, WorkflowSpec: workflowSpec, LastProcessGraphID: lastProcessGraphID, SuccessfulRuns: successfulRuns, FailedRuns: failedRuns}

		crons = append(crons, cron)
	}

	return crons, nil
}

func (db *PQDatabase) GetCronByID(cronID string) (*core.Cron, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `CRONS WHERE CRON_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, cronID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crons, err := db.parseCrons(rows)
	if err != nil {
		return nil, err
	}

	if len(crons) > 1 {
		return nil, errors.New("Expected one cron, cron id should be unique")
	}

	if len(crons) == 0 {
		return nil, nil
	}

	return crons[0], nil
}

func (db *PQDatabase) FindCronsByColonyID(colonyID string, count int) ([]*core.Cron, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `CRONS WHERE COLONY_ID=$1 LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	generators, err := db.parseCrons(rows)
	if err != nil {
		return nil, err
	}

	return generators, nil
}

func (db *PQDatabase) DeleteCronByID(cronID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `CRONS WHERE CRON_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, cronID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllCronsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `CRONS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
