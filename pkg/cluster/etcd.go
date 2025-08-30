package cluster

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

type EtcdServer struct {
	thisNode   Node
	config     Config
	ready      chan bool
	stop       chan bool
	stopped    chan bool
	dataPath   string
	etcd       *embed.Etcd
	cfg        *embed.Config
	etcdClient *clientv3.Client
}

func CreateEtcdServer(thisNode Node, config Config, dataPath string) *EtcdServer {
	server := &EtcdServer{thisNode: thisNode,
		config:   config,
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

	peerPort := strconv.Itoa(server.thisNode.EtcdPeerPort)
	clientPort := strconv.Itoa(server.thisNode.EtcdClientPort)

	initialAdvertisePeerURLs := "http://" + server.thisNode.Host + ":" + peerPort
	listenPeerURLs := "http://0.0.0.0:" + peerPort
	advertiseClientURLs := "http://" + server.thisNode.Host + ":" + clientPort
	listenClientURLs := "http://0.0.0.0:" + clientPort

	lpurl, _ := url.Parse(listenPeerURLs)
	apurl, _ := url.Parse(initialAdvertisePeerURLs)
	lcurl, _ := url.Parse(listenClientURLs)
	acurl, _ := url.Parse(advertiseClientURLs)

	cfg.ListenPeerUrls = []url.URL{*lpurl}
	cfg.ListenClientUrls = []url.URL{*lcurl}
	cfg.AdvertisePeerUrls = []url.URL{*apurl}
	cfg.AdvertiseClientUrls = []url.URL{*acurl}
	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)
	cfg.InitialCluster = server.buildInitialClusterStr()
	cfg.InitialClusterToken = "etcd-cluster-1"

	server.cfg = cfg

	return server
}

func (server *EtcdServer) buildInitialClusterStr() string {
	var str string
	for _, node := range server.config.Nodes {
		str += node.Name + "=" + "http://" + node.Host + ":" + strconv.Itoa(node.EtcdPeerPort) + ","
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
			// Initialize etcd client
			endpoints := []string{"localhost:" + strconv.Itoa(server.thisNode.EtcdClientPort)}
			client, err := clientv3.New(clientv3.Config{
				Endpoints:   endpoints,
				DialTimeout: 5 * time.Second,
			})
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to create etcd client")
				log.Fatal(err)
			}
			server.etcdClient = client

			log.WithFields(log.Fields{
				"Name":           server.thisNode.Name,
				"Host":           server.thisNode.Host,
				"DataPath":       server.dataPath,
				"EtcdClientPort": server.thisNode.EtcdClientPort,
				"EtcdPeerPort":   server.thisNode.EtcdPeerPort}).Info("EtcdServer is ready")
			server.ready <- true
			<-server.stop
			if server.etcdClient != nil {
				server.etcdClient.Close()
			}
			etcd.Server.Stop()
			log.WithFields(log.Fields{
				"Name":           server.thisNode.Name,
				"Host":           server.thisNode.Host,
				"DataPath":       server.dataPath,
				"EtcdClientPort": server.thisNode.EtcdClientPort,
				"EtcPeerPort":    server.thisNode.EtcdPeerPort}).Info("EtcdServer stopped")
			server.stopped <- true
		case <-time.After(600 * time.Second):
			log.WithFields(log.Fields{
				"Name":           server.thisNode.Name,
				"Host":           server.thisNode.Host,
				"DataPath":       server.dataPath,
				"EtcdClientPort": server.thisNode.EtcdClientPort,
				"EtcdPeerPort":   server.thisNode.EtcdPeerPort}).Error("EtcdServer took too long time to start")
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
		for _, node := range server.config.Nodes {
			if node.Name == member.Name {
				nodes = append(nodes, node)
			}

		}
	}

	return nodes
}

func (server *EtcdServer) CurrentCluster() Config {
	nodes := server.Members()
	leader := server.Leader()

	var leaderNode Node
	for _, node := range nodes {
		if node.Name == leader {
			leaderNode = node
			break
		}
	}

	return Config{Nodes: nodes, Leader: leaderNode}
}

func (server *EtcdServer) PauseColonyAssignments(colonyName string) error {
	if server.etcdClient == nil {
		return errors.New("etcd client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/colonies/colony/%s/assignments-paused", colonyName)
	_, err := server.etcdClient.Put(ctx, key, "true")
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Colony": colonyName}).Error("Failed to pause colony assignments in etcd")
		return err
	}

	log.WithFields(log.Fields{"Colony": colonyName}).Info("Colony process assignments have been paused")
	return nil
}

func (server *EtcdServer) ResumeColonyAssignments(colonyName string) error {
	if server.etcdClient == nil {
		return errors.New("etcd client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/colonies/colony/%s/assignments-paused", colonyName)
	_, err := server.etcdClient.Delete(ctx, key)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Colony": colonyName}).Error("Failed to resume colony assignments in etcd")
		return err
	}

	log.WithFields(log.Fields{"Colony": colonyName}).Info("Colony process assignments have been resumed")
	return nil
}

func (server *EtcdServer) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	if server.etcdClient == nil {
		return false, errors.New("etcd client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/colonies/colony/%s/assignments-paused", colonyName)
	resp, err := server.etcdClient.Get(ctx, key)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Colony": colonyName}).Error("Failed to check colony assignment pause state in etcd")
		return false, err
	}

	// If key doesn't exist, assignments are not paused
	return len(resp.Kvs) > 0, nil
}
