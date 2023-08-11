package postgresql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type Postgresql interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	Driver() driver.Driver
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

type PQDatabase struct {
	postgresql  Postgresql
	dbHost      string
	dbPort      int
	dbUser      string
	dbPassword  string
	dbName      string
	dbPrefix    string
	timescaleDB bool
}

func CreatePQDatabase(dbHost string, dbPort int, dbUser string, dbPassword string, dbName string, dbPrefix string, timescaleDB bool) *PQDatabase {
	return &PQDatabase{dbHost: dbHost, dbPort: dbPort, dbUser: dbUser, dbPassword: dbPassword, dbName: dbName, dbPrefix: dbPrefix, timescaleDB: timescaleDB}
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

func (db *PQDatabase) dropColoniesTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `COLONIES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropExecutorsTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `EXECUTORS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropFunctionsTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `FUNCTIONS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropProcessesTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `PROCESSES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropLogTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `LOGS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropFileTable() error {
	sqlStatement := `DROP SEQUENCE ` + db.dbPrefix + `FILE_SEQ`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `FILES`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropSnapshotTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `SNAPSHOTS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropAttributesTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `ATTRIBUTES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropProcessGraphsTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `PROCESSGRAPHS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropGeneratorsTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `GENERATORS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropGeneratorArgsTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `GENERATORARGS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropCronsTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `CRONS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) Drop() error {
	err := db.dropColoniesTable()
	if err != nil {
		return err
	}

	err = db.dropExecutorsTable()
	if err != nil {
		return err
	}

	err = db.dropFunctionsTable()
	if err != nil {
		return err
	}

	err = db.dropProcessesTable()
	if err != nil {
		return err
	}

	err = db.dropLogTable()
	if err != nil {
		return err
	}

	err = db.dropFileTable()
	if err != nil {
		return err
	}

	err = db.dropSnapshotTable()
	if err != nil {
		return err
	}

	err = db.dropAttributesTable()
	if err != nil {
		return err
	}

	err = db.dropProcessGraphsTable()
	if err != nil {
		return err
	}

	err = db.dropGeneratorsTable()
	if err != nil {
		return err
	}

	err = db.dropGeneratorArgsTable()
	if err != nil {
		return err
	}

	err = db.dropCronsTable()
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createHypertables() error {
	sqlStatement := `SELECT create_hypertable ('` + db.dbPrefix + `LOGS', 'TS', chunk_time_interval => 86400000000000)` // 24h chunks, assuming ts is nanosec
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createColoniesTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `COLONIES (COLONY_ID TEXT PRIMARY KEY NOT NULL, NAME TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createExecutorsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `EXECUTORS (EXECUTOR_ID TEXT PRIMARY KEY NOT NULL, EXECUTOR_TYPE TEXT NOT NULL, NAME TEXT NOT NULL, COLONY_ID TEXT NOT NULL, STATE INTEGER, REQUIRE_FUNC_REG BOOLEAN, COMMISSIONTIME TIMESTAMPTZ, LASTHEARDFROM TIMESTAMPTZ, LONG DOUBLE PRECISION, LAT DOUBLE PRECISION, LOCDESC TEXT, HWMODEL TEXT, HWCPU TEXT, HWMEM TEXT, HWSTORAGE TEXT, HWGPUNAME TEXT, HWGPUCOUNT TEXT, SWNAME TEXT, SWTYPE TEXT, SWVERSION TEXT)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}
	return nil
}

func (db *PQDatabase) createFunctionsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `FUNCTIONS (FUNCTION_ID TEXT PRIMARY KEY NOT NULL, EXECUTOR_ID TEXT NOT NULL, COLONY_ID TEXT NOT NULL, FUNCNAME TEXT NOT NULL, COUNTER INTEGER, MINWAITTIME FLOAT, MAXWAITTIME FLOAT, MINEXECTIME FLOAT, MAXEXECTIME FLOAT, AVGWAITTIME FLOAT, AVGEXECTIME FLOAT)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `PROCESSES (PROCESS_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, TARGET_EXECUTOR_IDS TEXT[], ASSIGNED_EXECUTOR_ID TEXT, STATE INTEGER, IS_ASSIGNED BOOLEAN, EXECUTOR_TYPE TEXT, SUBMISSION_TIME TIMESTAMPTZ, START_TIME TIMESTAMPTZ, END_TIME TIMESTAMPTZ, WAIT_DEADLINE TIMESTAMPTZ, EXEC_DEADLINE TIMESTAMPTZ, ERRORS TEXT[], NODENAME TEXT, FUNCNAME TEXT, ARGS TEXT, KWARGS TEXT, MAX_WAIT_TIME INTEGER, MAX_EXEC_TIME INTEGER, RETRIES INTEGER, MAX_RETRIES INTEGER, DEPENDENCIES TEXT[], PRIORITY INTEGER, PRIORITYTIME BIGINT, WAIT_FOR_PARENTS BOOLEAN, PARENTS TEXT[], CHILDREN TEXT[], PROCESSGRAPH_ID TEXT, INPUT TEXT, OUTPUT TEXT, LABEL TEXT)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createLogTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `LOGS (PROCESS_ID TEXT, COLONY_ID TEXT NOT NULL, EXECUTOR_ID TEXT NOT NULL, TS BIGINT, MSG TEXT NOT NULL, ADDED TIMESTAMPTZ)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createFileTable() error {
	sqlStatement := `CREATE SEQUENCE ` + db.dbPrefix + `FILE_SEQ`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `FILES (FILE_ID TEXT PRIMARY KEY NOT NULL, COLONY_ID TEXT NOT NULL, LABEL TEXT NOT NULL, NAME TEXT NOT NULL, SIZE BIGINT, SEQNR BIGINT, CHECKSUM TEXT, CHECKSUM_ALG TEXT, ADDED TIMESTAMPTZ, PROTOCOL TEXT, S3_SERVER TEXT, S3_PORT INTEGER, S3_TLS BOOLEAN, S3_ACCESSKEY TEXT, S3_SECRETKEY TEXT, S3_REGION TEXT, S3_ENCKEY TEXT, S3_ENCALG TEXT, S3_OBJ TEXT, S3_BUCKET TEXT)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createSnapshotTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `SNAPSHOTS (SNAPSHOT_ID TEXT PRIMARY KEY NOT NULL, COLONY_ID TEXT NOT NULL, LABEL TEXT NOT NULL, NAME TEXT NOT NULL, FILE_IDS TEXT[], ADDED TIMESTAMPTZ)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createAttributesTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID TEXT PRIMARY KEY NOT NULL, KEY TEXT NOT NULL, VALUE TEXT NOT NULL, ATTRIBUTE_TYPE INTEGER, TARGET_ID TEXT NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, PROCESSGRAPH_ID TEXT NOT NULL, ADDED TIMESTAMPTZ, STATE INTEGER)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessGraphsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `PROCESSGRAPHS (PROCESSGRAPH_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, ROOTS TEXT[], STATE INTEGER, SUBMISSION_TIME TIMESTAMPTZ, START_TIME TIMESTAMPTZ, END_TIME TIMESTAMPTZ)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createGeneratorsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `GENERATORS (GENERATOR_ID TEXT PRIMARY KEY NOT NULL, COLONY_ID TEXT NOT NULL, NAME TEXT NOT NULL UNIQUE, WORKFLOW_SPEC TEXT NOT NULL, TRIGGER INTEGER, TIMEOUT INTEGER, LASTRUN TIMESTAMPTZ, FIRSTPACK TIMESTAMPTZ)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createGeneratorArgsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `GENERATORARGS (GENERATORARG_ID TEXT PRIMARY KEY NOT NULL, GENERATOR_ID TEXT NOT NULL, COLONY_ID TEXT NOT NULL, ARG TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createCronsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `CRONS (CRON_ID TEXT PRIMARY KEY NOT NULL, COLONY_ID TEXT NOT NULL, NAME TEXT NOT NULL UNIQUE, CRON_EXPR TEXT NOT NULL, INTERVAL INT, RANDOM BOOLEAN, NEXT_RUN TIMESTAMPTZ, LAST_RUN TIMESTAMPTZ, WORKFLOW_SPEC TEXT NOT NULL, PREV_PROCESSGRAPH_ID TEXT NOT NULL, WAIT_FOR_PREV_PROCESSGRAPH BOOLEAN)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex1() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX1 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, SUBMISSION_TIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex2() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX2 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, START_TIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex3() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX3 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, END_TIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex4() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX4 ON ` + db.dbPrefix + `PROCESSES (IS_ASSIGNED, START_TIME, ASSIGNED_EXECUTOR_ID, STATE, PROCESS_ID)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex5() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX5 ON ` + db.dbPrefix + `PROCESSES (IS_ASSIGNED, START_TIME, ASSIGNED_EXECUTOR_ID, STATE, PROCESS_ID)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex6() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX6 ON ` + db.dbPrefix + `PROCESSES (STATE, EXECUTOR_TYPE, IS_ASSIGNED, WAIT_FOR_PARENTS, TARGET_COLONY_ID, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex7() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX7 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex8() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX8 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_ID, STATE, EXECUTOR_TYPE, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createAttributesIndex1() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `ATTRIBUTES_INDEX1 ON ` + db.dbPrefix + `ATTRIBUTES (TARGET_ID, ATTRIBUTE_TYPE)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createAttributesIndex2() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `ATTRIBUTES_INDEX2 ON ` + db.dbPrefix + `ATTRIBUTES (TARGET_ID)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createRetentionIndex1() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `RETENTION_INDEX1 ON ` + db.dbPrefix + `ATTRIBUTES (ADDED, STATE)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createRetentionIndex2() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `RETENTION_INDEX2 ON ` + db.dbPrefix + `PROCESSES (SUBMISSION_TIME, STATE)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createRetentionIndex3() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `RETENTION_INDEX3 ON ` + db.dbPrefix + `PROCESSGRAPHS (SUBMISSION_TIME, STATE)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createRetentionIndex4() error {
	if !db.timescaleDB {
		sqlStatement := `CREATE INDEX ` + db.dbPrefix + `RETENTION_INDEX4 ON ` + db.dbPrefix + `FILES (ADDED)`
		_, err := db.postgresql.Exec(sqlStatement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *PQDatabase) createFileIndex1() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `FILE_INDEX1 ON ` + db.dbPrefix + `FILES (COLONY_ID, LABEL, NAME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createFileIndex2() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `FILE_INDEX2 ON ` + db.dbPrefix + `FILES (COLONY_ID, FILE_ID)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createFileIndex3() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `FILE_INDEX3 ON ` + db.dbPrefix + `FILES (COLONY_ID, LABEL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) Initialize() error {
	err := db.createColoniesTable()
	if err != nil {
		return err
	}

	err = db.createExecutorsTable()
	if err != nil {
		return err
	}

	err = db.createFunctionsTable()
	if err != nil {
		return err
	}

	err = db.createProcessesTable()
	if err != nil {
		return err
	}

	err = db.createLogTable()
	if err != nil {
		return err
	}

	err = db.createFileTable()
	if err != nil {
		return err
	}

	err = db.createSnapshotTable()
	if err != nil {
		return err
	}

	err = db.createAttributesTable()
	if err != nil {
		return err
	}

	err = db.createProcessGraphsTable()
	if err != nil {
		return err
	}

	err = db.createGeneratorsTable()
	if err != nil {
		return err
	}

	err = db.createGeneratorArgsTable()
	if err != nil {
		return err
	}

	err = db.createCronsTable()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex1()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex2()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex3()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex4()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex5()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex6()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex7()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex8()
	if err != nil {
		return err
	}

	err = db.createAttributesIndex1()
	if err != nil {
		return err
	}

	err = db.createAttributesIndex2()
	if err != nil {
		return err
	}

	err = db.createRetentionIndex1()
	if err != nil {
		return err
	}

	err = db.createRetentionIndex2()
	if err != nil {
		return err
	}

	err = db.createRetentionIndex3()
	if err != nil {
		return err
	}

	err = db.createRetentionIndex4()
	if err != nil {
		return err
	}

	err = db.createFileIndex1()
	if err != nil {
		return err
	}

	err = db.createFileIndex2()
	if err != nil {
		return err
	}

	err = db.createFileIndex3()
	if err != nil {
		return err
	}

	if db.timescaleDB {
		log.Info("Creating TimescaleDB hypertables")
		err := db.createHypertables()
		if err != nil {
			return err
		}
	} else {
		sqlStatement := `CREATE INDEX ` + db.dbPrefix + `LOGS_INDEX1 ON ` + db.dbPrefix + `LOGS (PROCESS_ID)`
		_, err := db.postgresql.Exec(sqlStatement)
		if err != nil {
			return err
		}
	}

	return nil
}
