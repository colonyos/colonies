package postgresql

import (
	"database/sql"
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddAttributes(attributes []*core.Attribute) error { // TODO: Unit tests
	for _, attribute := range attributes {
		err := db.AddAttribute(attribute)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *PQDatabase) AddAttribute(attribute *core.Attribute) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID, KEY, VALUE, ATTRIBUTE_TYPE, TARGET_ID, TARGET_COLONY_ID, PROCESSGRAPH_ID) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.postgresql.Exec(sqlStatement, attribute.ID, attribute.Key, attribute.Value, attribute.AttributeType, attribute.TargetID, attribute.TargetColonyID, attribute.TargetProcessGraphID)
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
		var attributeType int
		var targetID string
		var targetColonyID string
		var targetProcessGraphID string
		if err := rows.Scan(&attributeID, &key, &value, &attributeType, &targetID, &targetColonyID, &targetProcessGraphID); err != nil {
			return nil, err
		}

		attribute := core.CreateAttribute(targetID, targetColonyID, targetProcessGraphID, attributeType, key, value)
		attributes = append(attributes, attribute)
	}

	return attributes, nil
}

func (db *PQDatabase) GetAttributeByID(attributeID string) (*core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE ATTRIBUTE_ID=$1`
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
		return nil, errors.New("Expected attributes to be unique")
	} else if len(attributes) == 0 {
		return nil, nil
	}

	return attributes[0], nil
}

func (db *PQDatabase) GetAttribute(targetID string, key string, attributeType int) (*core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1 AND KEY=$2 AND ATTRIBUTE_TYPE=$3`
	rows, err := db.postgresql.Query(sqlStatement, targetID, key, attributeType)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	attributes, err := db.parseAttributes(rows)
	if err != nil {
		return nil, err
	}
	if len(attributes) > 1 {
		return nil, errors.New("Expected attributes to be unique")
	} else if len(attributes) == 0 {
		return nil, nil
	}

	return attributes[0], nil
}

func (db *PQDatabase) GetAttributes(targetID string) ([]*core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, targetID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseAttributes(rows)
}

func (db *PQDatabase) GetAttributesByType(targetID string, attributeType int) ([]*core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1 AND ATTRIBUTE_TYPE=$2`
	rows, err := db.postgresql.Query(sqlStatement, targetID, attributeType)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseAttributes(rows)
}

func (db *PQDatabase) UpdateAttribute(attribute *core.Attribute) error {
	existingAttribute, err := db.GetAttributeByID(attribute.ID)
	if err != nil {
		return err
	}
	if existingAttribute == nil {
		return errors.New("Attribute <" + attribute.ID + "> does not exists")
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `ATTRIBUTES SET ATTRIBUTE_ID=$1, VALUE=$2`
	_, err = db.postgresql.Exec(sqlStatement, attribute.ID, attribute.Value)
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

func (db *PQDatabase) DeleteAllAttributesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllAttributesByProcessGraphID(processGraphID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE PROCESSGRAPH_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, processGraphID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllAttributesInProcessGraphsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE PROCESSGRAPH_ID!=$1 AND TARGET_COLONY_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, "", colonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAttributesByTargetID(targetID string, attributeType int) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1 AND ATTRIBUTE_TYPE=$2`
	_, err := db.postgresql.Exec(sqlStatement, targetID, attributeType)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllAttributesByTargetID(targetID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, targetID)
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
