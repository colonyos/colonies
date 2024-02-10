package dht

import "encoding/json"

type GetReq struct {
	Header RPCHeader `json:"header"`
	Key    string    `json:"key"`
}

func ConvertJSONToGetReq(jsonStr string) (*GetReq, error) {
	var req *GetReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *GetReq) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
