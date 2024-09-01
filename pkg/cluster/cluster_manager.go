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

type ClusterManager struct {
	ginHandler    *gin.Engine
	httpServer    *http.Server
	clusterConfig Config
	thisNode      Node
	relay         *Relay
	etcdServer    *EtcdServer
	coordinator   *Coordinator
}

func CreateClusterManager(thisNode Node, clusterConfig Config, etcdDataPath string) *ClusterManager {
	manager := &ClusterManager{}
	manager.ginHandler = gin.Default()
	manager.ginHandler.Use(cors.Default())
	manager.clusterConfig = clusterConfig
	manager.thisNode = thisNode

	manager.relay = CreateRelay(thisNode, clusterConfig, manager.ginHandler)

	manager.etcdServer = CreateEtcdServer(thisNode, clusterConfig, etcdDataPath)
	manager.etcdServer.Start()

	manager.coordinator = CreateCoordinator(thisNode, clusterConfig, manager.etcdServer, manager.ginHandler, NODE_LIST_CHECK_INTERVAL)

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(thisNode.RelayPort),
		Handler: manager.ginHandler,
	}

	manager.httpServer = httpServer

	log.WithFields(log.Fields{"Node": thisNode, "Port": thisNode.RelayPort}).Info("ClusterManager created")

	go manager.serveForever()

	return manager
}

func (manager *ClusterManager) serveForever() error {
	if err := manager.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (manager *ClusterManager) Relay() *Relay {
	return manager.relay
}

func (manager *ClusterManager) EtcdServer() *EtcdServer {
	return manager.etcdServer
}

func (manager *ClusterManager) Coordinator() *Coordinator {
	return manager.coordinator
}

func (manager *ClusterManager) BlockUntilReady() {
	manager.etcdServer.BlockUntilReady()
}

func (manager *ClusterManager) Shutdown() {
	manager.etcdServer.Stop()
	manager.etcdServer.BlockUntilStopped()
	os.RemoveAll(manager.etcdServer.StorageDir())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := manager.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("ClusterServer forced to shutdown")
	}
}
