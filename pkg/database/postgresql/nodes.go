package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
)

func (db *PQDatabase) AddNode(node *core.Node) error {
	if node == nil {
		return errors.New("Node is nil")
	}

	labelsJSON, err := json.Marshal(node.Labels)
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `NODES (ID, NAME, COLONY_NAME, LOCATION, PLATFORM, ARCHITECTURE, CPU, MEMORY, GPU, CAPABILITIES, LABELS, EXECUTORS, STATE, LAST_SEEN, CREATED) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	emptyExecutors := []string{}
	_, err = db.postgresql.Exec(sqlStatement, node.ID, node.Name, node.ColonyName, node.Location, node.Platform, node.Architecture, node.CPU, node.Memory, node.GPU, pq.Array(node.Capabilities), labelsJSON, pq.Array(emptyExecutors), node.State, node.LastSeen, node.Created)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) GetNodeByID(nodeID string) (*core.Node, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `NODES WHERE ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, nodeID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var node *core.Node
	for rows.Next() {
		node, err = db.parseNode(rows)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func (db *PQDatabase) GetNodeByName(colonyName string, nodeName string) (*core.Node, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `NODES WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, nodeName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var node *core.Node
	for rows.Next() {
		node, err = db.parseNode(rows)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func (db *PQDatabase) GetNodes(colonyName string) ([]*core.Node, error) {
	var sqlStatement string
	var rows *sql.Rows
	var err error

	if colonyName == "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `NODES ORDER BY CREATED DESC`
		rows, err = db.postgresql.Query(sqlStatement)
	} else {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `NODES WHERE COLONY_NAME=$1 ORDER BY CREATED DESC`
		rows, err = db.postgresql.Query(sqlStatement, colonyName)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*core.Node, 0)
	for rows.Next() {
		node, err := db.parseNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (db *PQDatabase) GetNodesByLocation(colonyName string, location string) ([]*core.Node, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `NODES WHERE COLONY_NAME=$1 AND LOCATION=$2 ORDER BY CREATED DESC`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, location)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*core.Node, 0)
	for rows.Next() {
		node, err := db.parseNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (db *PQDatabase) UpdateNode(node *core.Node) error {
	if node == nil {
		return errors.New("Node is nil")
	}

	labelsJSON, err := json.Marshal(node.Labels)
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `NODES SET LOCATION=$1, PLATFORM=$2, ARCHITECTURE=$3, CPU=$4, MEMORY=$5, GPU=$6, CAPABILITIES=$7, LABELS=$8, EXECUTORS=$9, STATE=$10, LAST_SEEN=$11 WHERE ID=$12`
	emptyExecutors := []string{}
	_, err = db.postgresql.Exec(sqlStatement, node.Location, node.Platform, node.Architecture, node.CPU, node.Memory, node.GPU, pq.Array(node.Capabilities), labelsJSON, pq.Array(emptyExecutors), node.State, node.LastSeen, node.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveNodeByID(nodeID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `NODES WHERE ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, nodeID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveNodeByName(colonyName string, nodeName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `NODES WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, nodeName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveNodesByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `NODES WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountNodes(colonyName string) (int, error) {
	var count int
	var sqlStatement string

	if colonyName == "" {
		sqlStatement = `SELECT COUNT(*) FROM ` + db.dbPrefix + `NODES`
		err := db.postgresql.QueryRow(sqlStatement).Scan(&count)
		if err != nil {
			return -1, err
		}
	} else {
		sqlStatement = `SELECT COUNT(*) FROM ` + db.dbPrefix + `NODES WHERE COLONY_NAME=$1`
		err := db.postgresql.QueryRow(sqlStatement, colonyName).Scan(&count)
		if err != nil {
			return -1, err
		}
	}

	return count, nil
}

func (db *PQDatabase) parseNode(rows *sql.Rows) (*core.Node, error) {
	var id string
	var name string
	var colonyName string
	var location string
	var platform string
	var architecture string
	var cpu int
	var memory int64
	var gpu int
	var capabilities []string
	var labelsJSON string
	var executors []string
	var state string
	var lastSeen sql.NullTime
	var created sql.NullTime

	if err := rows.Scan(&id, &name, &colonyName, &location, &platform, &architecture, &cpu, &memory, &gpu, pq.Array(&capabilities), &labelsJSON, pq.Array(&executors), &state, &lastSeen, &created); err != nil {
		return nil, err
	}

	// Parse labels from JSON
	labels := make(map[string]string)
	if labelsJSON != "" {
		err := json.Unmarshal([]byte(labelsJSON), &labels)
		if err != nil {
			return nil, err
		}
	}

	node := &core.Node{
		ID:           id,
		Name:         name,
		ColonyName:   colonyName,
		Location:     location,
		Platform:     platform,
		Architecture: architecture,
		CPU:          cpu,
		Memory:       memory,
		GPU:          gpu,
		Capabilities: capabilities,
		Labels:       labels,
		State:        state,
	}

	if lastSeen.Valid {
		node.LastSeen = lastSeen.Time
	}

	if created.Valid {
		node.Created = created.Time
	}

	return node, nil
}
