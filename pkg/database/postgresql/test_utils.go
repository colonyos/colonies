package postgresql

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// testTables lists every table created by Initialize. Kept in sync with
// Initialize so test cleanup can truncate instead of recreating the schema.
var testTables = []string{
	"USERS",
	"SERVER",
	"COLONIES",
	"EXECUTORS",
	"NODES",
	"FUNCTIONS",
	"PROCESSES",
	"LOGS",
	"FILES",
	"SNAPSHOTS",
	"ATTRIBUTES",
	"PROCESSGRAPHS",
	"GENERATORS",
	"GENERATORARGS",
	"CRONS",
	"BLUEPRINTDEFINITIONS",
	"BLUEPRINTS",
	"LOCATIONS",
	"BLUEPRINT_HISTORY",
}

var testSchemas = struct {
	sync.Mutex
	ready map[string]bool
}{ready: make(map[string]bool)}

var testDatabase struct {
	once sync.Once
	name string
	err  error
}

// pinnedTestConn holds one open connection to this process's test database for
// the lifetime of the process, so the database shows up in pg_stat_activity
// and is never considered stale by other test processes.
var pinnedTestConn *sql.DB

// testDatabaseName returns the name of a database dedicated to this test
// process, creating it on first use. Per-process databases let test packages
// run in parallel against one Postgres instance without interfering. Stale
// databases from finished or crashed runs are dropped opportunistically: a
// database is stale only if it has no active connections AND the process
// that created it (encoded in the name) is no longer alive on this host.
func testDatabaseName(host string, port int, user string, password string) (string, error) {
	testDatabase.once.Do(func() {
		adminDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable", host, port, user, password)
		admin, err := sql.Open("postgres", adminDSN)
		if err != nil {
			testDatabase.err = err
			return
		}
		defer admin.Close()

		rows, err := admin.Query(`SELECT datname FROM pg_database WHERE datname LIKE 'colonies_test_%' AND datname NOT IN (SELECT datname FROM pg_stat_activity WHERE datname IS NOT NULL)`)
		if err == nil {
			var stale []string
			for rows.Next() {
				var name string
				if rows.Scan(&name) == nil && !testDatabaseOwnerAlive(name) {
					stale = append(stale, name)
				}
			}
			rows.Close()
			for _, name := range stale {
				// Best effort; a database that just became active is skipped
				admin.Exec(`DROP DATABASE IF EXISTS ` + name)
			}
		}

		name := fmt.Sprintf("colonies_test_%d", os.Getpid())
		if _, err := admin.Exec(`DROP DATABASE IF EXISTS ` + name); err != nil {
			testDatabase.err = err
			return
		}
		if _, err := admin.Exec(`CREATE DATABASE ` + name); err != nil {
			testDatabase.err = err
			return
		}

		// Pin a connection so the new database is visible as in-use
		pinnedDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
		pinnedTestConn, err = sql.Open("postgres", pinnedDSN)
		if err == nil {
			pinnedTestConn.SetMaxIdleConns(1)
			pinnedTestConn.SetConnMaxLifetime(0)
			pinnedTestConn.SetConnMaxIdleTime(0)
			testDatabase.err = pinnedTestConn.Ping()
			if testDatabase.err != nil {
				return
			}
		}

		testDatabase.name = name
	})

	return testDatabase.name, testDatabase.err
}

// testDatabaseOwnerAlive reports whether the process that created a test
// database (pid encoded in the database name) is still running on this host.
func testDatabaseOwnerAlive(datname string) bool {
	var pid int
	if _, err := fmt.Sscanf(datname, "colonies_test_%d", &pid); err != nil {
		// Unrecognized name; leave it alone
		return true
	}
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return err == nil
}

func PrepareTests() (*PQDatabase, error) {
	return PrepareTestsWithPrefix("TEST_")
}

// PrepareTestsWithPrefix returns a connected database with an empty schema.
// The schema is dropped and recreated only on the first call per process and
// prefix; subsequent calls truncate all tables, which is roughly an order of
// magnitude faster than Drop+Initialize.
func PrepareTestsWithPrefix(prefix string) (*PQDatabase, error) {
	log.SetOutput(io.Discard)

	dbHost := os.Getenv("COLONIES_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := 5432
	dbUser := os.Getenv("COLONIES_DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("COLONIES_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
	}
	dbName, err := testDatabaseName(dbHost, dbPort, dbUser, dbPassword)
	if err != nil {
		return nil, err
	}

	db := CreatePQDatabase(dbHost, dbPort, dbUser, dbPassword, dbName, prefix, false)

	// Keep pools small so parallel test packages do not exhaust the Postgres
	// server's connection limit
	if os.Getenv("COLONIES_DB_MAX_OPEN_CONNS") == "" {
		os.Setenv("COLONIES_DB_MAX_OPEN_CONNS", "25")
	}
	if os.Getenv("COLONIES_DB_MAX_IDLE_CONNS") == "" {
		os.Setenv("COLONIES_DB_MAX_IDLE_CONNS", "5")
	}

	err = db.Connect()
	if err != nil {
		return nil, err
	}

	testSchemas.Lock()
	defer testSchemas.Unlock()

	if testSchemas.ready[prefix] {
		return db, db.clearTestData()
	}

	db.Drop()
	err = db.Initialize()
	if err != nil {
		return nil, err
	}

	testSchemas.ready[prefix] = true

	return db, nil
}

// clearTestData empties all tables and resets the file sequence, giving each
// test a clean database without paying for schema recreation.
func (db *PQDatabase) clearTestData() error {
	tables := make([]string, len(testTables))
	for i, table := range testTables {
		tables[i] = db.dbPrefix + table
	}

	_, err := db.postgresql.Exec(`TRUNCATE TABLE ` + strings.Join(tables, ", "))
	if err != nil {
		return err
	}

	_, err = db.postgresql.Exec(`ALTER SEQUENCE ` + db.dbPrefix + `FILE_SEQ RESTART WITH 1`)

	return err
}
