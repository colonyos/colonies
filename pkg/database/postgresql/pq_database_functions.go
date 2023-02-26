package postgresql

import (
	"database/sql"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
)

func (db *PQDatabase) AddFunction(function *core.Function) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `FUNCTIONS (FUNCTION_ID, EXECUTOR_ID, COLONY_ID, NAME, DESCRIPTION, COUNTER, MINWAITTIME, MAXWAITTIME, MINEXECTIME, MAXEXECTIME, AVGWAITTIME, AVGEXECTIME, ARGS) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := db.postgresql.Exec(sqlStatement, function.FunctionID, function.ExecutorID, function.ColonyID, function.Name, function.Desc, function.Counter, function.MinWaitTime, function.MaxWaitTime, function.MinExecTime, function.MaxExecTime, function.AvgWaitTime, function.AvgExecTime, pq.Array(function.Args))
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseFunctions(rows *sql.Rows) ([]*core.Function, error) {
	var functions []*core.Function

	for rows.Next() {
		var functionID string
		var executorID string
		var colonyID string
		var name string
		var desc string
		var counter int
		var minWaitTime float64
		var maxWaitTime float64
		var minExecTime float64
		var maxExecTime float64
		var avgWaitTime float64
		var avgExecTime float64
		var args []string
		if err := rows.Scan(&functionID, &executorID, &colonyID, &name, &desc, &counter, &minWaitTime, &maxWaitTime, &minExecTime, &maxExecTime, &avgWaitTime, &avgExecTime, pq.Array(&args)); err != nil {
			return nil, err
		}

		function := core.CreateFunction(functionID, executorID, colonyID, name, desc, counter, minWaitTime, maxWaitTime, minExecTime, maxExecTime, avgWaitTime, avgExecTime, args)
		functions = append(functions, function)
	}

	return functions, nil
}

func (db *PQDatabase) GetFunctionByID(functionID string) (*core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE FUNCTION_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, functionID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	functions, err := db.parseFunctions(rows)
	if err != nil {
		return nil, err
	}

	if len(functions) > 0 {
		return functions[0], nil
	}

	return functions[0], nil
}

func (db *PQDatabase) GetFunctionsByExecutorID(executorID string) ([]*core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE EXECUTOR_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, executorID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	functions, err := db.parseFunctions(rows)
	if err != nil {
		return nil, err
	}

	return functions, nil
}

func (db *PQDatabase) GetFunctionsByExecutorIDAndName(executorID string, name string) (*core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE EXECUTOR_ID=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, executorID, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	functions, err := db.parseFunctions(rows)
	if err != nil {
		return nil, err
	}

	if len(functions) > 0 {
		return functions[0], nil
	}

	return nil, nil
}

func (db *PQDatabase) GetFunctionsByColonyID(colonyID string) ([]*core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseFunctions(rows)
}

func (db *PQDatabase) UpdateFunctionStats(executorID string,
	name string,
	counter int,
	minWaitTime float64,
	maxWaitTime float64,
	minExecTime float64,
	maxExecTime float64,
	avgWaitTime float64,
	avgExecTime float64) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `FUNCTIONS SET COUNTER=$1, MINWAITTIME=$2, MAXWAITTIME=$3, MINEXECTIME=$4, MAXEXECTIME=$5, AVGWAITTIME=$6, AVGEXECTIME=$7 WHERE EXECUTOR_ID=$8 AND NAME=$9`
	_, err := db.postgresql.Exec(sqlStatement, counter, minWaitTime, maxWaitTime, minExecTime, maxExecTime, avgWaitTime, avgExecTime, executorID, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteFunctionByID(functionID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE FUNCTION_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, functionID)
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
