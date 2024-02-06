package dht

import "encoding/json"

type PingResp struct {
	Header RPCHeader `json:"header"`
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
