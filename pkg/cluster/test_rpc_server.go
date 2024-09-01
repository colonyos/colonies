package cluster

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type testRPCServer struct {
	ginHandler    *gin.Engine
	httpServer    *http.Server
	clusterConfig Config
	thisNode      Node
	rpc           *clusterRPC
}

func createTestRPCServer(thisNode Node, clusterConfig Config, etcdDataPath string, purgeInterval time.Duration) *testRPCServer {
	server := &testRPCServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())
	server.clusterConfig = clusterConfig
	server.thisNode = thisNode

	server.rpc = createClusterRPC(thisNode, clusterConfig, server.ginHandler, purgeInterval)

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(thisNode.RelayPort),
		Handler: server.ginHandler,
	}

	server.httpServer = httpServer

	log.WithFields(log.Fields{"Node": thisNode, "Port": thisNode.RelayPort}).Info("ClusterManager created")

	go server.serveForever()

	return server
}

func (server *testRPCServer) serveForever() error {
	if err := server.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (server *testRPCServer) clusterRPC() *clusterRPC {
	return server.rpc
}

func (server *testRPCServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("ClusterServer forced to shutdown")
	}
}
