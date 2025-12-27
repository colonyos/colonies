package postgresql

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddLocation(location *core.Location) error {
	if location == nil {
		return errors.New("Location is nil")
	}

	existingLocation, err := db.GetLocationByName(location.ColonyName, location.Name)
	if err != nil {
		return err
	}

	if existingLocation != nil {
		return errors.New("Location with name <" + location.Name + "> already exists in Colony with name <" + location.ColonyName + ">")
	}

	sqlStatement := `INSERT INTO ` + db.dbPrefix + `LOCATIONS (NAME, LOCATION_ID, COLONY_NAME, DESCRIPTION, LONG, LAT) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.postgresql.Exec(sqlStatement, location.ColonyName+":"+location.Name, location.ID, location.ColonyName, location.Description, location.Long, location.Lat)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseLocations(rows *sql.Rows) ([]*core.Location, error) {
	var locations []*core.Location

	for rows.Next() {
		var name string
		var locationID string
		var colonyName string
		var description string
		var long float64
		var lat float64
		if err := rows.Scan(&name, &locationID, &colonyName, &description, &long, &lat); err != nil {
			return nil, err
		}

		s := strings.Split(name, ":")
		if len(s) != 2 {
			return nil, errors.New("Failed to parse Location name")
		}
		name = s[1]

		location := core.CreateLocation(locationID, name, colonyName, description, long, lat)
		locations = append(locations, location)
	}

	return locations, nil
}

func (db *PQDatabase) GetLocationsByColonyName(colonyName string) ([]*core.Location, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `LOCATIONS WHERE COLONY_NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseLocations(rows)
}

func (db *PQDatabase) GetLocationByID(locationID string) (*core.Location, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `LOCATIONS WHERE LOCATION_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, locationID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	locations, err := db.parseLocations(rows)
	if err != nil {
		return nil, err
	}

	if len(locations) == 0 {
		return nil, nil
	}

	return locations[0], nil
}

func (db *PQDatabase) GetLocationByName(colonyName string, name string) (*core.Location, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `LOCATIONS WHERE NAME=$1 AND COLONY_NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName+":"+name, colonyName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	locations, err := db.parseLocations(rows)
	if err != nil {
		return nil, err
	}

	if len(locations) == 0 {
		return nil, nil
	}

	return locations[0], nil
}

func (db *PQDatabase) RemoveLocationByID(locationID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `LOCATIONS WHERE LOCATION_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, locationID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveLocationByName(colonyName string, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `LOCATIONS WHERE NAME=$1 AND COLONY_NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName+":"+name, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveLocationsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `LOCATIONS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}
