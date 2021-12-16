package postgresql

import (
	"colonies/pkg/core"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddWorker(worker *core.Worker) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `WORKERS (WORKER_ID, NAME, COLONY_ID, CPU, CORES, MEM, GPU, GPUS, STATUS) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.postgresql.Exec(sqlStatement, worker.ID(), worker.Name(), worker.ColonyID(), worker.CPU(), worker.Cores(), worker.Mem(), worker.GPU(), worker.GPUs(), 0)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseWorkers(rows *sql.Rows) ([]*core.Worker, error) {
	var workers []*core.Worker

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

		worker := core.CreateWorkerFromDB(id, name, colonyID, cpu, cores, mem, gpu, gpus, status)
		workers = append(workers, worker)
	}

	return workers, nil
}

func (db *PQDatabase) GetWorkers() ([]*core.Worker, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `WORKERS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseWorkers(rows)
}

func (db *PQDatabase) GetWorkerByID(workerID string) (*core.Worker, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `WORKERS WHERE WORKER_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, workerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workers, err := db.parseWorkers(rows)
	if err != nil {
		return nil, err
	}

	if len(workers) > 1 {
		return nil, errors.New("expected one worker, worker id should be unique")
	}

	if len(workers) == 0 {
		return nil, nil
	}

	return workers[0], nil
}

func (db *PQDatabase) GetWorkersByColonyID(workerID string) ([]*core.Worker, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `WORKERS WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, workerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workers, err := db.parseWorkers(rows)
	if err != nil {
		return nil, err
	}

	return workers, nil
}

func (db *PQDatabase) ApproveWorker(worker *core.Worker) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `WORKERS SET STATUS=1 WHERE WORKER_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, worker.ID())
	if err != nil {
		return err
	}

	worker.Approve()

	return nil
}

func (db *PQDatabase) RejectWorker(worker *core.Worker) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `WORKERS SET STATUS=2 WHERE WORKER_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, worker.ID())
	if err != nil {
		return err
	}

	worker.Reject()

	return nil
}

func (db *PQDatabase) DeleteWorkerByID(workerID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `WORKERS WHERE WORKER_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, workerID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteWorkersByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `WORKERS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
