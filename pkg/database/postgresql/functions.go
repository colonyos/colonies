package postgresql

import (
	"database/sql"
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddFunction(function *core.Function) error {
	argsJSON := ""
	if function.Args != nil {
		argsBytes, err := json.Marshal(function.Args)
		if err != nil {
			return err
		}
		argsJSON = string(argsBytes)
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `FUNCTIONS (FUNCTION_ID, EXECUTOR_NAME, EXECUTOR_TYPE, COLONY_NAME, FUNCNAME, DESCRIPTION, ARGS, COUNTER, MINWAITTIME, MAXWAITTIME, MINEXECTIME, MAXEXECTIME, AVGWAITTIME, AVGEXECTIME, LOCATION_NAME) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := db.postgresql.Exec(sqlStatement, function.FunctionID, function.ExecutorName, function.ExecutorType, function.ColonyName, function.FuncName, function.Description, argsJSON, function.Counter, function.MinWaitTime, function.MaxWaitTime, function.MinExecTime, function.MaxExecTime, function.AvgWaitTime, function.AvgExecTime, function.LocationName)
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
		var executorType string
		var colonyName string
		var name string
		var description sql.NullString
		var argsJSON sql.NullString
		var counter int
		var minWaitTime float64
		var maxWaitTime float64
		var minExecTime float64
		var maxExecTime float64
		var avgWaitTime float64
		var avgExecTime float64
		var locationName sql.NullString
		if err := rows.Scan(&functionID, &executorID, &executorType, &colonyName, &name, &description, &argsJSON, &counter, &minWaitTime, &maxWaitTime, &minExecTime, &maxExecTime, &avgWaitTime, &avgExecTime, &locationName); err != nil {
			return nil, err
		}

		function := core.CreateFunction(functionID, executorID, executorType, colonyName, name, counter, minWaitTime, maxWaitTime, minExecTime, maxExecTime, avgWaitTime, avgExecTime)
		if locationName.Valid {
			function.LocationName = locationName.String
		}

		if description.Valid {
			function.Description = description.String
		}

		if argsJSON.Valid && argsJSON.String != "" {
			var args []*core.FunctionArg
			if err := json.Unmarshal([]byte(argsJSON.String), &args); err != nil {
				return nil, err
			}
			function.Args = args
		}

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

	return nil, nil
}

func (db *PQDatabase) GetFunctionsByExecutorName(colonyName string, executorID string) ([]*core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, executorID)
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

func (db *PQDatabase) GetFunctionsByExecutorAndName(colonyName string, executorName string, name string) (*core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2 AND FUNCNAME=$3`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, executorName, name)
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

func (db *PQDatabase) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseFunctions(rows)
}

func (db *PQDatabase) UpdateFunctionStats(
	colonyName string,
	executorName string,
	name string,
	counter int,
	minWaitTime float64,
	maxWaitTime float64,
	minExecTime float64,
	maxExecTime float64,
	avgWaitTime float64,
	avgExecTime float64) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `FUNCTIONS SET COUNTER=$1, MINWAITTIME=$2, MAXWAITTIME=$3, MINEXECTIME=$4, MAXEXECTIME=$5, AVGWAITTIME=$6, AVGEXECTIME=$7 WHERE EXECUTOR_NAME=$8 AND FUNCNAME=$9 AND COLONY_NAME=$10`
	_, err := db.postgresql.Exec(sqlStatement, counter, minWaitTime, maxWaitTime, minExecTime, maxExecTime, avgWaitTime, avgExecTime, executorName, name, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveFunctionByID(functionID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE FUNCTION_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, functionID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveFunctionByName(colonyName string, executorName string, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2 AND FUNCNAME=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, executorName, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveFunctionsByExecutorName(colonyName string, executorName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, executorName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveFunctionsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveFunctions() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FUNCTIONS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
