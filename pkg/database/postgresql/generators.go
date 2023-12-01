package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddGenerator(generator *core.Generator) error {
	existingGenerator, err := db.GetGeneratorByName(generator.ColonyName, generator.Name)
	if err != nil {
		return err
	}

	if existingGenerator != nil {
		return errors.New("Generator with name <" + generator.Name + "> in Colony <" + generator.ColonyName + "> already exists")
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `GENERATORS (GENERATOR_ID, COLONY_NAME, NAME, WORKFLOW_SPEC, TRIGGER, TIMEOUT, LASTRUN, FIRSTPACK) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = db.postgresql.Exec(sqlStatement, generator.ID, generator.ColonyName, generator.Name, generator.WorkflowSpec, generator.Trigger, generator.Timeout, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseGenerators(rows *sql.Rows) ([]*core.Generator, error) {
	var generators []*core.Generator

	for rows.Next() {
		var generatorID string
		var colonyName string
		var name string
		var workflowSpec string
		var trigger int
		var timeout int
		var lastRun time.Time
		var firstPack time.Time
		if err := rows.Scan(&generatorID, &colonyName, &name, &workflowSpec, &trigger, &timeout, &lastRun, &firstPack); err != nil {
			return nil, err
		}

		generator := &core.Generator{ID: generatorID, ColonyName: colonyName, Name: name, WorkflowSpec: workflowSpec, Trigger: trigger, Timeout: timeout, LastRun: lastRun, FirstPack: firstPack}

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

	if len(generators) == 0 {
		return nil, nil
	}

	return generators[0], nil
}

func (db *PQDatabase) GetGeneratorByName(colonyName string, name string) (*core.Generator, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `GENERATORS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	generators, err := db.parseGenerators(rows)
	if err != nil {
		return nil, err
	}

	if len(generators) > 1 {
		return nil, errors.New("Expected one generator, generator name should be unique")
	}

	if len(generators) == 0 {
		return nil, nil
	}

	return generators[0], nil
}

func (db *PQDatabase) SetGeneratorLastRun(generatorID string) error {
	generator, err := db.GetGeneratorByID(generatorID)
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE  ` + db.dbPrefix + `GENERATORS SET LASTRUN=$1 WHERE GENERATOR_ID=$2`
	_, err = db.postgresql.Exec(sqlStatement, time.Now(), generator.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetGeneratorFirstPack(generatorID string) error {
	generator, err := db.GetGeneratorByID(generatorID)
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE  ` + db.dbPrefix + `GENERATORS SET FIRSTPACK=$1 WHERE GENERATOR_ID=$2`
	_, err = db.postgresql.Exec(sqlStatement, time.Now(), generator.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `GENERATORS WHERE COLONY_NAME=$1 LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, count)
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

func (db *PQDatabase) FindAllGenerators() ([]*core.Generator, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `GENERATORS`
	rows, err := db.postgresql.Query(sqlStatement)
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

func (db *PQDatabase) RemoveGeneratorByID(generatorID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `GENERATORS WHERE GENERATOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, generatorID)
	if err != nil {
		return err
	}

	return db.RemoveAllGeneratorArgsByGeneratorID(generatorID)
}

func (db *PQDatabase) RemoveAllGeneratorsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `GENERATORS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return db.RemoveAllGeneratorArgsByColonyName(colonyName)
}
