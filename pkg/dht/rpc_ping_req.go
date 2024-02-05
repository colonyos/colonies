package dht

import "encoding/json"

type PingRequest struct {
	Header RPCHeader `json:"header"`
}

func ConvertJSONToPingRequest(jsonStr string) (*PingRequest, error) {
	var req *PingRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *PingRequest) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
