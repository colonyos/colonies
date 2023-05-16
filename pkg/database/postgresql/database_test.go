package postgresql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type DBMock struct {
	funcName    string
	returnError bool
}

func (db *DBMock) setReturnError(returnError bool) {
	db.returnError = returnError
}

func (db *DBMock) returnErrorOnCaller(funcName string) {
	db.funcName = funcName
}

func (db *DBMock) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (db *DBMock) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}

func (db *DBMock) Close() error {
	return nil
}

func (db *DBMock) Conn(ctx context.Context) (*sql.Conn, error) {
	return nil, nil
}

func (db *DBMock) Driver() driver.Driver {
	return nil
}

func (db *DBMock) Exec(query string, args ...any) (sql.Result, error) {
	if db.returnError {
		return nil, errors.New("error")
	}

	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		if db.funcName != "" && strings.HasSuffix(details.Name(), db.funcName) {
			return nil, errors.New("error")
		}
	}

	return nil, nil
}

func (db *DBMock) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (db *DBMock) Ping() error {
	return nil
}

func (db *DBMock) PingContext(ctx context.Context) error {
	return nil
}

func (db *DBMock) Prepare(query string) (*sql.Stmt, error) {
	return nil, nil
}

func (db *DBMock) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}

func (db *DBMock) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (db *DBMock) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (db *DBMock) QueryRow(query string, args ...any) *sql.Row {
	return nil
}

func (db *DBMock) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (db *DBMock) SetConnMaxIdleTime(d time.Duration) {

}

func (db *DBMock) SetConnMaxLifetime(d time.Duration) {
}

func (db *DBMock) SetMaxIdleConns(n int) {

}

func (db *DBMock) SetMaxOpenConns(n int) {

}

func (db *DBMock) Stats() sql.DBStats {
	return sql.DBStats{}
}

func TestDropColoniesTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropColoniesTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDropExecutorsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropExecutorsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDropFunctionsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropFunctionsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDropProcessesTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropProcessesTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDropAttributesTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropAttributesTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDropProcessGraphsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropProcessGraphsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDropGeneratorsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropGeneratorsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDropCronsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.dropCronsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.dropColoniesTable()
	assert.Nil(t, err)
}

func TestDrop(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(false)

	dbMock.returnErrorOnCaller("dropColoniesTable")
	err := db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("dropExecutorsTable")
	err = db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("dropFunctionsTable")
	err = db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("dropProcessesTable")
	err = db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("dropAttributesTable")
	err = db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("dropProcessGraphsTable")
	err = db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("dropGeneratorArgsTable")
	err = db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("dropCronsTable")
	err = db.Drop()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("")
	err = db.Drop()
	assert.Nil(t, err)
}

func TestInitialize(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(false)

	dbMock.returnErrorOnCaller("createColoniesTable")
	err := db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createExecutorsTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessesTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createAttributesTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessGraphsTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createAttributesTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessGraphsTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createGeneratorsTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createGeneratorArgsTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createCronsTable")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessesIndex1")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessesIndex2")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessesIndex3")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessesIndex5")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createProcessesIndex6")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createAttributesIndex1")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createAttributesIndex2")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createRetentionIndex1")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createRetentionIndex2")
	err = db.Initialize()
	assert.NotNil(t, err)

	dbMock.returnErrorOnCaller("createRetentionIndex3")
	err = db.Initialize()
	assert.NotNil(t, err)
}

func TestCreateColoniesTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createColoniesTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createColoniesTable()
	assert.Nil(t, err)
}

func TestCreateExecutorsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createExecutorsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createExecutorsTable()
	assert.Nil(t, err)
}

func TestCreateFunctionsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createFunctionsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createFunctionsTable()
	assert.Nil(t, err)
}

func TestCreateProcessesTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessesTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessesTable()
	assert.Nil(t, err)
}

func TestCreateAttributesTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createAttributesTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createAttributesTable()
	assert.Nil(t, err)
}

func TestCreateProcessGraphsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessGraphsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessGraphsTable()
	assert.Nil(t, err)
}

func TestCreateGeneratorsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createGeneratorsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createGeneratorsTable()
	assert.Nil(t, err)
}

func TestCreateGeneratorArgssTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createGeneratorArgsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createGeneratorArgsTable()
	assert.Nil(t, err)
}

func TestCreateCronsTable(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createCronsTable()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createCronsTable()
	assert.Nil(t, err)
}

func TestCreateProcessIndex1(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessesIndex1()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessesIndex1()
	assert.Nil(t, err)
}

func TestCreateProcessIndex2(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessesIndex2()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessesIndex2()
	assert.Nil(t, err)
}

func TestCreateProcessIndex3(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessesIndex3()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessesIndex3()
	assert.Nil(t, err)
}

func TestCreateProcessIndex4(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessesIndex4()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessesIndex4()
	assert.Nil(t, err)
}

func TestCreateProcessIndex5(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessesIndex5()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessesIndex5()
	assert.Nil(t, err)
}

func TestCreateProcessIndex6(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createProcessesIndex6()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createProcessesIndex6()
	assert.Nil(t, err)
}

func TestCreateAttributesIndex1(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createAttributesIndex1()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createAttributesIndex1()
	assert.Nil(t, err)
}

func TestCreateAttributesIndex2(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createAttributesIndex2()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createAttributesIndex2()
	assert.Nil(t, err)
}

func TestCreateRetentionIndex1(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createRetentionIndex1()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createRetentionIndex1()
	assert.Nil(t, err)
}

func TestCreateRetentionIndex2(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createRetentionIndex2()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createRetentionIndex2()
	assert.Nil(t, err)
}

func TestCreateRetentionIndex3(t *testing.T) {
	dbMock := &DBMock{}
	db := &PQDatabase{postgresql: dbMock}
	dbMock.setReturnError(true)
	err := db.createRetentionIndex3()
	assert.NotNil(t, err)

	dbMock.setReturnError(false)
	err = db.createRetentionIndex3()
	assert.Nil(t, err)
}
