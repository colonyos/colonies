package postgresql

import (
	"database/sql"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddFile(file *core.File) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `FILES (FILE_ID, COLONY_ID, PREFIX, NAME, SIZE, SEQNR, CHECKSUM, CHECKSUM_ALG, ADDED, PROTOCOL, S3_SERVER, S3_PORT, S3_TLS, S3_ACCESSKEY, S3_SECRETKEY, S3_REGION, S3_ENCKEY, S3_ENCALG, S3_OBJ, S3_BUCKET) VALUES ($1, $2, $3, $4, $5, nextval('` + db.dbPrefix + `FILE_SEQ'), $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`
	_, err := db.postgresql.Exec(sqlStatement, file.ID, file.ColonyID, file.Prefix, file.Name, file.Size, file.Checksum, file.ChecksumAlg, time.Now(), file.Reference.Protocol, file.Reference.S3Object.Server, file.Reference.S3Object.Port, file.Reference.S3Object.TLS, file.Reference.S3Object.AccessKey, file.Reference.S3Object.SecretKey, file.Reference.S3Object.Region, file.Reference.S3Object.EncryptionKey, file.Reference.S3Object.EncryptionAlg, file.Reference.S3Object.Object, file.Reference.S3Object.Bucket)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseFiles(rows *sql.Rows) ([]*core.File, error) {
	var files []*core.File

	for rows.Next() {
		var fileID string
		var colonyID string
		var prefix string
		var name string
		var size int64
		var seqnr int64
		var checksum string
		var checksumAlg string
		var added time.Time
		var protocol string
		var s3Server string
		var s3Port int
		var s3TLS bool
		var s3AccessKey string
		var s3SecretKey string
		var s3Region string
		var s3EncryptionKey string
		var s3EncryptionAlg string
		var s3Object string
		var s3Bucket string

		if err := rows.Scan(&fileID, &colonyID, &prefix, &name, &size, &seqnr, &checksum, &checksumAlg, &added, &protocol, &s3Server, &s3Port, &s3TLS, &s3AccessKey, &s3SecretKey, &s3Region, &s3EncryptionKey, &s3EncryptionAlg, &s3Object, &s3Bucket); err != nil {
			return nil, err
		}

		s3ObjectStruct := core.S3Object{
			Server:        s3Server,
			Port:          s3Port,
			TLS:           s3TLS,
			AccessKey:     s3AccessKey,
			SecretKey:     s3SecretKey,
			Region:        s3Region,
			EncryptionKey: s3EncryptionKey,
			EncryptionAlg: s3EncryptionAlg,
			Object:        s3Object,
			Bucket:        s3Bucket,
		}
		ref := core.Reference{Protocol: "s3", S3Object: s3ObjectStruct}
		file := core.File{
			ID:             fileID,
			ColonyID:       colonyID,
			Prefix:         prefix,
			Name:           name,
			Size:           size,
			SequenceNumber: seqnr,
			Checksum:       checksum,
			ChecksumAlg:    checksumAlg,
			Reference:      ref,
			Added:          added}

		files = append(files, &file)
	}

	return files, nil
}

func (db *PQDatabase) parsePrefix(rows *sql.Rows) ([]string, error) {
	var prefixStr []string

	for rows.Next() {
		var prefix string

		if err := rows.Scan(&prefix); err != nil {
			return nil, err
		}

		prefixStr = append(prefixStr, prefix)

	}

	return prefixStr, nil
}

func (db *PQDatabase) GetFileByID(colonyID string, fileID string) (*core.File, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1 AND FILE_ID=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, fileID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files, err := db.parseFiles(rows)
	if err != nil {
		return nil, err
	}

	if len(files) == 1 {
		return files[0], nil
	}

	return nil, nil
}

func (db *PQDatabase) GetLatestFileByName(colonyID string, prefix string, name string) ([]*core.File, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1 AND NAME=$2 AND PREFIX=$3 ORDER BY SEQNR DESC LIMIT 1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, name, prefix)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files, err := db.parseFiles(rows)
	if err != nil {
		return nil, err
	}

	if len(files) == 1 {
		return files, nil
	}

	return nil, nil
}

func (db *PQDatabase) GetFileByName(colonyID string, prefix string, name string) ([]*core.File, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1 AND PREFIX=$2 AND NAME=$3 ORDER BY SEQNR DESC`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, prefix, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files, err := db.parseFiles(rows)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (db *PQDatabase) GetFileNamesByPrefix(colonyID string, prefix string) ([]string, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1 AND PREFIX=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, prefix)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files, err := db.parseFiles(rows)
	if err != nil {
		return nil, err
	}

	// Just to filter out duplicates as there can be many versions of the same file
	filemap := make(map[string]string)
	for _, file := range files {
		filemap[file.Name] = file.Name
	}

	var filenames []string
	for _, filename := range filemap {
		filenames = append(filenames, filename)
	}

	return filenames, nil
}

func (db *PQDatabase) DeleteFileByID(colonyID string, fileID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1 AND FILE_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, fileID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteFileByName(colonyID string, prefix string, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1 AND PREFIX=$2 AND NAME=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, prefix, name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteFilesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) GetFilePrefixes(colonyID string) ([]string, error) {
	sqlStatement := `SELECT DISTINCT (PREFIX) FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	prefixesStr, err := db.parsePrefix(rows)
	if err != nil {
		return nil, err
	}

	return prefixesStr, nil
}

func (db *PQDatabase) CountFiles(colonyID string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `FILES WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	rows.Next()
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}
