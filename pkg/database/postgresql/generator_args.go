package postgresql

import (
	"database/sql"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddGeneratorArg(generatorArg *core.GeneratorArg) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `GENERATORARGS (GENERATORARG_ID, GENERATOR_ID, COLONY_ID, ARG) VALUES ($1, $2, $3, $4)`
	_, err := db.postgresql.Exec(sqlStatement, generatorArg.ID, generatorArg.GeneratorID, generatorArg.ColonyID, generatorArg.Arg)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseGeneratorArgs(rows *sql.Rows) ([]*core.GeneratorArg, error) {
	var generatorArgs []*core.GeneratorArg

	for rows.Next() {
		var generatorArgID string
		var generatorID string
		var colonyID string
		var arg string
		if err := rows.Scan(&generatorArgID, &generatorID, &colonyID, &arg); err != nil {
			return nil, err
		}

		generatorArg := &core.GeneratorArg{ID: generatorArgID, GeneratorID: generatorID, ColonyID: colonyID, Arg: arg}

		generatorArgs = append(generatorArgs, generatorArg)
	}

	return generatorArgs, nil
}

func (db *PQDatabase) GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `GENERATORARGS WHERE GENERATOR_ID=$1 LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, generatorID, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	generatorArgs, err := db.parseGeneratorArgs(rows)
	if err != nil {
		return nil, err
	}

	return generatorArgs, nil
}

func (db *PQDatabase) CountGeneratorArgs(generatorID string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `GENERATORARGS WHERE GENERATOR_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, generatorID)
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

func (db *PQDatabase) DeleteGeneratorArgByID(generatorArgsID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `GENERATORARGS WHERE GENERATORARG_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, generatorArgsID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllGeneratorArgsByGeneratorID(generatorID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `GENERATORARGS WHERE GENERATOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, generatorID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllGeneratorArgsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `GENERATORARGS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
