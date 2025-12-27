package cluster

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// RelayMessage contains the message data and a channel to signal processing completion
type RelayMessage struct {
	Data []byte
	Done chan struct{}
}

// MessageHandler is a function that processes relay messages
type MessageHandler func(data []byte)

type RelayServer struct {
	ginHandler    *gin.Engine
	httpServer    *http.Server
	restyClient   *resty.Client
	clusterConfig Config
	thisNode      Node
	handlers      []MessageHandler
	handlersMu    sync.RWMutex
}

func CreateRelayServer(thisNode Node, clusterConfig Config) *RelayServer {
	server := &RelayServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())
	server.restyClient = resty.New()
	server.clusterConfig = clusterConfig
	server.thisNode = thisNode
	server.handlers = make([]MessageHandler, 0)

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
		return
	}

	// Call all registered handlers synchronously
	server.handlersMu.RLock()
	handlers := make([]MessageHandler, len(server.handlers))
	copy(handlers, server.handlers)
	server.handlersMu.RUnlock()

	for _, handler := range handlers {
		handler(jsonBytes)
	}

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

// Subscribe registers a handler to receive relay messages
// The handler is called synchronously for each incoming message
func (server *RelayServer) Subscribe(handler MessageHandler) {
	server.handlersMu.Lock()
	defer server.handlersMu.Unlock()
	server.handlers = append(server.handlers, handler)
}

// Receive returns a channel for receiving relay messages (legacy interface)
// Messages are dispatched asynchronously to avoid blocking the HTTP handler
func (server *RelayServer) Receive() chan RelayMessage {
	ch := make(chan RelayMessage, 1000) // Large buffer to handle bursts
	server.Subscribe(func(data []byte) {
		// Make a copy of data for async dispatch
		dataCopy := make([]byte, len(data))
		copy(dataCopy, data)
		// Dispatch in goroutine to avoid blocking HTTP handler while ensuring delivery
		go func() {
			// Non-blocking send - drop message if channel is full (receiver stopped)
			select {
			case ch <- RelayMessage{Data: dataCopy, Done: nil}:
			default:
				// Channel full, receiver likely stopped - drop message silently
			}
		}()
	})
	return ch
}

func (server *RelayServer) Shutdown() { // TODO: unittest
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("RelayServer forced to shutdown")
	}

}
