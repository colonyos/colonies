package server

import (
	"log"
	"net/url"
	"strconv"
	"time"

	"go.etcd.io/etcd/server/v3/embed"
)

type Node struct {
	Name     string
	Host     string
	Port     int // Etcd efault typically 2379
	PeerPort int // Etcd efault typically 2380
}

type Cluster struct {
	Nodes []Node
}

type EtcdServer struct {
	cluster  Cluster
	thisNode Node
	stop     chan bool
	etcd     *embed.Etcd
}

func (cluster *Cluster) AddNode(node Node) {
	cluster.Nodes = append(cluster.Nodes, node)
}

func CreateEtcdServer(thisNode Node, cluster Cluster) *EtcdServer {
	return &EtcdServer{thisNode: thisNode, cluster: cluster, stop: make(chan bool, 1)}
}

func (cluster Cluster) buildInitialClusterStr() string {
	var str string
	for _, node := range cluster.Nodes {
		str += node.Name + "=" + "http://" + node.Host + ":" + strconv.Itoa(node.PeerPort) + ","
	}

	if len(str) > 1 {
		return str[0 : len(str)-1]
	}

	return ""
}

func (server *EtcdServer) Start() chan bool {
	cfg := embed.NewConfig()
	cfg.LogLevel = "fatal"
	name := server.thisNode.Name
	cfg.Dir = name + ".etcd"
	cfg.Name = name
	cfg.Logger = "zap"

	peerPort := strconv.Itoa(server.thisNode.PeerPort)
	port := strconv.Itoa(server.thisNode.Port)

	DefaultInitialAdvertisePeerURLs := "http://" + server.thisNode.Host + ":" + peerPort
	DefaultListenPeerURLs := "http://" + server.thisNode.Host + ":" + peerPort
	DefaultAdvertiseClientURLs := "http://" + server.thisNode.Host + ":" + port
	DefaultListenClientURLs := "http://" + server.thisNode.Host + ":" + port

	lpurl, _ := url.Parse(DefaultListenPeerURLs)           // --listen-peer-urls http://127.0.0.1:12380
	apurl, _ := url.Parse(DefaultInitialAdvertisePeerURLs) // --initial-advertise-peer-urls http://127.0.0.1:12380
	lcurl, _ := url.Parse(DefaultListenClientURLs)         // --listen-client-urls http://127.0.0.1:2379
	acurl, _ := url.Parse(DefaultAdvertiseClientURLs)      // --advertise-client-urls http://127.0.0.1:2379

	cfg.LPUrls = []url.URL{*lpurl}
	cfg.LCUrls = []url.URL{*lcurl}
	cfg.APUrls = []url.URL{*apurl}
	cfg.ACUrls = []url.URL{*acurl}
	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)
	cfg.InitialCluster = server.cluster.buildInitialClusterStr()
	cfg.InitialClusterToken = "etcd-cluster-1"

	ready := make(chan bool)

	go func() {
		etcd, err := embed.StartEtcd(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer etcd.Close()
		server.etcd = etcd
		for {
			select {
			case <-server.stop:
				etcd.Server.Stop()
				return
			case <-etcd.Server.ReadyNotify():
				log.Printf("Server is ready!")
				ready <- true
			case <-time.After(60 * time.Second):
				etcd.Server.Stop() // trigger a shutdown
				log.Printf("Server took too long to start!")
			}
			log.Fatal(<-etcd.Err())
		}
	}()

	return ready
}

func (server *EtcdServer) Stop() {
	server.stop <- true
}

func (server *EtcdServer) Leader() string {
	leader := server.etcd.Server.Leader()
	for _, member := range server.etcd.Server.Cluster().Members() {
		if member.ID == leader {
			return member.Name
		}
	}
	return ""
}
