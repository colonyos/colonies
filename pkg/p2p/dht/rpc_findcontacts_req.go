package dht

import "encoding/json"

type FindContactsReq struct {
	Header     RPCHeader `json:"header"`
	KademliaID string    `json:"kademliaid"`
	Count      int       `json:"count"`
}

func ConvertJSONToFindContactsReq(jsonStr string) (*FindContactsReq, error) {
	var req *FindContactsReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (req *FindContactsReq) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
