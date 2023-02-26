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

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `EXECUTORS`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `FUNCTIONS`
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

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `GENERATORS`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `GENERATORARGS`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `CRONS`
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

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `EXECUTORS (EXECUTOR_ID TEXT PRIMARY KEY NOT NULL, EXECUTOR_TYPE TEXT NOT NULL, NAME TEXT NOT NULL UNIQUE, COLONY_ID TEXT NOT NULL, STATE INTEGER, REQUIRE_FUNC_REG BOOLEAN, COMMISSIONTIME TIMESTAMPTZ, LASTHEARDFROM TIMESTAMPTZ, LONG DOUBLE PRECISION, LAT DOUBLE PRECISION)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `FUNCTIONS (FUNCTION_ID TEXT PRIMARY KEY NOT NULL, EXECUTOR_ID TEXT NOT NULL, COLONY_ID TEXT NOT NULL, NAME TEXT NOT NULL, DESCRIPTION TEXT, COUNTER INTEGER, MINWAITTIME FLOAT, MAXWAITTIME FLOAT, MINEXECTIME FLOAT, MAXEXECTIME FLOAT, AVGWAITTIME FLOAT, AVGEXECTIME FLOAT, ARGS TEXT[])`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `PROCESSES (PROCESS_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, TARGET_EXECUTOR_IDS TEXT[], ASSIGNED_EXECUTOR_ID TEXT, STATE INTEGER, IS_ASSIGNED BOOLEAN, EXECUTOR_TYPE TEXT, SUBMISSION_TIME TIMESTAMPTZ, START_TIME TIMESTAMPTZ, END_TIME TIMESTAMPTZ, WAIT_DEADLINE TIMESTAMPTZ, EXEC_DEADLINE TIMESTAMPTZ, ERRORS TEXT[], NAME TEXT, FUNC TEXT, ARGS TEXT[], MAX_WAIT_TIME INTEGER, MAX_EXEC_TIME INTEGER, RETRIES INTEGER, MAX_RETRIES INTEGER, DEPENDENCIES TEXT[], PRIORITY INTEGER, WAIT_FOR_PARENTS BOOLEAN, PARENTS TEXT[], CHILDREN TEXT[], PROCESSGRAPH_ID TEXT, INPUT TEXT[], OUTPUT TEXT[])`
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

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `GENERATORS (GENERATOR_ID TEXT PRIMARY KEY NOT NULL, COLONY_ID TEXT NOT NULL, NAME TEXT NOT NULL UNIQUE, WORKFLOW_SPEC TEXT NOT NULL, TRIGGER INTEGER, LASTRUN TIMESTAMPTZ)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `GENERATORARGS (GENERATORARG_ID TEXT PRIMARY KEY NOT NULL, GENERATOR_ID TEXT NOT NULL, COLONY_ID TEXT NOT NULL, ARG TEXT NOT NULL)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `CRONS (CRON_ID TEXT PRIMARY KEY NOT NULL, COLONY_ID TEXT NOT NULL, NAME TEXT NOT NULL UNIQUE, CRON_EXPR TEXT NOT NULL, INTERVAL INT, RANDOM BOOLEAN, NEXT_RUN TIMESTAMPTZ, LAST_RUN TIMESTAMPTZ, WORKFLOW_SPEC TEXT NOT NULL, PREV_PROCESSGRAPH_ID TEXT NOT NULL, WAIT_FOR_PREV_PROCESSGRAPH BOOLEAN)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX1 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, SUBMISSION_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX2 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, START_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX3 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, END_TIME)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX4 ON ` + db.dbPrefix + `PROCESSES (IS_ASSIGNED, START_TIME, ASSIGNED_EXECUTOR_ID, STATE, PROCESS_ID)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX5 ON ` + db.dbPrefix + `PROCESSES (IS_ASSIGNED, START_TIME, ASSIGNED_EXECUTOR_ID, STATE, PROCESS_ID)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
