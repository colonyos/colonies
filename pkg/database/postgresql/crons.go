package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddCron(cron *core.Cron) error {
	existingCron, err := db.GetCronByName(cron.ColonyName, cron.Name)
	if err != nil {
		return err
	}

	if existingCron != nil {
		return errors.New("Cron with name <" + cron.Name + "> in Colony <" + cron.ColonyName + "> already exists")
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `CRONS (CRON_ID, COLONY_NAME, NAME, CRON_EXPR, INTERVAL, RANDOM, NEXT_RUN, LAST_RUN, WORKFLOW_SPEC, PREV_PROCESSGRAPH_ID, WAIT_FOR_PREV_PROCESSGRAPH, INITIATOR_ID, INITIATOR_NAME) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = db.postgresql.Exec(sqlStatement, cron.ID, cron.ColonyName, cron.Name, cron.CronExpression, cron.Interval, cron.Random, cron.NextRun, cron.LastRun, cron.WorkflowSpec, cron.PrevProcessGraphID, cron.WaitForPrevProcessGraph, cron.InitiatorID, cron.InitiatorName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error {
	sqlStatement := `UPDATE  ` + db.dbPrefix + `CRONS SET NEXT_RUN=$1, LAST_RUN=$2, PREV_PROCESSGRAPH_ID=$3 WHERE CRON_ID=$4`
	_, err := db.postgresql.Exec(sqlStatement, nextRun, lastRun, lastProcessGraphID, cronID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseCrons(rows *sql.Rows) ([]*core.Cron, error) {
	var crons []*core.Cron

	for rows.Next() {
		var cronID string
		var colonyName string
		var name string
		var cronExpr string
		var interval int
		var random bool
		var nextRun time.Time
		var lastRun time.Time
		var workflowSpec string
		var prevProcessGraphID string
		var waitForPrevProcessGraph bool
		var initiatorID string
		var initiatorName string

		if err := rows.Scan(&cronID, &colonyName, &name, &cronExpr, &interval, &random, &nextRun, &lastRun, &workflowSpec, &prevProcessGraphID, &waitForPrevProcessGraph, &initiatorID, &initiatorName); err != nil {
			return nil, err
		}

		cron := &core.Cron{ID: cronID, ColonyName: colonyName, Name: name, CronExpression: cronExpr, Interval: interval, Random: random, NextRun: nextRun, LastRun: lastRun, WorkflowSpec: workflowSpec, PrevProcessGraphID: prevProcessGraphID, WaitForPrevProcessGraph: waitForPrevProcessGraph}

		cron.InitiatorID = initiatorID
		cron.InitiatorName = initiatorName

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

	if len(crons) == 0 {
		return nil, nil
	}

	return crons[0], nil
}

func (db *PQDatabase) GetCronByName(colonyName string, cronName string) (*core.Cron, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `CRONS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, cronName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crons, err := db.parseCrons(rows)
	if err != nil {
		return nil, err
	}

	if len(crons) == 0 {
		return nil, nil
	}

	return crons[0], nil
}

func (db *PQDatabase) FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `CRONS WHERE COLONY_NAME=$1 LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crons, err := db.parseCrons(rows)
	if err != nil {
		return nil, err
	}

	return crons, nil
}

func (db *PQDatabase) FindAllCrons() ([]*core.Cron, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `CRONS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crons, err := db.parseCrons(rows)
	if err != nil {
		return nil, err
	}

	return crons, nil

}

func (db *PQDatabase) RemoveCronByID(cronID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `CRONS WHERE CRON_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, cronID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllCronsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `CRONS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}
