package database

import (
	"colonies/pkg/core"
	"database/sql"
	"errors"
)

func (db *PQDatabase) AddAttribute(attribute *core.Attribute) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID, KEY, VALUE, ATTRIBUTE_TYPE, TASK_ID) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.postgresql.Exec(sqlStatement, attribute.ID(), attribute.Key(), attribute.Value(), attribute.AttributeType(), attribute.TaskID())
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseAttributes(rows *sql.Rows) ([]*core.Attribute, error) {
	var attributes []*core.Attribute

	for rows.Next() {
		var attributeID string
		var key string
		var value string
		var taskType int
		var taskID string
		if err := rows.Scan(&attributeID, &key, &value, &taskType, &taskID); err != nil {
			return nil, err
		}

		// No need to pass attribute ID as it is derived from taskID and key
		attribute := core.CreateAttribute(taskID, taskType, key, value)
		attributes = append(attributes, attribute)
	}

	return attributes, nil
}

func (db *PQDatabase) GetAttributeByID(attributeID string) (*core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES where ATTRIBUTE_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, attributeID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	attributes, err := db.parseAttributes(rows)
	if err != nil {
		return nil, err
	}
	if len(attributes) > 1 {
		return nil, errors.New("expected attributes to be unique")
	} else if len(attributes) == 0 {
		return nil, nil
	}

	return attributes[0], nil
}

func (db *PQDatabase) GetAttribute(taskID string, key string, attributeType int) (*core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES where TASK_ID=$1 AND KEY=$2 AND ATTRIBUTE_TYPE=$3`
	rows, err := db.postgresql.Query(sqlStatement, taskID, key, attributeType)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	attributes, err := db.parseAttributes(rows)
	if err != nil {
		return nil, err
	}
	if len(attributes) > 1 {
		return nil, errors.New("expected attributes to be unique")
	} else if len(attributes) == 0 {
		return nil, nil
	}

	return attributes[0], nil
}

func (db *PQDatabase) GetAttributes(taskID string, attributeType int) ([]*core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES where TASK_ID=$1 AND ATTRIBUTE_TYPE=$2`
	rows, err := db.postgresql.Query(sqlStatement, taskID, attributeType)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseAttributes(rows)
}

func (db *PQDatabase) UpdateAttribute(attribute *core.Attribute) error {
	existingAttribute, err := db.GetAttributeByID(attribute.ID())
	if err != nil {
		return err
	}
	if existingAttribute == nil {
		return errors.New("attribute <" + attribute.ID() + "> does not exists")
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `ATTRIBUTES SET ATTRIBUTE_ID=$1, VALUE=$2`
	_, err = db.postgresql.Exec(sqlStatement, attribute.ID(), attribute.Value())
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAttributeByID(attributeID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE ATTRIBUTE_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, attributeID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAttributesByTaskID(taskID string, attributeType int) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TASK_ID=$1 AND ATTRIBUTE_TYPE=$2`
	_, err := db.postgresql.Exec(sqlStatement, taskID, attributeType)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllAttributesByTaskID(taskID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TASK_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, taskID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllAttributes() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
