package etcd

import (
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/server/v3/embed"
)

type EtcdServer struct {
	thisNode Node
	cluster  Cluster
	ready    chan bool
	stop     chan bool
	stopped  chan bool
	dataPath string
	etcd     *embed.Etcd
	cfg      *embed.Config
}

func CreateEtcdServer(thisNode Node, cluster Cluster, dataPath string) *EtcdServer {
	server := &EtcdServer{thisNode: thisNode,
		cluster:  cluster,
		ready:    make(chan bool, 1),
		stop:     make(chan bool, 1),
		stopped:  make(chan bool, 1),
		dataPath: dataPath}

	cfg := embed.NewConfig()
	cfg.LogLevel = "fatal"
	name := server.thisNode.Name
	cfg.Dir = server.dataPath + "/" + name + ".etcd"
	cfg.Name = name
	cfg.Logger = "zap"

	peerPort := strconv.Itoa(server.thisNode.PeerPort)
	clientPort := strconv.Itoa(server.thisNode.ClientPort)

	initialAdvertisePeerURLs := "http://" + server.thisNode.Host + ":" + peerPort
	listenPeerURLs := "http://0.0.0.0:" + peerPort
	advertiseClientURLs := "http://" + server.thisNode.Host + ":" + clientPort
	listenClientURLs := "http://0.0.0.0:" + clientPort

	lpurl, _ := url.Parse(listenPeerURLs)
	apurl, _ := url.Parse(initialAdvertisePeerURLs)
	lcurl, _ := url.Parse(listenClientURLs)
	acurl, _ := url.Parse(advertiseClientURLs)

	cfg.LPUrls = []url.URL{*lpurl}
	cfg.LCUrls = []url.URL{*lcurl}
	cfg.APUrls = []url.URL{*apurl}
	cfg.ACUrls = []url.URL{*acurl}
	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)
	cfg.InitialCluster = server.buildInitialClusterStr()
	cfg.InitialClusterToken = "etcd-cluster-1"

	server.cfg = cfg

	return server
}

func (server *EtcdServer) buildInitialClusterStr() string {
	var str string
	for _, node := range server.cluster.Nodes {
		str += node.Name + "=" + "http://" + node.Host + ":" + strconv.Itoa(node.PeerPort) + ","
	}

	if len(str) > 1 {
		return str[0 : len(str)-1]
	}

	return ""
}

func (server *EtcdServer) StorageDir() string {
	return server.cfg.Dir
}

func (server *EtcdServer) Start() {
	go func() {
		etcd, err := embed.StartEtcd(server.cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer etcd.Close()
		server.etcd = etcd
		select {
		case <-etcd.Server.ReadyNotify():
			log.WithFields(log.Fields{
				"Name":       server.thisNode.Name,
				"Host":       server.thisNode.Host,
				"DataPath":   server.dataPath,
				"ClientPort": server.thisNode.ClientPort,
				"PeerPort":   server.thisNode.PeerPort}).Info("EtcdServer is ready")
			server.ready <- true
			<-server.stop
			etcd.Server.Stop()
			log.WithFields(log.Fields{
				"Name":       server.thisNode.Name,
				"Host":       server.thisNode.Host,
				"DataPath":   server.dataPath,
				"ClientPort": server.thisNode.ClientPort,
				"PeerPort":   server.thisNode.PeerPort}).Info("EtcdServer stopped")
			server.stopped <- true
		case <-time.After(600 * time.Second):
			log.WithFields(log.Fields{
				"Name":       server.thisNode.Name,
				"Host":       server.thisNode.Host,
				"DataPath":   server.dataPath,
				"ClientPort": server.thisNode.ClientPort,
				"PeerPort":   server.thisNode.PeerPort}).Error("EtcdServer took too long time to start")
			etcd.Server.Stop()
			log.Fatal(<-etcd.Err())
		}
	}()
}

func (server *EtcdServer) Stop() {
	server.stop <- true
}

func (server *EtcdServer) WaitToStart() {
	<-server.ready
}

func (server *EtcdServer) WaitToStop() {
	<-server.stopped
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

func (server *EtcdServer) Members() []Node {
	var nodes []Node
	for _, member := range server.etcd.Server.Cluster().Members() {
		for _, node := range server.cluster.Nodes {
			if node.Name == member.Name {
				nodes = append(nodes, node)
			}

		}
	}

	return nodes
}

func (server *EtcdServer) CurrentCluster() Cluster {
	nodes := server.Members()
	leader := server.Leader()

	var leaderNode Node
	for _, node := range nodes {
		if node.Name == leader {
			leaderNode = node
			break
		}
	}

	return Cluster{Nodes: nodes, Leader: leaderNode}
}
