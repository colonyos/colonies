package cluster

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ClusterServer struct {
	ginHandler    *gin.Engine
	httpServer    *http.Server
	clusterConfig Config
	thisNode      Node
	relay         *Relay
	cluster       *Cluster
	etcdServer    *EtcdServer
}

func CreateClusterServer(thisNode Node, clusterConfig Config, etcdDataPath string) *ClusterServer {
	server := &ClusterServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())
	server.clusterConfig = clusterConfig
	server.thisNode = thisNode

	server.relay = CreateRelay(thisNode, clusterConfig, server.ginHandler)
	server.cluster = CreateCluster(thisNode, clusterConfig, server.ginHandler)

	server.etcdServer = CreateEtcdServer(thisNode, clusterConfig, etcdDataPath)
	server.etcdServer.Start()
	server.etcdServer.WaitToStart()

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(thisNode.RelayPort),
		Handler: server.ginHandler,
	}

	server.httpServer = httpServer

	log.WithFields(log.Fields{"Node": thisNode, "Port": thisNode.RelayPort}).Info("ClusterServer created")

	go server.serveForever()

	return server
}

func (server *ClusterServer) serveForever() error {
	if err := server.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (server *ClusterServer) Relay() *Relay {
	return server.relay
}

func (server *ClusterServer) Cluster() *Cluster {
	return server.cluster
}

func (server *ClusterServer) EtcdServer() *EtcdServer {
	return server.etcdServer
}

func (server *ClusterServer) Shutdown() {
	// Stop the Etcd server
	server.etcdServer.Stop()
	server.etcdServer.WaitToStop()
	os.RemoveAll(server.etcdServer.StorageDir())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("ClusterServer forced to shutdown")
	}

}
