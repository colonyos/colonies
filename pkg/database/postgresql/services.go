package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

// ResourceDefinition methods

func (db *PQDatabase) AddResourceDefinition(rd *core.ResourceDefinition) error {
	if rd == nil {
		return errors.New("ResourceDefinition is nil")
	}

	rdJSON, err := rd.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `SERVICEDEFINITIONS (ID, COLONY_NAME, NAME, API_GROUP, VERSION, KIND, DATA) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.postgresql.Exec(sqlStatement, rd.ID, rd.Metadata.Namespace, rd.Metadata.Name, rd.Spec.Group, rd.Spec.Version, rd.Spec.Names.Kind, rdJSON)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseResourceDefinitions(rows *sql.Rows) ([]*core.ResourceDefinition, error) {
	var rds []*core.ResourceDefinition

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

		rd, err := core.ConvertJSONToResourceDefinition(data)
		if err != nil {
			return nil, err
		}

		rds = append(rds, rd)
	}

	return rds, nil
}

func (db *PQDatabase) GetResourceDefinitionByID(id string) (*core.ResourceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rds, err := db.parseResourceDefinitions(rows)
	if err != nil {
		return nil, err
	}

	if len(rds) == 0 {
		return nil, nil
	}

	return rds[0], nil
}

func (db *PQDatabase) GetResourceDefinitionByName(namespace, name string) (*core.ResourceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, namespace, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rds, err := db.parseResourceDefinitions(rows)
	if err != nil {
		return nil, err
	}

	if len(rds) == 0 {
		return nil, nil
	}

	return rds[0], nil
}

func (db *PQDatabase) GetResourceDefinitions() ([]*core.ResourceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResourceDefinitions(rows)
}

func (db *PQDatabase) GetResourceDefinitionsByNamespace(namespace string) ([]*core.ResourceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResourceDefinitions(rows)
}

func (db *PQDatabase) GetResourceDefinitionsByGroup(group string) ([]*core.ResourceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE API_GROUP=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, group)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResourceDefinitions(rows)
}

func (db *PQDatabase) UpdateResourceDefinition(rd *core.ResourceDefinition) error {
	if rd == nil {
		return errors.New("ResourceDefinition is nil")
	}

	rdJSON, err := rd.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `SERVICEDEFINITIONS SET NAME=$1, API_GROUP=$2, VERSION=$3, KIND=$4, DATA=$5 WHERE ID=$6`
	_, err = db.postgresql.Exec(sqlStatement, rd.Metadata.Name, rd.Spec.Group, rd.Spec.Version, rd.Spec.Names.Kind, rdJSON, rd.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourceDefinitionByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourceDefinitionByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountResourceDefinitions() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `SERVICEDEFINITIONS`
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

// Service methods

func (db *PQDatabase) AddResource(service *core.Service) error {
	if service == nil {
		return errors.New("Service is nil")
	}

	existingResource, err := db.GetResourceByName(service.Metadata.Namespace, service.Metadata.Name)
	if err != nil {
		return err
	}

	if existingResource != nil {
		return errors.New("Service with name <" + service.Metadata.Name + "> in namespace <" + service.Metadata.Namespace + "> already exists")
	}

	resourceJSON, err := service.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `SERVICES (ID, COLONY_NAME, NAME, KIND, DATA) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.postgresql.Exec(sqlStatement, service.ID, service.Metadata.Namespace, service.Metadata.Name, service.Kind, resourceJSON)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseResources(rows *sql.Rows) ([]*core.Service, error) {
	var services []*core.Service

	for rows.Next() {
		var id string
		var colonyName string
		var name string
		var kind string
		var data string

		if err := rows.Scan(&id, &colonyName, &name, &kind, &data); err != nil {
			return nil, err
		}

		service, err := core.ConvertJSONToResource(data)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}

func (db *PQDatabase) GetResourceByID(id string) (*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	services, err := db.parseResources(rows)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	return services[0], nil
}

func (db *PQDatabase) GetResourceByName(namespace, name string) (*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, namespace, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	services, err := db.parseResources(rows)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	return services[0], nil
}

func (db *PQDatabase) GetResources() ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) GetResourcesByNamespace(namespace string) ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) GetResourcesByKind(kind string) ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE KIND=$1 ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) GetResourcesByNamespaceAndKind(namespace, kind string) ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 AND KIND=$2 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) UpdateResource(service *core.Service) error {
	if service == nil {
		return errors.New("Service is nil")
	}

	resourceJSON, err := service.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `SERVICES SET COLONY_NAME=$1, NAME=$2, KIND=$3, DATA=$4 WHERE ID=$5`
	_, err = db.postgresql.Exec(sqlStatement, service.Metadata.Namespace, service.Metadata.Name, service.Kind, resourceJSON, service.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) UpdateResourceStatus(id string, status map[string]interface{}) error {
	// Get the current service
	service, err := db.GetResourceByID(id)
	if err != nil {
		return err
	}
	if service == nil {
		return errors.New("Service not found")
	}

	// Update status
	service.Status = status

	// Save back
	return db.UpdateResource(service)
}

func (db *PQDatabase) RemoveResourceByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICES WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourceByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourcesByNamespace(namespace string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountResources() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `SERVICES`
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

func (db *PQDatabase) CountResourcesByNamespace(namespace string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1`
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

// AddResourceHistory adds a new ResourceHistory entry
func (db *PQDatabase) AddResourceHistory(history *core.ResourceHistory) error {
	specJSON, err := json.Marshal(history.Spec)
	if err != nil {
		return err
	}

	statusJSON, err := json.Marshal(history.Status)
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `SERVICE_HISTORY (
		ID, SERVICE_ID, KIND, NAMESPACE, NAME, GENERATION, SPEC, STATUS,
		TIMESTAMP, CHANGED_BY, CHANGE_TYPE)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.postgresql.Exec(
		sqlStatement,
		history.ID,
		history.ResourceID,
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

// GetResourceHistory retrieves history for a service (most recent first)
func (db *PQDatabase) GetResourceHistory(resourceID string, limit int) ([]*core.ResourceHistory, error) {
	sqlStatement := `SELECT ID, SERVICE_ID, KIND, NAMESPACE, NAME, GENERATION,
		SPEC, STATUS, TIMESTAMP, CHANGED_BY, CHANGE_TYPE
		FROM ` + db.dbPrefix + `SERVICE_HISTORY
		WHERE SERVICE_ID=$1
		ORDER BY TIMESTAMP DESC`

	if limit > 0 {
		sqlStatement += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.postgresql.Query(sqlStatement, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*core.ResourceHistory
	for rows.Next() {
		var history core.ResourceHistory
		var specJSON, statusJSON []byte

		err := rows.Scan(
			&history.ID,
			&history.ResourceID,
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

// GetResourceHistoryByGeneration retrieves a specific generation of a service
func (db *PQDatabase) GetResourceHistoryByGeneration(resourceID string, generation int64) (*core.ResourceHistory, error) {
	sqlStatement := `SELECT ID, SERVICE_ID, KIND, NAMESPACE, NAME, GENERATION,
		SPEC, STATUS, TIMESTAMP, CHANGED_BY, CHANGE_TYPE
		FROM ` + db.dbPrefix + `SERVICE_HISTORY
		WHERE SERVICE_ID=$1 AND GENERATION=$2`

	rows, err := db.postgresql.Query(sqlStatement, resourceID, generation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history core.ResourceHistory
	var specJSON, statusJSON []byte

	if rows.Next() {
		err := rows.Scan(
			&history.ID,
			&history.ResourceID,
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

// RemoveResourceHistory removes all history for a service
func (db *PQDatabase) RemoveResourceHistory(resourceID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICE_HISTORY WHERE SERVICE_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, resourceID)
	return err
}
