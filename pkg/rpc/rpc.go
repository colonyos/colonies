package rpc

import (
	"encoding/json"
)

type RPC struct {
	Method string `json:"method"`
}

func DetermineMsgType(jsonString string) string {
	var msgMap map[string]interface{}
	json.Unmarshal([]byte(jsonString), &msgMap)
	rpcMap := msgMap["rpc"].(map[string]interface{})

	return rpcMap["method"].(string)
}
