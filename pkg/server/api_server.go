package server

import (
	"colonies/pkg/security"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

/*
/products	GET	Hämtar en lista med alla produkter
/products	POST	Skapar en ny produkt
/products/{ProductID}	GET	Returnerar en produkt
/products/{ProductID}	PUT	Ersätter en produkt
/products/{ProductID}	DELETE	Tar bort en produkt
/product_groups/{ProductGroupID}	GET	Returnerar en produktgrupp
/products/{ProductID}	PATCH	Uppdaterar en produkt
*/

type APIServer struct {
	ginHandler         *gin.Engine
	coloniesController *ColoniesController
	apiKey             string
	tlsPrivateKeyPath  string
	tlsCertPath        string
	port               int
	httpServer         *http.Server
}

func CreateAPIServer(coloniesController *ColoniesController, port int, apiKey string, tlsPrivateKeyPath string, tlsCertPath string) *APIServer {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	apiServer := &APIServer{}
	apiServer.ginHandler = gin.Default()

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: apiServer.ginHandler,
	}

	apiServer.httpServer = httpServer
	apiServer.coloniesController = coloniesController
	apiServer.apiKey = apiKey
	apiServer.port = port
	apiServer.tlsPrivateKeyPath = tlsPrivateKeyPath
	apiServer.tlsCertPath = tlsCertPath

	apiServer.setupRoutes()

	return apiServer
}

func (apiServer *APIServer) setupRoutes() {
	apiServer.ginHandler.POST("/colonies", apiServer.handleAddColonyRequest)
	apiServer.ginHandler.POST("/colonies/:id", apiServer.handleAddWorkerRequest)
}

func (apiServer *APIServer) handleAddColonyRequest(c *gin.Context) {
	err := security.CheckAPIKey(c, apiServer.apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(string(jsonData))

	c.JSON(http.StatusOK, "")
}

func (apiServer *APIServer) handleAddWorkerRequest(c *gin.Context) {
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	colonyID := c.Param("id")
	fmt.Println("-colonyID---------------")
	fmt.Println(colonyID)
	fmt.Println("----------------")

	// TODO
	// 1. Parse JOSN and create a worker object
	// 2. Verify signature
	fmt.Println(string(jsonData))

	c.JSON(http.StatusOK, "")
}

func (apiServer *APIServer) ServeForever() error {
	if err := apiServer.httpServer.ListenAndServeTLS(apiServer.tlsCertPath, apiServer.tlsPrivateKeyPath); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (apiServer *APIServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
