package core

import (
	"encoding/json"
	"time"
)

type S3Object struct {
	Server        string `json:"server"`
	Port          int    `json:"port"`
	TLS           bool   `json:"tls"`
	AccessKey     string `json:"accesskey"`
	SecretKey     string `json:"secretkey"`
	Region        string `json:"region"`
	EncryptionKey string `json:"encryptionkey"`
	EncryptionAlg string `json:"encryptionalg"`
	Object        string `json:"object"`
	Bucket        string `json:"bucket"`
}

type Reference struct {
	Protocol string   `json:"protocol"`
	S3Object S3Object `json:"s3object"`
}

type File struct {
	ID             string    `json:"fileid"`
	ColonyID       string    `json:"colonyid"`
	Label          string    `json:"label"`
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	SequenceNumber int64     `json:"sequencenr"`
	Checksum       string    `json:"checksum"`
	ChecksumAlg    string    `json:"checksumalg"`
	Reference      Reference `json:"ref"`
	Added          time.Time `json:"added"`
}

func ConvertJSONToFile(jsonString string) (*File, error) {
	var file *File
	err := json.Unmarshal([]byte(jsonString), &file)
	if err != nil {
		return &File{}, err
	}

	return file, nil
}

func (file *File) Equals(file2 *File) bool {
	same := true

	if file.Reference.S3Object.Server != file2.Reference.S3Object.Server {
		same = false
	}
	if file.Reference.S3Object.Port != file2.Reference.S3Object.Port {
		same = false
	}
	if file.Reference.S3Object.TLS != file2.Reference.S3Object.TLS {
		same = false
	}
	if file.Reference.S3Object.AccessKey != file2.Reference.S3Object.AccessKey {
		same = false
	}
	if file.Reference.S3Object.SecretKey != file2.Reference.S3Object.SecretKey {
		same = false
	}
	if file.Reference.S3Object.Region != file2.Reference.S3Object.Region {
		same = false
	}
	if file.Reference.S3Object.EncryptionKey != file2.Reference.S3Object.EncryptionKey {
		same = false
	}
	if file.Reference.S3Object.EncryptionAlg != file2.Reference.S3Object.EncryptionAlg {
		same = false
	}
	if file.Reference.S3Object.Object != file2.Reference.S3Object.Object {
		same = false
	}
	if file.Reference.S3Object.Bucket != file2.Reference.S3Object.Bucket {
		same = false
	}

	if file.Reference.Protocol != file2.Reference.Protocol {
		same = false
	}

	if file.ID != file2.ID {
		same = false
	}
	if file.ColonyID != file2.ColonyID {
		same = false
	}
	if file.Label != file2.Label {
		same = false
	}
	if file.Name != file2.Name {
		same = false
	}
	if file.Size != file2.Size {
		same = false
	}
	if file.SequenceNumber != file2.SequenceNumber {
		same = false
	}
	if file.Checksum != file2.Checksum {
		same = false
	}
	if file.ChecksumAlg != file2.ChecksumAlg {
		same = false
	}
	if file.Added.Unix() != file2.Added.Unix() {
		same = false
	}

	return same
}

func ConvertFileArrayToJSON(files []*File) (string, error) {
	jsonBytes, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToFileArray(jsonString string) ([]*File, error) {
	var files []*File
	err := json.Unmarshal([]byte(jsonString), &files)
	if err != nil {
		return files, err
	}

	return files, nil
}

func IsFileArraysEqual(files1 []*File, files2 []*File) bool {
	if files1 == nil || files2 == nil {
		return false
	}

	if len(files1) != len(files2) {
		return false
	}

	counter := 0
	for i := range files1 {
		if files1[i].Equals(files2[i]) {
			counter++
		}
	}

	if counter == len(files1) {
		return true
	}

	return false
}

func (file *File) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(file)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
