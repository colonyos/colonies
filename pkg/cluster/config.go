package cluster

import "encoding/json"

type Node struct {
	Name           string `json:"name"`
	Host           string `json:"host"`
	EtcdClientPort int    `json:"port"`     // Etcd default 2379
	EtcdPeerPort   int    `json:"peerport"` // Etcd default 2380
	RelayPort      int    `json:"relayport"`
	APIPort        int    `json:"apiport"`
}

func (node *Node) Equals(node2 *Node) bool {
	if node.Name != node2.Name {
		return false
	}
	if node.Host != node2.Host {
		return false
	}
	if node.EtcdClientPort != node2.EtcdClientPort {
		return false
	}
	if node.EtcdPeerPort != node2.EtcdPeerPort {
		return false
	}

	if node.RelayPort != node2.RelayPort {
		return false
	}

	if node.APIPort != node2.APIPort {
		return false
	}

	return true
}

type Config struct {
	Nodes  []Node `json:"nodes"`
	Leader Node   `json:"leader"`
}

func (config *Config) AddNode(node Node) {
	config.Nodes = append(config.Nodes, node)
}

func ConvertJSONToConfig(jsonString string) (*Config, error) {
	var config *Config
	err := json.Unmarshal([]byte(jsonString), &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (config *Config) Equals(config2 *Config) bool {
	if config2 == nil {
		return false
	}

	if !config.Leader.Equals(&config2.Leader) {
		return false
	}

	counter := 0
	for _, node1 := range config.Nodes {
		for _, node2 := range config2.Nodes {
			if node1.Equals(&node2) {
				counter++
			}
		}
	}

	if counter == len(config.Nodes) && counter == len(config.Nodes) {
		return true
	}

	return false
}

func (config *Config) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
