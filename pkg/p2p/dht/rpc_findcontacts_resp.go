package dht

import "encoding/json"

const (
	FIND_CONTACTS_STATUS_SUCCESS = 0
	FIND_CONTACTS_STATUS_ERROR   = 1
)

type FindContactsResp struct {
	Header   RPCHeader `json:"header"`
	Contacts []Contact `json:"contacts"`
	Status   int       `json:"status"`
	Error    string    `json:"error"`
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
