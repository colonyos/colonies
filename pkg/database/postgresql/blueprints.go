package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

// BlueprintDefinition methods

func (db *PQDatabase) AddBlueprintDefinition(sd *core.BlueprintDefinition) error {
	if sd == nil {
		return errors.New("BlueprintDefinition is nil")
	}

	sdJSON, err := sd.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `BLUEPRINTDEFINITIONS (ID, COLONY_NAME, NAME, API_GROUP, VERSION, KIND, DATA) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.postgresql.Exec(sqlStatement, sd.ID, sd.Metadata.ColonyName, sd.Metadata.Name, sd.Spec.Group, sd.Spec.Version, sd.Spec.Names.Kind, sdJSON)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseBlueprintDefinitions(rows *sql.Rows) ([]*core.BlueprintDefinition, error) {
	var sds []*core.BlueprintDefinition

	for rows.Next() {
		var id string
		var colonyName string
		var name string
		var apiGroup string
		var version string
		var kind string
		var data string

		if err := rows.Scan(&id, &colonyName, &name, &apiGroup, &version, &kind, &data); err != nil {
			return nil, err
		}

		sd, err := core.ConvertJSONToBlueprintDefinition(data)
		if err != nil {
			return nil, err
		}

		sds = append(sds, sd)
	}

	return sds, nil
}

func (db *PQDatabase) GetBlueprintDefinitionByID(id string) (*core.BlueprintDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sds, err := db.parseBlueprintDefinitions(rows)
	if err != nil {
		return nil, err
	}

	if len(sds) == 0 {
		return nil, nil
	}

	return sds[0], nil
}

func (db *PQDatabase) GetBlueprintDefinitionByName(namespace, name string) (*core.BlueprintDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, namespace, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sds, err := db.parseBlueprintDefinitions(rows)
	if err != nil {
		return nil, err
	}

	if len(sds) == 0 {
		return nil, nil
	}

	return sds[0], nil
}

func (db *PQDatabase) GetBlueprintDefinitions() ([]*core.BlueprintDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprintDefinitions(rows)
}

func (db *PQDatabase) GetBlueprintDefinitionsByNamespace(namespace string) ([]*core.BlueprintDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprintDefinitions(rows)
}

func (db *PQDatabase) GetBlueprintDefinitionsByGroup(group string) ([]*core.BlueprintDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS WHERE API_GROUP=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, group)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprintDefinitions(rows)
}

func (db *PQDatabase) UpdateBlueprintDefinition(sd *core.BlueprintDefinition) error {
	if sd == nil {
		return errors.New("BlueprintDefinition is nil")
	}

	sdJSON, err := sd.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `BLUEPRINTDEFINITIONS SET NAME=$1, API_GROUP=$2, VERSION=$3, KIND=$4, DATA=$5 WHERE ID=$6`
	_, err = db.postgresql.Exec(sqlStatement, sd.Metadata.Name, sd.Spec.Group, sd.Spec.Version, sd.Spec.Names.Kind, sdJSON, sd.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveBlueprintDefinitionByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveBlueprintDefinitionByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountBlueprintDefinitions() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `BLUEPRINTDEFINITIONS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return -1, err
		}
	}

	return count, nil
}

// Blueprint methods

func (db *PQDatabase) AddBlueprint(blueprint *core.Blueprint) error {
	if blueprint == nil {
		return errors.New("Blueprint is nil")
	}

	existingBlueprint, err := db.GetBlueprintByName(blueprint.Metadata.ColonyName, blueprint.Metadata.Name)
	if err != nil {
		return err
	}

	if existingBlueprint != nil {
		return errors.New("Blueprint with name <" + blueprint.Metadata.Name + "> in namespace <" + blueprint.Metadata.ColonyName + "> already exists")
	}

	blueprintJSON, err := blueprint.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `BLUEPRINTS (ID, COLONY_NAME, NAME, KIND, DATA) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.postgresql.Exec(sqlStatement, blueprint.ID, blueprint.Metadata.ColonyName, blueprint.Metadata.Name, blueprint.Kind, blueprintJSON)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseBlueprints(rows *sql.Rows) ([]*core.Blueprint, error) {
	var blueprints []*core.Blueprint

	for rows.Next() {
		var id string
		var colonyName string
		var name string
		var kind string
		var data string

		if err := rows.Scan(&id, &colonyName, &name, &kind, &data); err != nil {
			return nil, err
		}

		blueprint, err := core.ConvertJSONToBlueprint(data)
		if err != nil {
			return nil, err
		}

		// Set the ID from the database (not stored in the JSON DATA column)
		blueprint.ID = id

		blueprints = append(blueprints, blueprint)
	}

	return blueprints, nil
}

func (db *PQDatabase) GetBlueprintByID(id string) (*core.Blueprint, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	blueprints, err := db.parseBlueprints(rows)
	if err != nil {
		return nil, err
	}

	if len(blueprints) == 0 {
		return nil, nil
	}

	return blueprints[0], nil
}

func (db *PQDatabase) GetBlueprintByName(namespace, name string) (*core.Blueprint, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, namespace, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	blueprints, err := db.parseBlueprints(rows)
	if err != nil {
		return nil, err
	}

	if len(blueprints) == 0 {
		return nil, nil
	}

	return blueprints[0], nil
}

func (db *PQDatabase) GetBlueprints() ([]*core.Blueprint, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprints(rows)
}

func (db *PQDatabase) GetBlueprintsByNamespace(namespace string) ([]*core.Blueprint, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprints(rows)
}

func (db *PQDatabase) GetBlueprintsByKind(kind string) ([]*core.Blueprint, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE KIND=$1 ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprints(rows)
}

func (db *PQDatabase) GetBlueprintsByNamespaceAndKind(namespace, kind string) ([]*core.Blueprint, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1 AND KIND=$2 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprints(rows)
}

func (db *PQDatabase) GetBlueprintsByNamespaceKindAndLocation(namespace, kind, locationName string) ([]*core.Blueprint, error) {
	var rows *sql.Rows
	var err error

	if locationName == "" {
		// If no location filter, return all blueprints for namespace and kind
		sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1 AND KIND=$2 ORDER BY NAME`
		rows, err = db.postgresql.Query(sqlStatement, namespace, kind)
	} else {
		// Filter by location using JSONB query (case-insensitive)
		sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1 AND KIND=$2 AND LOWER((DATA::jsonb)->'metadata'->>'locationname')=LOWER($3) ORDER BY NAME`
		rows, err = db.postgresql.Query(sqlStatement, namespace, kind, locationName)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseBlueprints(rows)
}

func (db *PQDatabase) UpdateBlueprint(blueprint *core.Blueprint) error {
	if blueprint == nil {
		return errors.New("Blueprint is nil")
	}

	blueprintJSON, err := blueprint.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `BLUEPRINTS SET COLONY_NAME=$1, NAME=$2, KIND=$3, DATA=$4 WHERE ID=$5`
	_, err = db.postgresql.Exec(sqlStatement, blueprint.Metadata.ColonyName, blueprint.Metadata.Name, blueprint.Kind, blueprintJSON, blueprint.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) UpdateBlueprintStatus(id string, status map[string]interface{}) error {
	// Get the current blueprint data to preserve spec and metadata
	sqlStatement := `SELECT DATA FROM ` + db.dbPrefix + `BLUEPRINTS WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return errors.New("Blueprint not found")
	}

	var dataStr string
	if err := rows.Scan(&dataStr); err != nil {
		return err
	}

	// Parse the JSON data
	var blueprintData map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &blueprintData); err != nil {
		return fmt.Errorf("failed to unmarshal blueprint data: %w", err)
	}

	// Update only the status field
	blueprintData["status"] = status

	// Convert back to JSON
	updatedJSON, err := json.Marshal(blueprintData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated blueprint data: %w", err)
	}

	// Update the database
	// Note: This still has a potential race condition, but it's much smaller window
	// than the original read-modify-write pattern
	updateStatement := `UPDATE ` + db.dbPrefix + `BLUEPRINTS SET DATA=$1 WHERE ID=$2`
	result, err := db.postgresql.Exec(updateStatement, string(updatedJSON), id)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Blueprint not found")
	}

	return nil
}

func (db *PQDatabase) RemoveBlueprintByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `BLUEPRINTS WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveBlueprintByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveBlueprintsByNamespace(namespace string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountBlueprints() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `BLUEPRINTS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return -1, err
		}
	}

	return count, nil
}

func (db *PQDatabase) CountBlueprintsByNamespace(namespace string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `BLUEPRINTS WHERE COLONY_NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return -1, err
		}
	}

	return count, nil
}

// AddBlueprintHistory adds a new BlueprintHistory entry
func (db *PQDatabase) AddBlueprintHistory(history *core.BlueprintHistory) error {
	specJSON, err := json.Marshal(history.Spec)
	if err != nil {
		return err
	}

	statusJSON, err := json.Marshal(history.Status)
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `BLUEPRINT_HISTORY (
		ID, BLUEPRINT_ID, KIND, NAMESPACE, NAME, GENERATION, SPEC, STATUS,
		TIMESTAMP, CHANGED_BY, CHANGE_TYPE)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.postgresql.Exec(
		sqlStatement,
		history.ID,
		history.BlueprintID,
		history.Kind,
		history.Namespace,
		history.Name,
		history.Generation,
		specJSON,
		statusJSON,
		history.Timestamp,
		history.ChangedBy,
		history.ChangeType,
	)

	return err
}

// GetBlueprintHistory retrieves history for a blueprint (most recent first)
func (db *PQDatabase) GetBlueprintHistory(blueprintID string, limit int) ([]*core.BlueprintHistory, error) {
	sqlStatement := `SELECT ID, BLUEPRINT_ID, KIND, NAMESPACE, NAME, GENERATION,
		SPEC, STATUS, TIMESTAMP, CHANGED_BY, CHANGE_TYPE
		FROM ` + db.dbPrefix + `BLUEPRINT_HISTORY
		WHERE BLUEPRINT_ID=$1
		ORDER BY TIMESTAMP DESC`

	if limit > 0 {
		sqlStatement += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.postgresql.Query(sqlStatement, blueprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*core.BlueprintHistory
	for rows.Next() {
		var history core.BlueprintHistory
		var specJSON, statusJSON []byte

		err := rows.Scan(
			&history.ID,
			&history.BlueprintID,
			&history.Kind,
			&history.Namespace,
			&history.Name,
			&history.Generation,
			&specJSON,
			&statusJSON,
			&history.Timestamp,
			&history.ChangedBy,
			&history.ChangeType,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(specJSON, &history.Spec); err != nil {
			return nil, err
		}

		if len(statusJSON) > 0 {
			if err := json.Unmarshal(statusJSON, &history.Status); err != nil {
				return nil, err
			}
		}

		histories = append(histories, &history)
	}

	return histories, nil
}

// GetBlueprintHistoryByGeneration retrieves a specific generation of a blueprint
func (db *PQDatabase) GetBlueprintHistoryByGeneration(blueprintID string, generation int64) (*core.BlueprintHistory, error) {
	sqlStatement := `SELECT ID, BLUEPRINT_ID, KIND, NAMESPACE, NAME, GENERATION,
		SPEC, STATUS, TIMESTAMP, CHANGED_BY, CHANGE_TYPE
		FROM ` + db.dbPrefix + `BLUEPRINT_HISTORY
		WHERE BLUEPRINT_ID=$1 AND GENERATION=$2`

	rows, err := db.postgresql.Query(sqlStatement, blueprintID, generation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history core.BlueprintHistory
	var specJSON, statusJSON []byte

	if rows.Next() {
		err := rows.Scan(
			&history.ID,
			&history.BlueprintID,
			&history.Kind,
			&history.Namespace,
			&history.Name,
			&history.Generation,
			&specJSON,
			&statusJSON,
			&history.Timestamp,
			&history.ChangedBy,
			&history.ChangeType,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(specJSON, &history.Spec); err != nil {
			return nil, err
		}

		if len(statusJSON) > 0 {
			if err := json.Unmarshal(statusJSON, &history.Status); err != nil {
				return nil, err
			}
		}

		return &history, nil
	}

	return nil, nil
}

// RemoveBlueprintHistory removes all history for a blueprint
func (db *PQDatabase) RemoveBlueprintHistory(blueprintID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `BLUEPRINT_HISTORY WHERE BLUEPRINT_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, blueprintID)
	return err
}
