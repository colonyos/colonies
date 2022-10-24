package server

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/security/validator"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ColoniesServer struct {
	ginHandler        *gin.Engine
	controller        *coloniesController
	serverID          string
	tls               bool
	tlsPrivateKeyPath string
	tlsCertPath       string
	port              int
	httpServer        *http.Server
	crypto            security.Crypto
	validator         security.Validator
	db                database.Database
}

func CreateColoniesServer(db database.Database,
	port int,
	serverID string,
	tls bool,
	tlsPrivateKeyPath string,
	tlsCertPath string,
	thisNode cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string) *ColoniesServer {

	server := &ColoniesServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())

	server.db = db

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: server.ginHandler,
	}

	server.httpServer = httpServer
	server.controller = createColoniesController(db, thisNode, clusterConfig, etcdDataPath)
	server.serverID = serverID
	server.tls = tls
	server.port = port
	server.tlsPrivateKeyPath = tlsPrivateKeyPath
	server.tlsCertPath = tlsCertPath
	server.crypto = crypto.CreateCrypto()
	server.validator = validator.CreateValidator(db)

	server.setupRoutes()

	return server
}

func (server *ColoniesServer) setupRoutes() {
	server.ginHandler.POST("/api", server.handleAPIRequest)
	server.ginHandler.GET("/health", server.handleHealthRequest)
	server.ginHandler.GET("/pubsub", server.handleWSRequest)
}

func (server *ColoniesServer) parseSignature(jsonString string, signature string) (string, error) {
	recoveredID, err := server.crypto.RecoverID(jsonString, signature)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call crypto.RecoverID()")
		return "", err
	}

	return recoveredID, nil
}

func (server *ColoniesServer) handleHealthRequest(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func (server *ColoniesServer) handleAPIRequest(c *gin.Context) {
	jsonBytes, err := ioutil.ReadAll(c.Request.Body)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Error("Bad request")
		return
	}

	rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(jsonBytes))
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	// Version does not require a valid private key
	if rpcMsg.PayloadType == rpc.VersionPayloadType {
		server.handleVersionHTTPRequest(c, rpcMsg.PayloadType, rpcMsg.DecodePayload())
		return
	}

	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	switch rpcMsg.PayloadType {
	// Colony handlers
	case rpc.AddColonyPayloadType:
		server.handleAddColonyHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteColonyPayloadType:
		server.handleDeleteColonyHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColoniesPayloadType:
		server.handleGetColoniesHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColonyPayloadType:
		server.handleGetColonyHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Runtime handlers
	case rpc.AddRuntimePayloadType:
		server.handleAddRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetRuntimesPayloadType:
		server.handleGetRuntimesHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetRuntimePayloadType:
		server.handleGetRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ApproveRuntimePayloadType:
		server.handleApproveRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RejectRuntimePayloadType:
		server.handleRejectRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteRuntimePayloadType:
		server.handleDeleteRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Process handlers
	case rpc.SubmitProcessSpecPayloadType:
		server.handleSubmitProcessSpecHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.AssignProcessPayloadType:
		server.handleAssignProcessHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessHistPayloadType:
		server.handleGetProcessHistHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessesPayloadType:
		server.handleGetProcessesHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessPayloadType:
		server.handleGetProcessHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteProcessPayloadType:
		server.handleDeleteProcessHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteAllProcessesPayloadType:
		server.handleDeleteAllProcessesHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.CloseSuccessfulPayloadType:
		server.handleCloseSuccessfulHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.CloseFailedPayloadType:
		server.handleCloseFailedHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColonyStatisticsPayloadType:
		server.handleColonyStatisticsHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Attribute handlers
	case rpc.AddAttributePayloadType:
		server.handleAddAttributeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetAttributePayloadType:
		server.handleGetAttributeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Workflow and processgraph handlers
	case rpc.SubmitWorkflowSpecPayloadType:
		server.handleSubmitWorkflowHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessGraphPayloadType:
		server.handleGetProcessGraphHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessGraphsPayloadType:
		server.handleGetProcessGraphsHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteProcessGraphPayloadType:
		server.handleDeleteProcessGraphHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteAllProcessGraphsPayloadType:
		server.handleDeleteAllProcessGraphsHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Generators handlers
	case rpc.AddGeneratorPayloadType:
		server.handleAddGeneratorHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetGeneratorPayloadType:
		server.handleGetGeneratorHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ResolveGeneratorPayloadType:
		server.handleResolveGeneratorHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetGeneratorsPayloadType:
		server.handleGetGeneratorsHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.PackGeneratorPayloadType:
		server.handlePackGeneratorHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteGeneratorPayloadType:
		server.handleDeleteGeneratorHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Cron handlers
	case rpc.AddCronPayloadType:
		server.handleAddCronHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetCronPayloadType:
		server.handleGetCronHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetCronsPayloadType:
		server.handleGetCronsHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RunCronPayloadType:
		server.handleRunCronHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteCronPayloadType:
		server.handleDeleteCronHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Server handlers
	case rpc.GetStatisiticsPayloadType:
		server.handleStatisticsHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetClusterPayloadType:
		server.handleGetClusterHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	default:
		errMsg := "invalid rpcMsg.PayloadType, " + rpcMsg.PayloadType
		if server.handleHTTPError(c, errors.New(errMsg), http.StatusForbidden) {
			log.Error(errMsg)
			return
		}
	}
}

func (server *ColoniesServer) generateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error) {
	failure := core.CreateFailure(errorCode, err.Error())
	jsonString, err := failure.ToJSON()
	if err != nil {
		return nil, err
	}
	rpcReplyMsg, err := rpc.CreateRPCErrorReplyMsg(rpc.ErrorPayloadType, jsonString)
	if err != nil {
		return nil, err
	}

	return rpcReplyMsg, nil
}

func (server *ColoniesServer) handleHTTPError(c *gin.Context, err error, errorCode int) bool {
	if err != nil {
		if !strings.HasPrefix(err.Error(), "No processes can be selected for runtime with Id") {
			log.Debug(err)
		}

		rpcReplyMsg, err := server.generateRPCErrorMsg(err, errorCode)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to call server.generateRPCErrorMsg()")
		}
		rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to call pcReplyMsg.ToJSON()")
		}

		c.String(errorCode, rpcReplyMsgJSONString)
		return true
	}

	return false
}

func (server *ColoniesServer) sendHTTPReply(c *gin.Context, payloadType string, jsonString string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *ColoniesServer) sendEmptyHTTPReply(c *gin.Context, payloadType string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, "{}")
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *ColoniesServer) ServeForever() error {
	if server.tls {
		if err := server.httpServer.ListenAndServeTLS(server.tlsCertPath, server.tlsPrivateKeyPath); err != nil && errors.Is(err, http.ErrServerClosed) {
			return err
		}
	} else {
		if err := server.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

func (server *ColoniesServer) Shutdown() {
	server.controller.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("ColoniesServer forced to shutdown")
	}

}
