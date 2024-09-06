package cluster

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Relay struct {
	ginHandler    *gin.Engine
	restyClient   *resty.Client
	clusterConfig *Config
	thisNode      *Node
	incoming      chan []byte
}

func CreateRelay(thisNode *Node, clusterConfig *Config, ginHandler *gin.Engine) *Relay {
	relay := &Relay{}
	relay.ginHandler = ginHandler
	relay.restyClient = resty.New()
	relay.clusterConfig = clusterConfig
	relay.thisNode = thisNode
	relay.incoming = make(chan []byte)

	relay.setupRoutes()

	return relay
}

func (relay *Relay) setupRoutes() {
	relay.ginHandler.POST("/relay", relay.handleRelayRequest)
}

func (relay *Relay) handleRelayRequest(c *gin.Context) {
	jsonBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		errMsg := "Bad relay request"
		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
		c.String(http.StatusBadRequest, errMsg)
	}

	relay.incoming <- jsonBytes

	c.String(http.StatusOK, "")
}

// Send a message to all ReplayServers in the Cluster
func (relay *Relay) Broadcast(msg []byte) error {
	for _, node := range relay.clusterConfig.Nodes {
		if node.Name != relay.thisNode.Name {
			_, err := relay.restyClient.R().
				SetBody(msg).
				Post("http://" + node.Host + ":" + strconv.Itoa(node.RelayPort) + "/relay")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (relay *Relay) Receive() chan []byte {
	return relay.incoming
}
