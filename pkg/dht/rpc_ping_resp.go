package dht

import "encoding/json"

type PingResponse struct {
	Header RPCHeader `json:"header"`
}

func ConvertJSONToPingResponse(jsonStr string) (*PingResponse, error) {
	var req *PingResponse
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *PingResponse) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
