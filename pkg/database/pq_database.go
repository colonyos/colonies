package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

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
	log.Println("Connecting to PostgreSQL database")
	log.Println("   dbHost: " + db.dbHost)
	log.Println("   dbPort: " + strconv.Itoa(db.dbPort))
	log.Println("   dbUser: " + db.dbUser)
	log.Println("   dbPassword: " + "****************")
	log.Println("   dbName: " + db.dbName)
	log.Println("   dbPrefix: " + db.dbPrefix)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		db.dbHost, db.dbPort, db.dbUser, db.dbPassword, db.dbName)

	postgresql, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	db.postgresql = postgresql

	err = db.postgresql.Ping()
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `WORKERS`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `TASKS`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}

	sqlStatement = `DROP TABLE ` + db.dbPrefix + `ATTRIBUTES`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (db *PQDatabase) Initialize() error {
	sqlStatement := `CREATE TABLE ` + db.dbPrefix + `COLONIES (COLONY_ID TEXT PRIMARY KEY NOT NULL, PRIVATE_KEY TEXT NOT NULL, NAME TEXT NOT NULL)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `WORKERS (WORKER_ID TEXT PRIMARY KEY NOT NULL, NAME TEXT NOT NULL, COLONY_ID TEXT NOT NULL, CPU TEXT, CORES INTEGER, MEM INTEGER, GPU TEXT NOT NULL, GPUS INTEGER, STATUS INTEGER)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `TASKS (TASK_ID TEXT PRIMARY KEY NOT NULL, TARGET_COLONY_ID TEXT NOT NULL, TARGET_WORKER_IDS TEXT[], ASSIGNED_WORKER_ID TEXT, STATUS INTEGER, IS_ASSIGNED BOOLEAN, WORKER_TYPE TEXT, SUBMISSION_TIME TIMESTAMP, START_TIME TIMESTAMP, END_TIME TIMESTAMP, DEADLINE TIMESTAMP, TIMEOUT INTEGER, RETRIES INTEGER, MAX_RETRIES INTEGER, LOG TEXT, MEM INTEGER, CORES INTEGER, GPUS INTEGER)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}

	sqlStatement = `CREATE TABLE ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID TEXT PRIMARY KEY NOT NULL, KEY TEXT NOT NULL, VALUE TEXT NOT NULL, ATTRIBUTE_TYPE INTEGER, TASK_ID TEXT NOT NULL)`
	_, err = db.postgresql.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}

	return nil
}
