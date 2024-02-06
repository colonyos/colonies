package dht

import "encoding/json"

type PingReq struct {
	Header RPCHeader `json:"header"`
}

func ConvertJSONToPingReq(jsonStr string) (*PingReq, error) {
	var req *PingReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *PingReq) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
