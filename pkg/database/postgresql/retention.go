package postgresql

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) calcTimestamp(retentionPeriod int64) (time.Time, time.Time) {
	now := time.Now()
	period := time.Duration(retentionPeriod) * time.Second
	timestamp := now.Add(-period)

	return now, timestamp
}

// retentionPeriod in seconds
func (db *PQDatabase) ApplyRetentionPolicy(retentionPeriod int64) error {
	_, timestamp := db.calcTimestamp(retentionPeriod)

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `ATTRIBUTES WHERE ADDED<$1 AND STATE=$2`
	_, err := db.postgresql.Exec(sqlStatement, timestamp, core.SUCCESS)
	if err != nil {
		return err
	}

	sqlStatement = `DELETE FROM ` + db.dbPrefix + `LOGS WHERE ADDED<$1`
	_, err = db.postgresql.Exec(sqlStatement, timestamp)
	if err != nil {
		return err
	}

	sqlStatement = `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE SUBMISSION_TIME<$1 AND STATE=$2`
	_, err = db.postgresql.Exec(sqlStatement, timestamp, core.SUCCESS)
	if err != nil {
		return err
	}

	sqlStatement = `DELETE FROM ` + db.dbPrefix + `PROCESSGRAPHS WHERE SUBMISSION_TIME<$1 AND STATE=$2`
	_, err = db.postgresql.Exec(sqlStatement, timestamp, core.SUCCESS)
	if err != nil {
		return err
	}

	return nil
}
