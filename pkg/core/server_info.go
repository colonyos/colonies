package core

import (
	"encoding/json"
)

type BackendInfo struct {
	Type     string `json:"type"`
	Port     int    `json:"port"`
	Host     string `json:"host,omitempty"`
	TLS      bool   `json:"tls,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
}

type ServerInfo struct {
	BuildVersion string        `json:"buildversion"`
	BuildTime    string        `json:"buildtime"`
	Backends     []BackendInfo `json:"backends"`
}

func CreateServerInfo(buildVersion string, buildTime string) *ServerInfo {
	return &ServerInfo{
		BuildVersion: buildVersion,
		BuildTime:    buildTime,
		Backends:     make([]BackendInfo, 0),
	}
}

func (s *ServerInfo) AddBackend(backendType string, port int, host string, tls bool, insecure bool) {
	s.Backends = append(s.Backends, BackendInfo{
		Type:     backendType,
		Port:     port,
		Host:     host,
		TLS:      tls,
		Insecure: insecure,
	})
}

func (s *ServerInfo) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (s *ServerInfo) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func CreateServerInfoFromJSON(jsonString string) (*ServerInfo, error) {
	var serverInfo *ServerInfo
	err := json.Unmarshal([]byte(jsonString), &serverInfo)
	if err != nil {
		return nil, err
	}
	return serverInfo, nil
}
