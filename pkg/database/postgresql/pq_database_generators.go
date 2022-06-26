package postgresql

import (
	"database/sql"
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddGenerator(generator *core.Generator) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `GENERATORS (GENERATOR_ID, COLONY_ID, NAME, WORKFLOW_SPEC, TRIGGER, COUNTER, TIMEOUT) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.postgresql.Exec(sqlStatement, generator.ID, generator.ColonyID, generator.Name, generator.WorkflowSpec, generator.Trigger, generator.Counter, generator.Timeout)
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
		if err := rows.Scan(&generatorID, &colonyID, &name, &workflowSpec, &trigger, &counter, &timeout); err != nil {
			return nil, err
		}

		generator := &core.Generator{ID: generatorID, ColonyID: colonyID, Name: name, WorkflowSpec: workflowSpec, Trigger: trigger, Counter: counter, Timeout: timeout}

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
	return nil
}

func (db *PQDatabase) DeleteAllGeneratorsByColonyID(colonyID string) error {
	return nil
}

func (db *PQDatabase) IncreaseCounter(generatorID string) error {
	return nil
}

func (db *PQDatabase) ResetCounter(generatorID string) error {
	return nil
}
