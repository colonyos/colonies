package dht

import "encoding/json"

type FindContactsRequest struct {
	RPCHeader
}

func ConvertJSONToFindContactsRequest(jsonStr string) (*FindContactsRequest, error) {
	var req *FindContactsRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *FindContactsRequest) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
