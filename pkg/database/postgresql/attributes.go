package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) AddAttributes(attributes []core.Attribute) error {
	for _, attribute := range attributes {
		err := db.AddAttribute(attribute)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *PQDatabase) AddAttribute(attribute core.Attribute) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `ATTRIBUTES (ATTRIBUTE_ID, KEY, VALUE, ATTRIBUTE_TYPE, TARGET_ID, TARGET_COLONY_NAME, PROCESSGRAPH_ID, ADDED, STATE) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.postgresql.Exec(sqlStatement, attribute.ID, attribute.Key, attribute.Value, attribute.AttributeType, attribute.TargetID, attribute.TargetColonyName, attribute.TargetProcessGraphID, time.Now(), attribute.State)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseAttributes(rows *sql.Rows) ([]core.Attribute, error) {
	var attributes []core.Attribute

	for rows.Next() {
		var attributeID string
		var key string
		var value string
		var attributeType int
		var targetID string
		var targetColonyName string
		var targetProcessGraphID string
		var added time.Time
		var state int
		if err := rows.Scan(&attributeID, &key, &value, &attributeType, &targetID, &targetColonyName, &targetProcessGraphID, &added, &state); err != nil {
			return nil, err
		}

		attribute := core.CreateAttribute(targetID, targetColonyName, targetProcessGraphID, attributeType, key, value)
		attribute.State = state
		attributes = append(attributes, attribute)
	}

	return attributes, nil
}

func (db *PQDatabase) GetAttributeByID(attributeID string) (core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE ATTRIBUTE_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, attributeID)
	if err != nil {
		return core.Attribute{}, err
	}

	defer rows.Close()

	attributes, err := db.parseAttributes(rows)
	if err != nil {
		return core.Attribute{}, err
	}

	if len(attributes) > 1 {
		return core.Attribute{}, errors.New("Expected attributes to be unique")
	} else if len(attributes) == 0 {
		return core.Attribute{}, errors.New("Attribute does not exists")
	}

	return attributes[0], nil
}

func (db *PQDatabase) GetAttributesByColonyName(colonyName string) ([]core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_COLONY_NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
	if err != nil {
		return []core.Attribute{}, err
	}

	defer rows.Close()

	attributes, err := db.parseAttributes(rows)
	if err != nil {
		return []core.Attribute{}, err
	}

	return attributes, nil
}

func (db *PQDatabase) GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1 AND KEY=$2 AND ATTRIBUTE_TYPE=$3`
	rows, err := db.postgresql.Query(sqlStatement, targetID, key, attributeType)
	if err != nil {
		return core.Attribute{}, err
	}

	defer rows.Close()

	attributes, err := db.parseAttributes(rows)
	if err != nil {
		return core.Attribute{}, err
	}
	if len(attributes) > 1 {
		return core.Attribute{}, errors.New("Expected attributes to be unique")
	} else if len(attributes) == 0 {
		return core.Attribute{}, errors.New("Attribute does not exists")
	}

	return attributes[0], nil
}

func (db *PQDatabase) GetAttributes(targetID string) ([]core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, targetID)
	if err != nil {
		return []core.Attribute{}, err
	}

	defer rows.Close()

	return db.parseAttributes(rows)
}

func (db *PQDatabase) GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1 AND ATTRIBUTE_TYPE=$2`
	rows, err := db.postgresql.Query(sqlStatement, targetID, attributeType)
	if err != nil {
		return []core.Attribute{}, err
	}

	defer rows.Close()

	return db.parseAttributes(rows)
}

func (db *PQDatabase) UpdateAttribute(attribute core.Attribute) error {
	_, err := db.GetAttributeByID(attribute.ID)
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `ATTRIBUTES SET ATTRIBUTE_ID=$1, VALUE=$2`
	_, err = db.postgresql.Exec(sqlStatement, attribute.ID, attribute.Value)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetAttributeState(processID string, state int) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `ATTRIBUTES SET STATE=$1 WHERE TARGET_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, state, processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAttributeByID(attributeID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE ATTRIBUTE_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, attributeID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllAttributesByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_COLONY_NAME=$1 AND STATE=$2 AND PROCESSGRAPH_ID=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, state, "")
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllAttributesByProcessGraphID(processGraphID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE PROCESSGRAPH_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, processGraphID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE PROCESSGRAPH_ID!=$1 AND TARGET_COLONY_NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, "", colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_COLONY_NAME=$1 AND STATE=$2 AND PROCESSGRAPH_ID!=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, state, "")
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAttributesByTargetID(targetID string, attributeType int) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1 AND ATTRIBUTE_TYPE=$2`
	_, err := db.postgresql.Exec(sqlStatement, targetID, attributeType)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllAttributesByTargetID(targetID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE TARGET_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, targetID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllAttributes() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
