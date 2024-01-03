package postgresql

func (db *PQDatabase) SetServerID(oldServerID, newServerID string) error {
	if oldServerID == "" {
		sqlStatement := `INSERT INTO  ` + db.dbPrefix + `SERVER (SERVER_ID) VALUES ($1)`
		_, err := db.postgresql.Exec(sqlStatement, newServerID)
		if err != nil {
			return err
		}
	} else {
		sqlStatement := `UPDATE ` + db.dbPrefix + `SERVER SET SERVER_ID = $1 WHERE SERVER_ID = $2`
		_, err := db.postgresql.Exec(sqlStatement, newServerID, oldServerID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *PQDatabase) GetServerID() (string, error) {
	sqlStatement := `SELECT SERVER_ID FROM ` + db.dbPrefix + `SERVER`
	row := db.postgresql.QueryRow(sqlStatement)
	var serverID string

	err := row.Scan(&serverID)
	if err != nil {
		return "", err
	}

	return serverID, nil
}
