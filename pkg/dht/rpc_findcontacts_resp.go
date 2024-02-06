package dht

import "encoding/json"

type FindContactsResp struct {
	Header   RPCHeader `json:"header"`
	Contacts []Contact `json:"contacts"`
}

func ConvertJSONToFindContactsResp(jsonStr string) (*FindContactsResp, error) {
	var resp *FindContactsResp
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (resp *FindContactsResp) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
