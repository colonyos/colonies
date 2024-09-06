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
	Nodes  map[string]*Node `json:"nodes"`
	Leader *Node            `json:"leader"`
}

func EmptyConfig() *Config {
	return &Config{
		Nodes: make(map[string]*Node),
	}
}

func NewConfig(nodes []*Node, leaderNode *Node) *Config {
	config := EmptyConfig()

	for _, node := range nodes {
		config.Nodes[node.Name] = node
	}

	config.Leader = leaderNode

	return config
}

func (config *Config) AddNode(node *Node) {
	config.Nodes[node.Name] = node
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

	if len(config.Nodes) != len(config2.Nodes) {
		return false
	}

	if config.Leader == nil && config2.Leader != nil {
		return false
	}

	if config.Leader != nil && config2.Leader == nil {
		return false
	}

	if config.Leader != nil && config2.Leader != nil && config.Leader.Name != config2.Leader.Name {
		return false
	}

	counter := 0

	for _, node := range config.Nodes {
		if _, ok := config2.Nodes[node.Name]; ok {
			counter++
		}
	}

	if counter != len(config.Nodes) && counter != len(config.Nodes) {
		return false
	}

	return true
}

func (config *Config) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
