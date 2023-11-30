package postgresql

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddUser(user *core.User) error {
	if user == nil {
		return errors.New("User is nil")
	}

	existingUser, err := db.GetUserByName(user.ColonyName, user.Name)
	if err != nil {
		return err
	}

	if existingUser != nil {
		return errors.New("User with name <" + user.Name + "> already exists in Colony with name <" + user.ColonyName + ">")
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `USERS (NAME, USER_ID, COLONY_NAME, EMAIL, PHONE) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.postgresql.Exec(sqlStatement, user.ColonyName+":"+user.Name, user.ID, user.ColonyName, user.Email, user.Phone)
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
		var colonyName string
		var email string
		var phone string
		if err := rows.Scan(&name, &userID, &colonyName, &email, &phone); err != nil {
			return nil, err
		}

		s := strings.Split(name, ":")
		if len(s) != 2 {
			return nil, errors.New("Failed to parse User name")
		}
		name = s[1]

		user := core.CreateUser(colonyName, userID, name, email, phone)
		users = append(users, user)
	}

	return users, nil
}

func (db *PQDatabase) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `USERS WHERE NAME=$1 AND COLONY_NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName+":"+name, colonyName)
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

func (db *PQDatabase) RemoveUserByID(colonyName string, userID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1 AND USER_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, userID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveUserByName(colonyName string, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `USERS WHERE NAME=$1 AND COLONY_NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName+":"+name, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveUsersByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `USERS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}
