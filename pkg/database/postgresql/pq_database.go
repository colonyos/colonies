package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

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
	tz := os.Getenv("TZ")
	if tz == "" {
		return errors.New("Timezon environmental variable missing, try e.g. export TZ=Europe/Stockholm")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s", db.dbHost, db.dbPort, db.dbUser, db.dbPassword, db.dbName, tz)

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

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `PROCESSGRAPHS`
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

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `RUNTIMES (RUNTIME_ID TEXT PRIMARY KEY NOT NULL, RUNTIME_TYPE TEXT NOT NULL, NAME TEXT NOT NULL, COLONY_ID TEXT NOT NULL, CPU TEXT, CORES INTEGER, MEM INTEGER, GPU TEXT NOT NULL, GPUS INTEGER, STATE INTEGER, COMMISSIONTIME TIMESTAMPTZ, LASTHEARDFROM TIMESTAMPTZ)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `PROCESSES (PROCESS_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, TARGET_RUNTIME_IDS TEXT[], ASSIGNED_RUNTIME_ID TEXT, STATE INTEGER, IS_ASSIGNED BOOLEAN, RUNTIME_TYPE TEXT, SUBMISSION_TIME TIMESTAMPTZ, START_TIME TIMESTAMPTZ, END_TIME TIMESTAMPTZ, DEADLINE TIMESTAMPTZ, NAME TEXT, IMAGE TEXT, CMD TEXT, ARGS TEXT[], VOLUMES TEXT[], PORTS TEXT[], MAX_EXEC_TIME INTEGER, RETRIES INTEGER, MAX_RETRIES INTEGER, MEM INTEGER, CORES INTEGER, GPUS INTEGER, DEPENDENCIES TEXT[], PRIORITY INTEGER, WAIT_FOR_PARENTS BOOLEAN, PARENTS TEXT[], CHILDREN TEXT[], PROCESSGRAPH_ID TEXT)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID TEXT PRIMARY KEY NOT NULL, KEY TEXT NOT NULL, VALUE TEXT NOT NULL, ATTRIBUTE_TYPE INTEGER, TARGET_ID TEXT NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, PROCESSGRAPH_ID TEXT NOT NULL)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `PROCESSGRAPHS (PROCESSGRAPH_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, ROOTS TEXT[], STATE INTEGER, SUBMISSION_TIME TIMESTAMPTZ, START_TIME TIMESTAMPTZ, END_TIME TIMESTAMPTZ)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX1_` + db.dbPrefix + ` ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, SUBMISSION_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX2_` + db.dbPrefix + ` ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, START_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX3_` + db.dbPrefix + ` ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, END_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX PROCESSES_INDEX4_` + db.dbPrefix + ` ON ` + db.dbPrefix + `PROCESSES (IS_ASSIGNED, START_TIME, ASSIGNED_RUNTIME_ID, STATE, PROCESS_ID)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
