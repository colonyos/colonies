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
	attributehandlers "github.com/colonyos/colonies/pkg/server/handlers/attribute"
	cronhandlers "github.com/colonyos/colonies/pkg/server/handlers/cron"
	filehandlers "github.com/colonyos/colonies/pkg/server/handlers/file"
	functionhandlers "github.com/colonyos/colonies/pkg/server/handlers/function"
	generatorhandlers "github.com/colonyos/colonies/pkg/server/handlers/generator"
	securityhandlers "github.com/colonyos/colonies/pkg/server/handlers/security"
	"github.com/colonyos/colonies/pkg/server/handlers/user"
	"github.com/colonyos/colonies/pkg/server/handlers/colony"
	"github.com/colonyos/colonies/pkg/server/handlers/executor"
	loghandlers "github.com/colonyos/colonies/pkg/server/handlers/log"
	"github.com/colonyos/colonies/pkg/server/handlers/process"
	"github.com/colonyos/colonies/pkg/server/handlers/processgraph"
	serverhandlers "github.com/colonyos/colonies/pkg/server/handlers/server"
	snapshothandlers "github.com/colonyos/colonies/pkg/server/handlers/snapshot"
	websockethandlers "github.com/colonyos/colonies/pkg/server/handlers/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ColoniesServer struct {
	ginHandler              *gin.Engine
	controller              controller
	serverID                string
	tls                     bool
	tlsPrivateKeyPath       string
	tlsCertPath             string
	port                    int
	httpServer              *http.Server
	crypto                  security.Crypto
	validator               security.Validator
	userDB                  database.UserDatabase
	colonyDB                database.ColonyDatabase
	executorDB              database.ExecutorDatabase
	functionDB              database.FunctionDatabase
	processDB               database.ProcessDatabase
	attributeDB             database.AttributeDatabase
	processGraphDB          database.ProcessGraphDatabase
	generatorDB             database.GeneratorDatabase
	cronDB                  database.CronDatabase
	logDB                   database.LogDatabase
	fileDB                  database.FileDatabase
	snapshotDB              database.SnapshotDatabase
	securityDB              database.SecurityDatabase
	exclusiveAssign         bool
	allowExecutorReregister bool
	retention               bool
	retentionPolicy         int64
	retentionPeriod         int
	
	// Handler composition
	serverAdapter        *ServerAdapter
	userHandlers         *user.Handlers
	colonyHandlers       *colony.Handlers
	executorHandlers     *executor.Handlers
	processHandlers      *process.Handlers
	processgraphHandlers *processgraph.Handlers
	serverHandlers       *serverhandlers.Handlers
	logHandlers          *loghandlers.Handlers
	snapshotHandlers     *snapshothandlers.Handlers
	attributeHandlers    *attributehandlers.Handlers
	cronHandlers         *cronhandlers.Handlers
	functionHandlers     *functionhandlers.Handlers
	generatorHandlers    *generatorhandlers.Handlers
	securityHandlers     *securityhandlers.Handlers
	fileHandlers         *filehandlers.Handlers
	websocketHandlers    *websockethandlers.Handlers
}

func CreateColoniesServer(db database.Database,
	port int,
	tls bool,
	tlsPrivateKeyPath string,
	tlsCertPath string,
	thisNode cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	generatorPeriod int,
	cronPeriod int,
	exclusiveAssign bool,
	allowExecutorReregister bool,
	retention bool,
	retentionPolicy int64,
	retentionPeriod int) *ColoniesServer {
	server := &ColoniesServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())

	// Set all the specific database interfaces
	server.userDB = db
	server.colonyDB = db
	server.executorDB = db
	server.functionDB = db
	server.processDB = db
	server.attributeDB = db
	server.processGraphDB = db
	server.generatorDB = db
	server.cronDB = db
	server.logDB = db
	server.fileDB = db
	server.snapshotDB = db
	server.securityDB = db

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: server.ginHandler,
	}

	server.httpServer = httpServer
	server.controller = createColoniesController(db, thisNode, clusterConfig, etcdDataPath, generatorPeriod, cronPeriod, retention, retentionPolicy, retentionPeriod)

	server.tls = tls
	server.port = port
	server.tlsPrivateKeyPath = tlsPrivateKeyPath
	server.tlsCertPath = tlsCertPath
	server.crypto = crypto.CreateCrypto()
	server.validator = validator.CreateValidator(db)
	server.exclusiveAssign = exclusiveAssign
	server.allowExecutorReregister = allowExecutorReregister
	server.retention = retention
	server.retentionPolicy = retentionPolicy

	// Initialize server adapter and handler structs
	server.serverAdapter = NewServerAdapter(server)
	server.userHandlers = user.NewHandlers(server.serverAdapter)
	server.colonyHandlers = colony.NewHandlers(server.serverAdapter)
	server.executorHandlers = executor.NewHandlers(server.serverAdapter)
	server.processHandlers = process.NewHandlers(server.serverAdapter)
	server.processgraphHandlers = processgraph.NewHandlers(server.serverAdapter.ProcessgraphServer())
	server.serverHandlers = serverhandlers.NewHandlers(server.serverAdapter.ServerServer())
	server.logHandlers = loghandlers.NewHandlers(server.serverAdapter)
	server.snapshotHandlers = snapshothandlers.NewHandlers(server.serverAdapter)
	server.attributeHandlers = attributehandlers.NewHandlers(server.serverAdapter)
	server.cronHandlers = cronhandlers.NewHandlers(server.serverAdapter)
	server.fileHandlers = filehandlers.NewHandlers(server.serverAdapter)
	server.functionHandlers = functionhandlers.NewHandlers(server.serverAdapter)
	server.generatorHandlers = generatorhandlers.NewHandlers(server.serverAdapter)
	server.securityHandlers = securityhandlers.NewHandlers(server.serverAdapter)
	server.websocketHandlers = websockethandlers.NewHandlers(server.serverAdapter)

	log.WithFields(log.Fields{"Port": port,
		"TLS":                     tls,
		"TLSPrivateKeyPath":       tlsPrivateKeyPath,
		"TLSCertPath":             tlsCertPath,
		"APIPort":                 thisNode.APIPort,
		"EtcdClientPort":          thisNode.EtcdClientPort,
		"EtcdPeerPort":            thisNode.EtcdPeerPort,
		"EtcdDataPath":            etcdDataPath,
		"Host":                    thisNode.Host,
		"RelayPort":               thisNode.RelayPort,
		"Name":                    thisNode.Name,
		"GeneratorPeriod":         generatorPeriod,
		"CronPeriod":              cronPeriod,
		"AllowExecutorReregister": allowExecutorReregister,
		"ExclusiveAssign":         exclusiveAssign,
		"Retention":               retention,
		"RetentionPolicy":         retentionPolicy}).
		Info("Starting Colonies server")

	server.setupRoutes()

	return server
}

func (server *ColoniesServer) SetAllowExecutorReregister(allow bool) {
	server.allowExecutorReregister = allow
}

func (server *ColoniesServer) getServerID() (string, error) {
	return server.securityDB.GetServerID()
}

func (server *ColoniesServer) setupRoutes() {
	server.ginHandler.POST("/api", server.handleAPIRequest)
	server.ginHandler.GET("/health", server.handleHealthRequest)
	server.ginHandler.GET("/pubsub", server.websocketHandlers.HandleWSRequest)
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
		server.serverHandlers.HandleVersion(c, rpcMsg.PayloadType, rpcMsg.DecodePayload())
		return
	}

	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	switch rpcMsg.PayloadType {

	// User handlers
	case rpc.AddUserPayloadType:
		server.userHandlers.HandleAddUser(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetUsersPayloadType:
		server.userHandlers.HandleGetUsers(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetUserPayloadType:
		server.userHandlers.HandleGetUser(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveUserPayloadType:
		server.userHandlers.HandleRemoveUser(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Colony handlers
	case rpc.AddColonyPayloadType:
		server.colonyHandlers.HandleAddColony(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveColonyPayloadType:
		server.colonyHandlers.HandleRemoveColony(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColoniesPayloadType:
		server.colonyHandlers.HandleGetColonies(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColonyPayloadType:
		server.colonyHandlers.HandleGetColony(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Executor handlers
	case rpc.AddExecutorPayloadType:
		server.executorHandlers.HandleAddExecutor(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetExecutorsPayloadType:
		server.executorHandlers.HandleGetExecutors(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetExecutorPayloadType:
		server.executorHandlers.HandleGetExecutor(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ApproveExecutorPayloadType:
		server.executorHandlers.HandleApproveExecutor(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RejectExecutorPayloadType:
		server.executorHandlers.HandleRejectExecutor(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveExecutorPayloadType:
		server.executorHandlers.HandleRemoveExecutor(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ReportAllocationsPayloadType:
		server.executorHandlers.HandleReportAllocations(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	//Function handlers
	case rpc.AddFunctionPayloadType:
		server.functionHandlers.HandleAddFunction(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetFunctionsPayloadType:
		server.functionHandlers.HandleGetFunctions(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveFunctionPayloadType:
		server.functionHandlers.HandleRemoveFunction(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Process handlers
	case rpc.SubmitFunctionSpecPayloadType:
		server.processHandlers.HandleSubmit(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.AssignProcessPayloadType:
		server.processHandlers.HandleAssignProcess(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload(), string(jsonBytes))
	case rpc.PauseAssignmentsPayloadType:
		server.processHandlers.HandlePauseAssignments(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ResumeAssignmentsPayloadType:
		server.processHandlers.HandleResumeAssignments(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetPauseStatusPayloadType:
		server.processHandlers.HandleGetPauseStatus(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessHistPayloadType:
		server.processHandlers.HandleGetProcessHist(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessesPayloadType:
		server.processHandlers.HandleGetProcesses(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessPayloadType:
		server.processHandlers.HandleGetProcess(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveProcessPayloadType:
		server.processHandlers.HandleRemoveProcess(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveAllProcessesPayloadType:
		server.processHandlers.HandleRemoveAllProcesses(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.CloseSuccessfulPayloadType:
		server.processHandlers.HandleCloseSuccessful(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.CloseFailedPayloadType:
		server.processHandlers.HandleCloseFailed(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.SetOutputPayloadType:
		server.processHandlers.HandleSetOutput(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColonyStatisticsPayloadType:
		server.colonyHandlers.HandleColonyStatistics(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Attribute handlers
	case rpc.AddAttributePayloadType:
		server.attributeHandlers.HandleAddAttribute(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetAttributePayloadType:
		server.attributeHandlers.HandleGetAttribute(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Workflow and processgraph handlers
	case rpc.SubmitWorkflowSpecPayloadType:
		server.processgraphHandlers.HandleSubmitWorkflow(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessGraphPayloadType:
		server.processgraphHandlers.HandleGetProcessGraph(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessGraphsPayloadType:
		server.processgraphHandlers.HandleGetProcessGraphs(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveProcessGraphPayloadType:
		server.processgraphHandlers.HandleRemoveProcessGraph(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveAllProcessGraphsPayloadType:
		server.processgraphHandlers.HandleRemoveAllProcessGraphs(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.AddChildPayloadType:
		server.processgraphHandlers.HandleAddChild(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Generators handlers
	case rpc.AddGeneratorPayloadType:
		server.generatorHandlers.HandleAddGenerator(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetGeneratorPayloadType:
		server.generatorHandlers.HandleGetGenerator(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ResolveGeneratorPayloadType:
		server.generatorHandlers.HandleResolveGenerator(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetGeneratorsPayloadType:
		server.generatorHandlers.HandleGetGenerators(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.PackGeneratorPayloadType:
		server.generatorHandlers.HandlePackGenerator(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveGeneratorPayloadType:
		server.generatorHandlers.HandleRemoveGenerator(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Cron handlers
	case rpc.AddCronPayloadType:
		server.cronHandlers.HandleAddCron(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetCronPayloadType:
		server.cronHandlers.HandleGetCron(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetCronsPayloadType:
		server.cronHandlers.HandleGetCrons(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RunCronPayloadType:
		server.cronHandlers.HandleRunCron(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveCronPayloadType:
		server.cronHandlers.HandleRemoveCron(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Server handlers
	case rpc.GetStatisiticsPayloadType:
		server.serverHandlers.HandleStatistics(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetClusterPayloadType:
		server.serverHandlers.HandleGetCluster(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Log handlers
	case rpc.AddLogPayloadType:
		server.logHandlers.HandleAddLog(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetLogsPayloadType:
		server.logHandlers.HandleGetLogs(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.SearchLogsPayloadType:
		server.logHandlers.HandleSearchLogs(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

		// File handlers
	case rpc.AddFilePayloadType:
		server.fileHandlers.HandleAddFile(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetFilePayloadType:
		server.fileHandlers.HandleGetFile(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetFilesPayloadType:
		server.fileHandlers.HandleGetFiles(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetFileLabelsPayloadType:
		server.fileHandlers.HandleGetFileLabels(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveFilePayloadType:
		server.fileHandlers.HandleRemoveFile(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

		// Snapshot handlers
	case rpc.CreateSnapshotPayloadType:
		server.snapshotHandlers.HandleCreateSnapshot(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetSnapshotPayloadType:
		server.snapshotHandlers.HandleGetSnapshot(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetSnapshotsPayloadType:
		server.snapshotHandlers.HandleGetSnapshots(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveSnapshotPayloadType:
		server.snapshotHandlers.HandleRemoveSnapshot(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RemoveAllSnapshotsPayloadType:
		server.snapshotHandlers.HandleRemoveAllSnapshots(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

		// Security handlers
	case rpc.ChangeUserIDPayloadType:
		server.securityHandlers.HandleChangeUserID(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ChangeExecutorIDPayloadType:
		server.securityHandlers.HandleChangeExecutorID(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ChangeColonyIDPayloadType:
		server.securityHandlers.HandleChangeColonyID(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ChangeServerIDPayloadType:
		server.securityHandlers.HandleChangeServerID(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

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
		if !strings.HasPrefix(err.Error(), "No processes can be selected for executor with Id") {
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

func (server *ColoniesServer) FileDB() database.FileDatabase {
	return server.fileDB
}

func (server *ColoniesServer) Shutdown() {
	server.controller.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("ColoniesServer forced to shutdown")
	}
}
