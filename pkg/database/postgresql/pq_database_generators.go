package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddGenerator(generator *core.Generator) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `GENERATORS (GENERATOR_ID, COLONY_ID, NAME, WORKFLOW_SPEC, TRIGGER, COUNTER, TIMEOUT, LASTRUN) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.postgresql.Exec(sqlStatement, generator.ID, generator.ColonyID, generator.Name, generator.WorkflowSpec, generator.Trigger, generator.Counter, generator.Timeout, time.Time{})
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseGenerators(rows *sql.Rows) ([]*core.Generator, error) {
	var generators []*core.Generator

	for rows.Next() {
		var generatorID string
		var colonyID string
		var name string
		var workflowSpec string
		var trigger int
		var counter int
		var timeout int
		var lastRun time.Time
		if err := rows.Scan(&generatorID, &colonyID, &name, &workflowSpec, &trigger, &counter, &timeout, &lastRun); err != nil {
			return nil, err
		}

		generator := &core.Generator{ID: generatorID, ColonyID: colonyID, Name: name, WorkflowSpec: workflowSpec, Trigger: trigger, Counter: counter, Timeout: timeout, LastRun: lastRun}

		generators = append(generators, generator)
	}

	return generators, nil
}

func (db *PQDatabase) GetGeneratorByID(generatorID string) (*core.Generator, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `GENERATORS WHERE GENERATOR_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, generatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	generators, err := db.parseGenerators(rows)
	if err != nil {
		return nil, err
	}

	if len(generators) > 1 {
		return nil, errors.New("Expected one generator, generator id should be unique")
	}

	if len(generators) == 0 {
		return nil, nil
	}

	return generators[0], nil
}

func (db *PQDatabase) FindGeneratorsByColonyID(colonyID string, count int) ([]*core.Generator, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `GENERATORS WHERE COLONY_ID=$1 LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	generators, err := db.parseGenerators(rows)
	if err != nil {
		return nil, err
	}

	return generators, nil
}

func (db *PQDatabase) DeleteGeneratorByID(generatorID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `GENERATORS WHERE GENERATOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, generatorID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllGeneratorsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `GENERATORS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) IncreaseGeneratorCounter(generatorID string) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `GENERATORS SET COUNTER=COUNTER+1 WHERE GENERATOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, generatorID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) ResetGenerator(generatorID string) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `GENERATORS SET COUNTER=0, LASTRUN=$1 WHERE GENERATOR_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, time.Now(), generatorID)
	if err != nil {
		return err
	}

	return nil
}
