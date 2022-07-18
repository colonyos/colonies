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
	path     string
	etcd     *embed.Etcd
	cfg      *embed.Config
}

func CreateEtcdServer(thisNode Node, cluster Cluster, path string) *EtcdServer {
	server := &EtcdServer{thisNode: thisNode,
		cluster: cluster,
		ready:   make(chan bool, 1),
		stop:    make(chan bool, 1),
		stopped: make(chan bool, 1),
		path:    path}

	cfg := embed.NewConfig()
	cfg.LogLevel = "fatal"
	name := server.thisNode.Name
	cfg.Dir = server.path + "/" + name + ".etcd"
	cfg.Name = name
	cfg.Logger = "zap"

	peerPort := strconv.Itoa(server.thisNode.PeerPort)
	port := strconv.Itoa(server.thisNode.Port)

	initialAdvertisePeerURLs := "http://" + server.thisNode.Host + ":" + peerPort
	listenPeerURLs := "http://" + server.thisNode.Host + ":" + peerPort
	advertiseClientURLs := "http://" + server.thisNode.Host + ":" + port
	listenClientURLs := "http://" + server.thisNode.Host + ":" + port

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
			log.WithFields(log.Fields{"Name": server.thisNode.Name,
				"Host":     server.thisNode.Host,
				"Port":     server.thisNode.Port,
				"PeerPort": server.thisNode.PeerPort}).Info("Etcd server is ready")
			server.ready <- true
			<-server.stop
			etcd.Server.Stop()
			log.WithFields(log.Fields{"Name": server.thisNode.Name,
				"Host":     server.thisNode.Host,
				"Port":     server.thisNode.Port,
				"PeerPort": server.thisNode.PeerPort}).Info("Etcd server stopped")
			server.stopped <- true
		case <-time.After(60 * time.Second):
			log.WithFields(log.Fields{"Name": server.thisNode.Name,
				"Host":     server.thisNode.Host,
				"Port":     server.thisNode.Port,
				"PeerPort": server.thisNode.PeerPort}).Error("Etcd server took too long time to start")
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
