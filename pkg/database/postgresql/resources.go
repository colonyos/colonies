package postgresql

import (
	"database/sql"
	"errors"

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

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `RESOURCEDEFINITIONS (ID, COLONY_NAME, NAME, API_GROUP, VERSION, KIND, DATA) VALUES ($1, $2, $3, $4, $5, $6, $7)`
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS WHERE ID=$1`
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResourceDefinitions(rows)
}

func (db *PQDatabase) GetResourceDefinitionsByNamespace(namespace string) ([]*core.ResourceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResourceDefinitions(rows)
}

func (db *PQDatabase) GetResourceDefinitionsByGroup(group string) ([]*core.ResourceDefinition, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS WHERE API_GROUP=$1 ORDER BY NAME`
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

	sqlStatement := `UPDATE ` + db.dbPrefix + `RESOURCEDEFINITIONS SET NAME=$1, API_GROUP=$2, VERSION=$3, KIND=$4, DATA=$5 WHERE ID=$6`
	_, err = db.postgresql.Exec(sqlStatement, rd.Metadata.Name, rd.Spec.Group, rd.Spec.Version, rd.Spec.Names.Kind, rdJSON, rd.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourceDefinitionByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourceDefinitionByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountResourceDefinitions() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `RESOURCEDEFINITIONS`
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

// Resource methods

func (db *PQDatabase) AddResource(resource *core.Resource) error {
	if resource == nil {
		return errors.New("Resource is nil")
	}

	existingResource, err := db.GetResourceByName(resource.Metadata.Namespace, resource.Metadata.Name)
	if err != nil {
		return err
	}

	if existingResource != nil {
		return errors.New("Resource with name <" + resource.Metadata.Name + "> in namespace <" + resource.Metadata.Namespace + "> already exists")
	}

	resourceJSON, err := resource.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `RESOURCES (ID, COLONY_NAME, NAME, KIND, DATA) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.postgresql.Exec(sqlStatement, resource.ID, resource.Metadata.Namespace, resource.Metadata.Name, resource.Kind, resourceJSON)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseResources(rows *sql.Rows) ([]*core.Resource, error) {
	var resources []*core.Resource

	for rows.Next() {
		var id string
		var colonyName string
		var name string
		var kind string
		var data string

		if err := rows.Scan(&id, &colonyName, &name, &kind, &data); err != nil {
			return nil, err
		}

		resource, err := core.ConvertJSONToResource(data)
		if err != nil {
			return nil, err
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (db *PQDatabase) GetResourceByID(id string) (*core.Resource, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `RESOURCES WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resources, err := db.parseResources(rows)
	if err != nil {
		return nil, err
	}

	if len(resources) == 0 {
		return nil, nil
	}

	return resources[0], nil
}

func (db *PQDatabase) GetResourceByName(namespace, name string) (*core.Resource, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `RESOURCES WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, namespace, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resources, err := db.parseResources(rows)
	if err != nil {
		return nil, err
	}

	if len(resources) == 0 {
		return nil, nil
	}

	return resources[0], nil
}

func (db *PQDatabase) GetResources() ([]*core.Resource, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `RESOURCES ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) GetResourcesByNamespace(namespace string) ([]*core.Resource, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `RESOURCES WHERE COLONY_NAME=$1 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) GetResourcesByKind(kind string) ([]*core.Resource, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `RESOURCES WHERE KIND=$1 ORDER BY COLONY_NAME, NAME`
	rows, err := db.postgresql.Query(sqlStatement, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) GetResourcesByNamespaceAndKind(namespace, kind string) ([]*core.Resource, error) {
	sqlStatement := `SELECT ID, COLONY_NAME, NAME, KIND, DATA FROM ` + db.dbPrefix + `RESOURCES WHERE COLONY_NAME=$1 AND KIND=$2 ORDER BY NAME`
	rows, err := db.postgresql.Query(sqlStatement, namespace, kind)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseResources(rows)
}

func (db *PQDatabase) UpdateResource(resource *core.Resource) error {
	if resource == nil {
		return errors.New("Resource is nil")
	}

	resourceJSON, err := resource.ToJSON()
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `RESOURCES SET COLONY_NAME=$1, NAME=$2, KIND=$3, DATA=$4 WHERE ID=$5`
	_, err = db.postgresql.Exec(sqlStatement, resource.Metadata.Namespace, resource.Metadata.Name, resource.Kind, resourceJSON, resource.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) UpdateResourceStatus(id string, status map[string]interface{}) error {
	// Get the current resource
	resource, err := db.GetResourceByID(id)
	if err != nil {
		return err
	}
	if resource == nil {
		return errors.New("Resource not found")
	}

	// Update status
	resource.Status = status

	// Save back
	return db.UpdateResource(resource)
}

func (db *PQDatabase) RemoveResourceByID(id string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `RESOURCES WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourceByName(namespace, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `RESOURCES WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, namespace, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveResourcesByNamespace(namespace string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `RESOURCES WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountResources() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `RESOURCES`
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
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `RESOURCES WHERE COLONY_NAME=$1`
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
