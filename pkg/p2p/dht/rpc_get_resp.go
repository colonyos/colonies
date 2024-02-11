package dht

import "encoding/json"

const (
	GET_STATUS_SUCCESS = 0
	GET_STATUS_ERROR   = 1
)

type GetResp struct {
	Header RPCHeader `json:"header"`
	KVS    []KV      `json:"kvs"`
	Status int       `json:"status"`
	Error  string    `json:"error"`
}

func ConvertJSONToGetResp(jsonStr string) (*GetResp, error) {
	var resp *GetResp
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (resp *GetResp) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
