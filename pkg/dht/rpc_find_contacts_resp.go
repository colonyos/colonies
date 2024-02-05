package dht

import "encoding/json"

type FindContactsResponse struct {
	RPCHeader
}

func ConvertJSONToFindContactsResponse(jsonStr string) (*FindContactsResponse, error) {
	var req *FindContactsResponse
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *FindContactsResponse) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
