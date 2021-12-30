package postgresql

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PQDatabase struct {
	postgresql *sql.DB
	dbHost     string
	dbPort     int
	dbUser     string
	dbPassword string
	dbName     string
	dbPrefix   string
}

func CreatePQDatabase(dbHost string, dbPort int, dbUser string, dbPassword string, dbName string, dbPrefix string) *PQDatabase {
	return &PQDatabase{dbHost: dbHost, dbPort: dbPort, dbUser: dbUser, dbPassword: dbPassword, dbName: dbName, dbPrefix: dbPrefix}
}

func (db *PQDatabase) Connect() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db.dbHost, db.dbPort, db.dbUser, db.dbPassword, db.dbName)

	postgresql, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	db.postgresql = postgresql

	err = db.postgresql.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) Close() {
	db.postgresql.Close()
}

func (db *PQDatabase) Drop() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `COLONIES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `RUNTIMES`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `PROCESSES`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `ATTRIBUTES`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP INDEX PROCESSES_INDEX1`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP INDEX PROCESSES_INDEX2`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP INDEX PROCESSES_INDEX3`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP INDEX PROCESSES_INDEX4`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) Initialize() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `COLONIES (COLONY_ID TEXT PRIMARY KEY NOT NULL, NAME TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `RUNTIMES (RUNTIME_ID TEXT PRIMARY KEY NOT NULL, RUNTIME_TYPE TEXT NOT NULL, NAME TEXT NOT NULL, COLONY_ID TEXT NOT NULL, CPU TEXT, CORES INTEGER, MEM INTEGER, GPU TEXT NOT NULL, GPUS INTEGER, STATUS INTEGER)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `PROCESSES (PROCESS_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, TARGET_runtime_IDS TEXT[], ASSIGNED_RUNTIME_ID TEXT, STATUS INTEGER, IS_ASSIGNED BOOLEAN, runtime_TYPE TEXT, SUBMISSION_TIME TIMESTAMP, START_TIME TIMESTAMP, END_TIME TIMESTAMP, DEADLINE TIMESTAMP, TIMEOUT INTEGER, RETRIES INTEGER, MAX_RETRIES INTEGER, MEM INTEGER, CORES INTEGER, GPUS INTEGER)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID TEXT PRIMARY KEY NOT NULL, KEY TEXT NOT NULL, VALUE TEXT NOT NULL, ATTRIBUTE_TYPE INTEGER, TARGET_ID TEXT NOT NULL)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX1 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATUS, SUBMISSION_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX2 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATUS, START_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX3 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATUS, END_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX4 ON ` + db.dbPrefix + `PROCESSES (IS_ASSIGNED, START_TIME, ASSIGNED_RUNTIME_ID, STATUS, PROCESS_ID)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
