package server

import (
	"colonies/pkg/core"
	"colonies/pkg/database"
	"colonies/pkg/logging"
	"colonies/pkg/security"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	ownership          security.Ownership
}

func CreateAPIServer(db database.Database, port int, apiKey string, tlsPrivateKeyPath string, tlsCertPath string) *APIServer {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	apiServer := &APIServer{}
	apiServer.ginHandler = gin.Default()

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: apiServer.ginHandler,
	}

	apiServer.httpServer = httpServer
	apiServer.coloniesController = CreateColoniesController(db)
	apiServer.ownership = security.CreateOwnership(db)
	apiServer.apiKey = apiKey
	apiServer.port = port
	apiServer.tlsPrivateKeyPath = tlsPrivateKeyPath
	apiServer.tlsCertPath = tlsCertPath

	apiServer.setupRoutes()

	logrus.SetLevel(logrus.DebugLevel)
	logging.Log().Info("Starting Colonies API server at port: " + strconv.Itoa(port))

	return apiServer
}

func (apiServer *APIServer) setupRoutes() {
	apiServer.ginHandler.GET("/colonies", apiServer.handleGetColoniesRequest)
	apiServer.ginHandler.GET("/colonies/:id", apiServer.handleGetColonyRequest)
	apiServer.ginHandler.POST("/colonies", apiServer.handleAddColonyRequest)
	apiServer.ginHandler.POST("/colonies/:id/workers", apiServer.handleAddWorkerRequest)
}

func (apiServer *APIServer) handleGetColoniesRequest(c *gin.Context) {
	err := security.VerifyAPIKey(c.GetHeader("Api-Key"), apiServer.apiKey)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	colonies, err := apiServer.coloniesController.GetColonies()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := core.ColonyArrayToJSON(colonies)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (apiServer *APIServer) handleGetColonyRequest(c *gin.Context) {
	colonyID := c.Param("id")

	params := c.Request.URL.Query()
	randomData := params["dummydata"][0]

	err := security.VerifyColonyOwnership(colonyID, string(randomData), c.GetHeader("Signature"), apiServer.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	colony, err := apiServer.coloniesController.GetColony(colonyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := colony.ToJSON()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (apiServer *APIServer) handleAddColonyRequest(c *gin.Context) {
	err := security.VerifyAPIKey(c.GetHeader("Api-Key"), apiServer.apiKey)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	colony, err := core.CreateColonyFromJSON(string(jsonData))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = apiServer.coloniesController.AddColony(colony)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func (apiServer *APIServer) handleAddWorkerRequest(c *gin.Context) {
	colonyID := c.Param("id")

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	worker, err := core.CreateWorkerFromJSON(string(jsonData))
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = security.VerifyColonyOwnership(colonyID, string(jsonData), c.GetHeader("Signature"), apiServer.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = apiServer.coloniesController.AddWorker(worker)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": err.Error()})
		return
	}

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
		logging.Log().Warning("Server forced to shutdown:", err)
	}
}
