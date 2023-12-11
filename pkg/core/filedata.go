package core

import "encoding/json"

type FileData struct {
	Name       string `json:"name"`
	Checksum   string `json:"checksum"`
	Size       int64  `json:"size"`
	S3Filename string `json:"s3filename"`
}

func ConvertJSONToFileData(jsonString string) (*FileData, error) {
	var fileData *FileData
	err := json.Unmarshal([]byte(jsonString), &fileData)
	if err != nil {
		return &FileData{}, err
	}

	return fileData, nil
}

func ConvertFileDataArrayToJSON(fileDataArr []*FileData) (string, error) {
	jsonBytes, err := json.Marshal(fileDataArr)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToFileDataArray(jsonString string) ([]*FileData, error) {
	var fileData []*FileData
	err := json.Unmarshal([]byte(jsonString), &fileData)
	if err != nil {
		return fileData, err
	}

	return fileData, nil
}

func IsFileDataArraysEqual(fileDataArr1 []*FileData, fileDataArr2 []*FileData) bool {
	if fileDataArr1 == nil || fileDataArr2 == nil {
		return false
	}

	if len(fileDataArr1) != len(fileDataArr2) {
		return false
	}

	counter := 0
	for i := range fileDataArr1 {
		if fileDataArr1[i].Equals(fileDataArr2[i]) {
			counter++
		}
	}

	if counter == len(fileDataArr1) {
		return true
	}

	return false
}

func (fileData *FileData) Equals(fileData2 *FileData) bool {
	if fileData.Name != fileData2.Name {
		return false
	}

	if fileData.Checksum != fileData2.Checksum {
		return false
	}

	if fileData.Size != fileData2.Size {
		return false
	}

	if fileData.S3Filename != fileData2.S3Filename {
		return false
	}

	return true
}

func (fileData *FileData) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(fileData)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
