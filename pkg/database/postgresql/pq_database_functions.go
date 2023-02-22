package postgresql

import (
	"database/sql"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
)

func (db *PQDatabase) AddFunction(function core.Function) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `FUNCTIONS (FUNCTION_ID, EXECUTOR_ID, COLONY_ID, NAME, DESCRIPTION, AVGWAITTIME, AVGEXECTIME, ARGS) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.postgresql.Exec(sqlStatement, function.FunctionID, function.ExecutorID, function.ColonyID, function.Name, function.Desc, function.AvgWaitTime, function.AvgExecTime, pq.Array(function.Args))
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseFunctions(rows *sql.Rows) ([]core.Function, error) {
	var functions []core.Function

	for rows.Next() {
		var functionID string
		var executorID string
		var colonyID string
		var name string
		var desc string
		var avgWaitTime float64
		var avgExecTime float64
		var args []string
		if err := rows.Scan(&functionID, &executorID, &colonyID, &name, &desc, &avgWaitTime, &avgExecTime, pq.Array(&args)); err != nil {
			return nil, err
		}

		function := core.CreateFunction(functionID, executorID, colonyID, name, desc, avgWaitTime, avgExecTime, args)
		functions = append(functions, function)
	}

	return functions, nil
}

func (db *PQDatabase) GetFunctionsByExecutorID(executorID string) ([]core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE EXECUTOR_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, executorID)
	if err != nil {
		return []core.Function{}, err
	}

	defer rows.Close()

	functions, err := db.parseFunctions(rows)
	if err != nil {
		return []core.Function{}, err
	}

	return functions, nil
}

func (db *PQDatabase) GetFunctionsByColonyID(colonyID string) ([]core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return []core.Function{}, err
	}

	defer rows.Close()

	return db.parseFunctions(rows)
}

func (db *PQDatabase) UpdateFunctionTimes(executorID string, name string, avgWaitTime float64, avgExecTime float64) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `FUNCTIONS SET AVGWAITTIME=$1, AVGEXECTIME=$2 WHERE EXECUTOR_ID=$3 AND NAME=$4`
	_, err := db.postgresql.Exec(sqlStatement, avgWaitTime, avgExecTime, executorID, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteFunctionByName(executorID string, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE EXECUTOR_ID=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, executorID, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteFunctionsByExecutorID(executorID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE EXECUTOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, executorID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteFunctionsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteFunctions() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
