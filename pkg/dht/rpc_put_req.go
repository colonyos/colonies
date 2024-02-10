package dht

import "encoding/json"

type PutReq struct {
	Header RPCHeader `json:"header"`
	KV     KV        `json:"kv"`
}

func ConvertJSONToPutReq(jsonStr string) (*PutReq, error) {
	var req *PutReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *PutReq) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
