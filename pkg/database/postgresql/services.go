package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

// ServiceDefinition methods

func (db *PQDatabase) AddServiceDefinition(sd *core.ServiceDefinition) error {
	if sd == nil {
		return errors.New("ServiceDefinition is nil")
	}

	sdJSON, err := sd.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `SERVICEDEFINITIONS (ID, COLONY_NAME, NAME, API_GROUP, VERSION, KIND, DATA) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.postgresql.Exec(sqlStatement, sd.ID, sd.Metadata.Namespace, sd.Metadata.Name, sd.Spec.Group, sd.Spec.Version, sd.Spec.Names.Kind, sdJSON)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseServiceDefinitions(rows *sql.Rows) ([]*core.ServiceDefinition, error) {
	var sds []*core.ServiceDefinition

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

		sd, err := core.ConvertJSONToServiceDefinition(data)
		if err != nil {
			return nil, err
		}

		sds = append(sds, sd)
	}

	return sds, nil
}

func (db *PQDatabase) GetServiceDefinitionByID(id string) (*core.ServiceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sds, err := db.parseServiceDefinitions(rows)
	if err != nil {
		return nil, err
	}

	if len(sds) == 0 {
		return nil, nil
	}

	return sds[0], nil
}

func (db *PQDatabase) GetServiceDefinitionByName(namespace, name string) (*core.ServiceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, namespace, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sds, err := db.parseServiceDefinitions(rows)
	if err != nil {
		return nil, err
	}

	if len(sds) == 0 {
		return nil, nil
	}

	return sds[0], nil
}

func (db *PQDatabase) GetServiceDefinitions() ([]*core.ServiceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseServiceDefinitions(rows)
}

func (db *PQDatabase) GetServiceDefinitionsByNamespace(namespace string) ([]*core.ServiceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseServiceDefinitions(rows)
}

func (db *PQDatabase) GetServiceDefinitionsByGroup(group string) ([]*core.ServiceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE API_GROUP=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, group)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseServiceDefinitions(rows)
}

func (db *PQDatabase) UpdateServiceDefinition(sd *core.ServiceDefinition) error {
	if sd == nil {
		return errors.New("ServiceDefinition is nil")
	}

	sdJSON, err := sd.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `SERVICEDEFINITIONS SET NAME=$1, API_GROUP=$2, VERSION=$3, KIND=$4, DATA=$5 WHERE ID=$6`
	_, err = db.postgresql.Exec(sqlStatement, sd.Metadata.Name, sd.Spec.Group, sd.Spec.Version, sd.Spec.Names.Kind, sdJSON, sd.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveServiceDefinitionByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveServiceDefinitionByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICEDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountServiceDefinitions() (int, error) {
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

func (db *PQDatabase) AddService(service *core.Service) error {
	if service == nil {
		return errors.New("Service is nil")
	}

	existingService, err := db.GetServiceByName(service.Metadata.Namespace, service.Metadata.Name)
	if err != nil {
		return err
	}

	if existingService != nil {
		return errors.New("Service with name <" + service.Metadata.Name + "> in namespace <" + service.Metadata.Namespace + "> already exists")
	}

	serviceJSON, err := service.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `SERVICES (ID, COLONY_NAME, NAME, KIND, DATA) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.postgresql.Exec(sqlStatement, service.ID, service.Metadata.Namespace, service.Metadata.Name, service.Kind, serviceJSON)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseServices(rows *sql.Rows) ([]*core.Service, error) {
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

		service, err := core.ConvertJSONToService(data)
		if err != nil {
			return nil, err
		}

		// Set the ID from the database (not stored in the JSON DATA column)
		service.ID = id

		services = append(services, service)
	}

	return services, nil
}

func (db *PQDatabase) GetServiceByID(id string) (*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	services, err := db.parseServices(rows)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	return services[0], nil
}

func (db *PQDatabase) GetServiceByName(namespace, name string) (*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, namespace, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	services, err := db.parseServices(rows)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	return services[0], nil
}

func (db *PQDatabase) GetServices() ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseServices(rows)
}

func (db *PQDatabase) GetServicesByNamespace(namespace string) ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseServices(rows)
}

func (db *PQDatabase) GetServicesByKind(kind string) ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE KIND=$1 ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseServices(rows)
}

func (db *PQDatabase) GetServicesByNamespaceAndKind(namespace, kind string) ([]*core.Service, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 AND KIND=$2 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseServices(rows)
}

func (db *PQDatabase) UpdateService(service *core.Service) error {
	if service == nil {
		return errors.New("Service is nil")
	}

	serviceJSON, err := service.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `SERVICES SET COLONY_NAME=$1, NAME=$2, KIND=$3, DATA=$4 WHERE ID=$5`
	_, err = db.postgresql.Exec(sqlStatement, service.Metadata.Namespace, service.Metadata.Name, service.Kind, serviceJSON, service.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) UpdateServiceStatus(id string, status map[string]interface{}) error {
	// Get the current service data to preserve spec and metadata
	sqlStatement := `SELECT DATA FROM ` + db.dbPrefix + `SERVICES WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return errors.New("Service not found")
	}

	var dataStr string
	if err := rows.Scan(&dataStr); err != nil {
		return err
	}

	// Parse the JSON data
	var serviceData map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &serviceData); err != nil {
		return fmt.Errorf("failed to unmarshal service data: %w", err)
	}

	// Update only the status field
	serviceData["status"] = status

	// Convert back to JSON
	updatedJSON, err := json.Marshal(serviceData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated service data: %w", err)
	}

	// Update the database
	// Note: This still has a potential race condition, but it's much smaller window
	// than the original read-modify-write pattern
	updateStatement := `UPDATE ` + db.dbPrefix + `SERVICES SET DATA=$1 WHERE ID=$2`
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
		return errors.New("Service not found")
	}

	return nil
}

func (db *PQDatabase) RemoveServiceByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICES WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveServiceByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveServicesByNamespace(namespace string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICES WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountServices() (int, error) {
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

func (db *PQDatabase) CountServicesByNamespace(namespace string) (int, error) {
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

// AddServiceHistory adds a new ServiceHistory entry
func (db *PQDatabase) AddServiceHistory(history *core.ServiceHistory) error {
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
		history.ServiceID,
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

// GetServiceHistory retrieves history for a service (most recent first)
func (db *PQDatabase) GetServiceHistory(serviceID string, limit int) ([]*core.ServiceHistory, error) {
	sqlStatement := `SELECT ID, SERVICE_ID, KIND, NAMESPACE, NAME, GENERATION,
		SPEC, STATUS, TIMESTAMP, CHANGED_BY, CHANGE_TYPE
		FROM ` + db.dbPrefix + `SERVICE_HISTORY
		WHERE SERVICE_ID=$1
		ORDER BY TIMESTAMP DESC`

	if limit > 0 {
		sqlStatement += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.postgresql.Query(sqlStatement, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*core.ServiceHistory
	for rows.Next() {
		var history core.ServiceHistory
		var specJSON, statusJSON []byte

		err := rows.Scan(
			&history.ID,
			&history.ServiceID,
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

// GetServiceHistoryByGeneration retrieves a specific generation of a service
func (db *PQDatabase) GetServiceHistoryByGeneration(serviceID string, generation int64) (*core.ServiceHistory, error) {
	sqlStatement := `SELECT ID, SERVICE_ID, KIND, NAMESPACE, NAME, GENERATION,
		SPEC, STATUS, TIMESTAMP, CHANGED_BY, CHANGE_TYPE
		FROM ` + db.dbPrefix + `SERVICE_HISTORY
		WHERE SERVICE_ID=$1 AND GENERATION=$2`

	rows, err := db.postgresql.Query(sqlStatement, serviceID, generation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history core.ServiceHistory
	var specJSON, statusJSON []byte

	if rows.Next() {
		err := rows.Scan(
			&history.ID,
			&history.ServiceID,
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

// RemoveServiceHistory removes all history for a service
func (db *PQDatabase) RemoveServiceHistory(serviceID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SERVICE_HISTORY WHERE SERVICE_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, serviceID)
	return err
}
