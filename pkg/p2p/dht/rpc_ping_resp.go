package dht

import "encoding/json"

const (
	PING_STATUS_SUCCESS = 0
	PING_STATUS_ERROR   = 1
)

type PingResp struct {
	Header RPCHeader `json:"header"`
	Status int       `json:"status"`
	Error  string    `json:"error"`
}

func ConvertJSONToPingResp(jsonStr string) (*PingResp, error) {
	var req *PingResp
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *PingResp) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
