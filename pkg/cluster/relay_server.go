package cluster

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type RelayServer struct {
	ginHandler    *gin.Engine
	httpServer    *http.Server
	restyClient   *resty.Client
	clusterConfig Config
	thisNode      Node
	incoming      chan []byte
}

func CreateRelayServer(thisNode Node, clusterConfig Config) *RelayServer {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	server := &RelayServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())
	server.restyClient = resty.New()
	server.clusterConfig = clusterConfig
	server.thisNode = thisNode
	server.incoming = make(chan []byte)

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(thisNode.RelayPort),
		Handler: server.ginHandler,
	}
	server.httpServer = httpServer

	go server.serveForever()
	server.setupRoutes()

	return server
}

func (server *RelayServer) serveForever() error {
	if err := server.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (server *RelayServer) setupRoutes() {
	server.ginHandler.POST("/relay", server.handleRelayRequest)
}

func (server *RelayServer) handleRelayRequest(c *gin.Context) {
	jsonBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errMsg := "Bad relay request"
		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
		c.String(http.StatusBadRequest, errMsg)
	}

	server.incoming <- jsonBytes

	c.String(http.StatusOK, "")
}

// Send a message to all ReplayServers in the Cluster
func (server *RelayServer) Broadcast(msg []byte) error {
	for _, node := range server.clusterConfig.Nodes {
		if node.Name != server.thisNode.Name {
			_, err := server.restyClient.R().
				SetBody(msg).
				Post("http://" + node.Host + ":" + strconv.Itoa(node.RelayPort) + "/relay")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (server *RelayServer) Receive() chan []byte {
	return server.incoming
}

func (server *RelayServer) Shutdown() { // TODO: unittest
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("RelayServer forced to shutdown")
	}

}
