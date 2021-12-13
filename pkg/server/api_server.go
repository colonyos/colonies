package server

import (
	"fmt"
	"strconv"

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
	engine             *gin.Engine
	coloniesController *ColoniesController
	tlsPrivateKeyPath  string
	tlsCertPath        string
	port               int
}

func CreateAPIServer(coloniesController *ColoniesController, port int, tlsPrivateKeyPath string, tlsCertPath string) *APIServer {
	apiServer := &APIServer{}
	apiServer.engine = gin.Default()
	apiServer.coloniesController = coloniesController
	apiServer.port = port
	apiServer.tlsPrivateKeyPath = tlsPrivateKeyPath
	apiServer.tlsCertPath = tlsCertPath
	apiServer.setupRoutes()
	gin.SetMode(gin.ReleaseMode)
	return apiServer
}

func (apiServer *APIServer) setupRoutes() {
	apiServer.engine.POST("/colonies", apiServer.handleAddColonyRequest)
}

func (apiServer *APIServer) handleAddColonyRequest(c *gin.Context) {
	fmt.Println("handleAddColonyRequest")
}

func (apiServer *APIServer) ServeForever() {
	apiServer.engine.RunTLS(":"+strconv.Itoa(apiServer.port), apiServer.tlsCertPath, apiServer.tlsPrivateKeyPath)
}
