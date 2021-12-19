package postgresql

import (
	"colonies/pkg/core"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddComputer(computer *core.Computer) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `COMPUTERS (COMPUTER_ID, NAME, COLONY_ID, CPU, CORES, MEM, GPU, GPUS, STATUS) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.postgresql.Exec(sqlStatement, computer.ID(), computer.Name(), computer.ColonyID(), computer.CPU(), computer.Cores(), computer.Mem(), computer.GPU(), computer.GPUs(), 0)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseComputers(rows *sql.Rows) ([]*core.Computer, error) {
	var computers []*core.Computer

	for rows.Next() {
		var id string
		var name string
		var colonyID string
		var cpu string
		var cores int
		var mem int
		var gpu string
		var gpus int
		var status int
		if err := rows.Scan(&id, &name, &colonyID, &cpu, &cores, &mem, &gpu, &gpus, &status); err != nil {
			return nil, err
		}

		computer := core.CreateComputerFromDB(id, name, colonyID, cpu, cores, mem, gpu, gpus, status)
		computers = append(computers, computer)
	}

	return computers, nil
}

func (db *PQDatabase) GetComputers() ([]*core.Computer, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `COMPUTERS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseComputers(rows)
}

func (db *PQDatabase) GetComputerByID(computerID string) (*core.Computer, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `COMPUTERS WHERE COMPUTER_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, computerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	computers, err := db.parseComputers(rows)
	if err != nil {
		return nil, err
	}

	if len(computers) > 1 {
		return nil, errors.New("Expected one computer, computer id should be unique")
	}

	if len(computers) == 0 {
		return nil, nil
	}

	return computers[0], nil
}

func (db *PQDatabase) GetComputersByColonyID(colonyID string) ([]*core.Computer, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `COMPUTERS WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	computers, err := db.parseComputers(rows)
	if err != nil {
		return nil, err
	}

	return computers, nil
}

func (db *PQDatabase) ApproveComputer(computer *core.Computer) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `COMPUTERS SET STATUS=1 WHERE COMPUTER_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, computer.ID())
	if err != nil {
		return err
	}

	computer.Approve()

	return nil
}

func (db *PQDatabase) RejectComputer(computer *core.Computer) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `COMPUTERS SET STATUS=2 WHERE COMPUTER_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, computer.ID())
	if err != nil {
		return err
	}

	computer.Reject()

	return nil
}

func (db *PQDatabase) DeleteComputerByID(computerID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `COMPUTERS WHERE COMPUTER_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, computerID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteComputersByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `COMPUTERS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
