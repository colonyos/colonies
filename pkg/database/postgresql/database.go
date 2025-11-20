package postgresql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"strings"
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

func (db *PQDatabase) dropUsersTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `USERS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
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

func (db *PQDatabase) dropNodesTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `NODES`
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

func (db *PQDatabase) dropBlueprintDefinitionsTable() error {
	sqlStatement := `DROP TABLE IF EXISTS ` + db.dbPrefix + `BLUEPRINTDEFINITIONS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropBlueprintsTable() error {
	sqlStatement := `DROP TABLE IF EXISTS ` + db.dbPrefix + `BLUEPRINTS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) dropServerTable() error {
	sqlStatement := `DROP TABLE ` + db.dbPrefix + `SERVER`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) Drop() error {
	err := db.dropUsersTable()
	if err != nil {
		return err
	}

	err = db.dropColoniesTable()
	if err != nil {
		return err
	}

	err = db.dropExecutorsTable()
	if err != nil {
		return err
	}

	err = db.dropNodesTable()
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

	err = db.dropBlueprintDefinitionsTable()
	if err != nil {
		return err
	}

	err = db.dropBlueprintsTable()
	if err != nil {
		return err
	}

	err = db.dropServerTable()
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createHypertables() error {
	prefix := strings.ToLower(db.dbPrefix)
	//sqlStatement := `SELECT create_hypertable ('` + prefix + `logs', 'ts', chunk_time_interval => 86400000000000)` // 24h chunks, assuming ts is nanosec
	sqlStatement := `SELECT create_hypertable ('` + prefix + `logs', by_range('added', INTERVAL '1 day'))`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createServerTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `SERVER (SERVER_ID TEXT PRIMARY KEY NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createColoniesTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `COLONIES (NAME TEXT PRIMARY KEY NOT NULL, COLONY_ID TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createUsersTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `USERS (NAME TEXT PRIMARY KEY NOT NULL, USER_ID TEXT NOT NULL, COLONY_NAME TEXT NOT NULL, EMAIL TEXT NOT NULL, PHONE TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createExecutorsTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `EXECUTORS (NAME TEXT PRIMARY KEY NOT NULL, EXECUTOR_TYPE TEXT NOT NULL, EXECUTOR_ID TEXT NOT NULL, COLONY_NAME TEXT NOT NULL, STATE INTEGER, REQUIRE_FUNC_REG BOOLEAN, COMMISSIONTIME TIMESTAMPTZ, LASTHEARDFROM TIMESTAMPTZ, LONG DOUBLE PRECISION, LAT DOUBLE PRECISION, LOCDESC TEXT, HWMODEL TEXT, HWNODES INT, HWCPU TEXT, HWMEM TEXT, HWSTORAGE TEXT, HWGPUNAME TEXT, HWGPUCOUNT TEXT, HWGPUNODECOUNT INTEGER, HWGPUMEM TEXT, SWNAME TEXT, SWTYPE TEXT, SWVERSION TEXT, ALLOCATIONS TEXT NOT NULL, NODE_ID TEXT)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	// Add NODE_ID column to existing tables if it doesn't exist
	alterStatement := `ALTER TABLE ` + db.dbPrefix + `EXECUTORS ADD COLUMN IF NOT EXISTS NODE_ID TEXT`
	_, err = db.postgresql.Exec(alterStatement)
	if err != nil {
		return err
	}

	// Add BLUEPRINT_ID column to existing tables if it doesn't exist
	alterStatement = `ALTER TABLE ` + db.dbPrefix + `EXECUTORS ADD COLUMN IF NOT EXISTS BLUEPRINT_ID TEXT`
	_, err = db.postgresql.Exec(alterStatement)
	if err != nil {
		return err
	}

	// Add BLUEPRINT_GEN column to existing tables if it doesn't exist
	alterStatement = `ALTER TABLE ` + db.dbPrefix + `EXECUTORS ADD COLUMN IF NOT EXISTS BLUEPRINT_GEN BIGINT`
	_, err = db.postgresql.Exec(alterStatement)
	if err != nil {
		return err
	}

	// Add HWPLATFORM column to existing tables if it doesn't exist
	alterStatement = `ALTER TABLE ` + db.dbPrefix + `EXECUTORS ADD COLUMN IF NOT EXISTS HWPLATFORM TEXT`
	_, err = db.postgresql.Exec(alterStatement)
	if err != nil {
		return err
	}

	// Add HWARCHITECTURE column to existing tables if it doesn't exist
	alterStatement = `ALTER TABLE ` + db.dbPrefix + `EXECUTORS ADD COLUMN IF NOT EXISTS HWARCHITECTURE TEXT`
	_, err = db.postgresql.Exec(alterStatement)
	if err != nil {
		return err
	}

	// Add HWNETWORK column to existing tables if it doesn't exist (JSON array of network addresses)
	alterStatement = `ALTER TABLE ` + db.dbPrefix + `EXECUTORS ADD COLUMN IF NOT EXISTS HWNETWORK TEXT`
	_, err = db.postgresql.Exec(alterStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createNodesTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `NODES (ID TEXT PRIMARY KEY NOT NULL, NAME TEXT NOT NULL, COLONY_NAME TEXT NOT NULL, LOCATION TEXT, PLATFORM TEXT, ARCHITECTURE TEXT, CPU INTEGER, MEMORY BIGINT, GPU INTEGER, CAPABILITIES TEXT[], LABELS JSONB, EXECUTORS TEXT[], STATE TEXT, LAST_SEEN TIMESTAMPTZ, CREATED TIMESTAMPTZ)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}
	return nil
}

func (db *PQDatabase) createFunctionsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `FUNCTIONS (FUNCTION_ID TEXT PRIMARY KEY NOT NULL, EXECUTOR_NAME TEXT NOT NULL, EXECUTOR_TYPE TEXT NOT NULL, COLONY_NAME TEXT NOT NULL, FUNCNAME TEXT NOT NULL, COUNTER INTEGER, MINWAITTIME FLOAT, MAXWAITTIME FLOAT, MINEXECTIME FLOAT, MAXEXECTIME FLOAT, AVGWAITTIME FLOAT, AVGEXECTIME FLOAT)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `PROCESSES (PROCESS_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_NAME TEXT NOT NULL, TARGET_EXECUTOR_NAMES TEXT[], ASSIGNED_EXECUTOR_ID TEXT, STATE INTEGER, IS_ASSIGNED BOOLEAN, EXECUTOR_TYPE TEXT, SUBMISSION_TIME TIMESTAMPTZ, START_TIME TIMESTAMPTZ, END_TIME TIMESTAMPTZ, WAIT_DEADLINE TIMESTAMPTZ, EXEC_DEADLINE TIMESTAMPTZ, ERRORS TEXT[], NODENAME TEXT, FUNCNAME TEXT, ARGS TEXT, KWARGS TEXT, MAX_WAIT_TIME INTEGER, MAX_EXEC_TIME INTEGER, RETRIES INTEGER, MAX_RETRIES INTEGER, DEPENDENCIES TEXT[], PRIORITY INTEGER, PRIORITYTIME BIGINT, WAIT_FOR_PARENTS BOOLEAN, PARENTS TEXT[], CHILDREN TEXT[], PROCESSGRAPH_ID TEXT, INPUT TEXT, OUTPUT TEXT, LABEL TEXT, FS TEXT, NODES INTEGER, CPU BIGINT, PROCESSES INTEGER, PROCESSES_PER_NODE INTEGER, MEMORY BIGINT, STORAGE BIGINT, GPUNAME TEXT, GPUCOUNT TEXT, GPUMEM BIGINT, WALLTIME BIGINT, INITIATOR_ID TEXT NOT NULL, INITIATOR_NAME TEXT NOT NULL, RECONCILIATION TEXT, BLUEPRINT TEXT)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createLogTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `LOGS (PROCESS_ID TEXT, COLONY_NAME TEXT NOT NULL, EXECUTOR_NAME TEXT NOT NULL, TS BIGINT, MSG TEXT NOT NULL, ADDED TIMESTAMPTZ)`
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

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `FILES (FILE_ID TEXT PRIMARY KEY NOT NULL, COLONY_NAME TEXT NOT NULL, LABEL TEXT NOT NULL, NAME TEXT NOT NULL, SIZE BIGINT, SEQNR BIGINT, CHECKSUM TEXT, CHECKSUM_ALG TEXT, ADDED TIMESTAMPTZ, PROTOCOL TEXT, S3_SERVER TEXT, S3_PORT INTEGER, S3_TLS BOOLEAN, S3_ACCESSKEY TEXT, S3_SECRETKEY TEXT, S3_REGION TEXT, S3_ENCKEY TEXT, S3_ENCALG TEXT, S3_OBJ TEXT, S3_BUCKET TEXT)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createSnapshotTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `SNAPSHOTS (SNAPSHOT_ID TEXT PRIMARY KEY NOT NULL, COLONY_NAME TEXT NOT NULL, LABEL TEXT NOT NULL, NAME TEXT NOT NULL UNIQUE, FILE_IDS TEXT[], ADDED TIMESTAMPTZ)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createAttributesTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID TEXT PRIMARY KEY NOT NULL, KEY TEXT NOT NULL, VALUE TEXT NOT NULL, ATTRIBUTE_TYPE INTEGER, TARGET_ID TEXT NOT NULL, TARGET_COLONY_NAME TEXT NOT NULL, PROCESSGRAPH_ID TEXT NOT NULL, ADDED TIMESTAMPTZ, STATE INTEGER)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessGraphsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `PROCESSGRAPHS (PROCESSGRAPH_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_NAME TEXT NOT NULL, ROOTS TEXT[], STATE INTEGER, SUBMISSION_TIME TIMESTAMPTZ, START_TIME TIMESTAMPTZ, END_TIME TIMESTAMPTZ, INITIATOR_ID TEXT NOT NULL, INITIATOR_NAME TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createGeneratorsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `GENERATORS (GENERATOR_ID TEXT PRIMARY KEY NOT NULL, COLONY_NAME TEXT NOT NULL, NAME TEXT NOT NULL, WORKFLOW_SPEC TEXT NOT NULL, TRIGGER INTEGER, TIMEOUT INTEGER, LASTRUN TIMESTAMPTZ, FIRSTPACK TIMESTAMPTZ, INITIATOR_ID TEXT NOT NULL, INITIATOR_NAME TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createGeneratorArgsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `GENERATORARGS (GENERATORARG_ID TEXT PRIMARY KEY NOT NULL, GENERATOR_ID TEXT NOT NULL, COLONY_NAME TEXT NOT NULL, ARG TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createCronsTable() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `CRONS (CRON_ID TEXT PRIMARY KEY NOT NULL, COLONY_NAME TEXT NOT NULL, NAME TEXT NOT NULL UNIQUE, CRON_EXPR TEXT NOT NULL, INTERVAL INT, RANDOM BOOLEAN, NEXT_RUN TIMESTAMPTZ, LAST_RUN TIMESTAMPTZ, WORKFLOW_SPEC TEXT NOT NULL, PREV_PROCESSGRAPH_ID TEXT NOT NULL, WAIT_FOR_PREV_PROCESSGRAPH BOOLEAN, INITIATOR_ID TEXT NOT NULL, INITIATOR_NAME TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createBlueprintDefinitionsTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `BLUEPRINTDEFINITIONS (ID TEXT PRIMARY KEY NOT NULL, COLONY_NAME TEXT NOT NULL, NAME TEXT NOT NULL, API_GROUP TEXT NOT NULL, VERSION TEXT NOT NULL, KIND TEXT NOT NULL, DATA TEXT NOT NULL, UNIQUE(COLONY_NAME, NAME))`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createBlueprintsTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `BLUEPRINTS (ID TEXT PRIMARY KEY NOT NULL, COLONY_NAME TEXT NOT NULL, NAME TEXT NOT NULL, KIND TEXT NOT NULL, DATA TEXT NOT NULL, UNIQUE(COLONY_NAME, NAME))`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createBlueprintHistoryTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `BLUEPRINT_HISTORY (
		ID TEXT PRIMARY KEY NOT NULL,
		BLUEPRINT_ID TEXT NOT NULL,
		KIND TEXT NOT NULL,
		NAMESPACE TEXT NOT NULL,
		NAME TEXT NOT NULL,
		GENERATION BIGINT NOT NULL,
		SPEC TEXT NOT NULL,
		STATUS TEXT,
		TIMESTAMP TIMESTAMP NOT NULL,
		CHANGED_BY TEXT NOT NULL,
		CHANGE_TYPE TEXT NOT NULL
	)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	// Create index for faster queries by resource_id
	indexStatement := `CREATE INDEX IF NOT EXISTS ` + db.dbPrefix + `BLUEPRINT_HISTORY_INDEX1 ON ` + db.dbPrefix + `BLUEPRINT_HISTORY (BLUEPRINT_ID, TIMESTAMP DESC)`
	_, err = db.postgresql.Exec(indexStatement)
	return err
}

func (db *PQDatabase) createProcessesIndex1() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX1 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_NAME, STATE, SUBMISSION_TIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex2() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX2 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_NAME, STATE, START_TIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex3() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX3 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_NAME, STATE, END_TIME)`
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
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX6 ON ` + db.dbPrefix + `PROCESSES (STATE, EXECUTOR_TYPE, IS_ASSIGNED, WAIT_FOR_PARENTS, TARGET_COLONY_NAME, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex7() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX7 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_NAME, STATE, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex8() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX8 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_NAME, STATE, EXECUTOR_TYPE, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex9() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX9 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_NAME, STATE, INITIATOR_NAME, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex10() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX10 ON ` + db.dbPrefix + `PROCESSES (TARGET_COLONY_NAME, STATE, LABEL, PRIORITYTIME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createProcessesIndex11() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `PROCESSES_INDEX11 ON ` + db.dbPrefix + `PROCESSES (STATE, EXECUTOR_TYPE, IS_ASSIGNED, WAIT_FOR_PARENTS, TARGET_COLONY_NAME, EXECUTOR_TYPE, IS_ASSIGNED, TARGET_EXECUTOR_NAMES, CPU, MEMORY, GPUNAME, GPUMEM, GPUCOUNT, STORAGE, NODES, PROCESSES, PROCESSES_PER_NODE, PRIORITYTIME)`
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
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `FILE_INDEX1 ON ` + db.dbPrefix + `FILES (COLONY_NAME, LABEL, NAME)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createFileIndex2() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `FILE_INDEX2 ON ` + db.dbPrefix + `FILES (COLONY_NAME, FILE_ID)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) createFileIndex3() error {
	sqlStatement := `CREATE INDEX ` + db.dbPrefix + `FILE_INDEX3 ON ` + db.dbPrefix + `FILES (COLONY_NAME, LABEL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) Initialize() error {
	err := db.createUsersTable()
	if err != nil {
		return err
	}

	err = db.createServerTable()
	if err != nil {
		return err
	}

	err = db.createColoniesTable()
	if err != nil {
		return err
	}

	err = db.createExecutorsTable()
	if err != nil {
		return err
	}

	err = db.createNodesTable()
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

	err = db.createBlueprintDefinitionsTable()
	if err != nil {
		return err
	}

	err = db.createBlueprintsTable()
	if err != nil {
		return err
	}

	err = db.createBlueprintHistoryTable()
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

	err = db.createProcessesIndex9()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex10()
	if err != nil {
		return err
	}

	err = db.createProcessesIndex11()
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
