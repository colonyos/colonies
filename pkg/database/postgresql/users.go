package postgresql

import (
	"database/sql"
	"errors"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddUser(user *core.User) error {
	if user == nil {
		return errors.New("User is nil")
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `USERS (COLONY_NAME, USER_ID, NAME) VALUES ($1, $2, $3)`
	_, err := db.postgresql.Exec(sqlStatement, user.ColonyName, user.ID, user.Name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseUsers(rows *sql.Rows) ([]*core.User, error) {
	var users []*core.User

	for rows.Next() {
		var name string
		var userID string
		var colonyID string
		if err := rows.Scan(&name, &userID, &colonyID); err != nil {
			return nil, err
		}

		user := core.CreateUser(colonyID, userID, name)
		users = append(users, user)
	}

	return users, nil
}

func (db *PQDatabase) GetUsers(colonyName string) ([]*core.User, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseUsers(rows)
}

func (db *PQDatabase) GetUserByID(colonyName string, userID string) (*core.User, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1 AND USER_ID=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users, err := db.parseUsers(rows)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

func (db *PQDatabase) GetUserByName(colonyName string, name string) (*core.User, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users, err := db.parseUsers(rows)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

func (db *PQDatabase) DeleteUserByID(colonyName string, userID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1 AND USER_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, userID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteUserByName(colonyName string, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteUsersByColonyID(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}
