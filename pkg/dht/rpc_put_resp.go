package dht

import "encoding/json"

const (
	PUT_STATUS_SUCCESS = 0
	PUT_STATUS_ERROR   = 1
)

type PutResp struct {
	Header RPCHeader `json:"header"`
	Status int       `json:"status"`
	Error  string    `json:"error"`
}

func ConvertJSONToPutResp(jsonStr string) (*PutResp, error) {
	var resp *PutResp
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (resp *PutResp) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
