package cli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/client/libp2p"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/validate"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Register the LibP2P backend factory
	client.RegisterBackendFactory(libp2p.NewLibP2PClientBackendFactory())
}

func StrArr2Str(args []string) string {
	if len(args) == 0 {
		return ""
	}

	str := ""
	for _, arg := range args {
		str += arg + " "
	}

	return str[0 : len(str)-1]
}

func StrMap2Str(args map[string]string) string {
	if len(args) == 0 {
		return ""
	}

	str := ""
	for k, arg := range args {
		str += k + ":" + arg + " "
	}

	return str[0 : len(str)-1]
}

func StrArr2StrWithCommas(args []string) string {
	if len(args) == 0 {
		return ""
	}

	str := ""
	for _, arg := range args {
		str += arg + ","
	}

	return str[0 : len(str)-1]
}

func parseArgs(process *core.Process) (string, string) {
	args := StrArr2Str(IfArr2StringArr(process.FunctionSpec.Args))
	kwArgs := StrMap2Str(IfMap2StringMap(process.FunctionSpec.KwArgs))

	if len(args) > MaxArgLength {
		args = args[0:MaxArgLength] + "..."
	}

	if len(kwArgs) > MaxArgLength {
		kwArgs = kwArgs[0:MaxArgLength] + "..."
	}

	return args, kwArgs
}

func IfArr2StringArr(ifarr []interface{}) []string {
	strarr := make([]string, len(ifarr))
	for k, v := range ifarr {
		strarr[k] = fmt.Sprint(v)
	}

	return strarr
}

func IfMap2StringMap(ifarr map[string]interface{}) map[string]string {
	strarr := make(map[string]string)
	for k, v := range ifarr {
		strarr[k] = fmt.Sprint(v)
	}

	return strarr
}

func State2String(state int) string {
	var stateStr string
	switch state {
	case core.WAITING:
		stateStr = "Waiting"
	case core.RUNNING:
		stateStr = "Running"
	case core.SUCCESS:
		stateStr = "Successful"
	case core.FAILED:
		stateStr = "Failed"
	default:
		stateStr = "Unkown"
	}

	return stateStr
}

func CheckError(err error) {
	if err != nil {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Error(err.Error())
		os.Exit(-1)
	}
}

func CheckJSONParseErr(err error, jsonStr string) {
	if err != nil {
		jsonErrStr, err := validate.JSON(err, jsonStr, true)
		CheckError(err)
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Error(jsonErrStr)
		os.Exit(-1)
	}
}

func setupProfiler() {
	profilerStr := os.Getenv("COLONIES_SERVER_PROFILER")
	profiler := false
	if profilerStr == "true" {
		profiler = true
	}

	if profiler {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}

	profilerPortStr := os.Getenv("COLONIES_SERVER_PROFILER_PORT")
	var err error
	if profilerPortStr != "" {
		_, err = strconv.Atoi(profilerPortStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_SERVER_PROFILER_PORT")
		}
		CheckError(err)
	}

	if profiler {
		go func() {
			log.WithFields(log.Fields{"ProfilerPort": profilerPortStr}).Info("Enabling profiler")
			http.ListenAndServe(":"+profilerPortStr, nil)
		}()
	}
}

func parseEnv() {
	var err error
	ASCII = false
	ASCIIStr := os.Getenv("COLONIES_CLI_ASCII")
	if ASCIIStr == "true" {
		ASCII = true
	}

	ServerHostEnv := os.Getenv("COLONIES_SERVER_HOST")
	if ServerHostEnv != "" {
		ServerHost = ServerHostEnv
	}

	// Try new variable COLONIES_SERVER_HTTP_PORT first, fall back to legacy COLONIES_SERVER_PORT
	ServerPortEnvStr := os.Getenv("COLONIES_SERVER_HTTP_PORT")
	if ServerPortEnvStr == "" {
		ServerPortEnvStr = os.Getenv("COLONIES_SERVER_PORT") // Backward compatibility
	}
	if ServerPortEnvStr != "" {
		if ServerPort == -1 {
			ServerPort, err = strconv.Atoi(ServerPortEnvStr)
			if err != nil {
				log.Error("Failed to parse COLONIES_SERVER_HTTP_PORT/COLONIES_SERVER_PORT")
			}
			CheckError(err)
		}
	}

	LibP2PPortEnvStr := os.Getenv("COLONIES_LIBP2P_PORT")
	if LibP2PPortEnvStr != "" {
		LibP2PPort, err = strconv.Atoi(LibP2PPortEnvStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_LIBP2P_PORT")
		}
		CheckError(err)
	}

	// COLONIES_TLS controls whether to use HTTPS (true) or HTTP (false)
	TLSEnv := os.Getenv("COLONIES_TLS")
	if TLSEnv == "true" {
		UseTLS = true
		Insecure = false
	} else if TLSEnv == "false" {
		UseTLS = false
		Insecure = true
	}

	if TLSKey == "" {
		TLSKey = os.Getenv("COLONIES_SERVER_TLSKEY")
	}

	if TLSCert == "" {
		TLSCert = os.Getenv("COLONIES_SERVER_TLSCERT")
	}

	VerboseEnv := os.Getenv("COLONIES_VERBOSE")
	if VerboseEnv == "true" {
		Verbose = true
	} else if VerboseEnv == "false" {
		Verbose = false
	}

	if Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = ioutil.Discard
	}

	CronPeriodCheckerEnvStr := os.Getenv("COLONIES_CRON_CHECKER_PERIOD")
	if CronPeriodCheckerEnvStr != "" {
		CronCheckerPeriod, err = strconv.Atoi(CronPeriodCheckerEnvStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_CRON_CHECKER_PERIOD")
		}
		CheckError(err)
	} else {
		CronCheckerPeriod = constants.CRON_TRIGGER_PERIOD
	}

	GeneratorPeriodCheckerEnvStr := os.Getenv("COLONIES_GENERATOR_CHECKER_PERIOD")
	if GeneratorPeriodCheckerEnvStr != "" {
		GeneratorCheckerPeriod, err = strconv.Atoi(GeneratorPeriodCheckerEnvStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_GENERATOR_CHECKER_PERIOD")
		}
		CheckError(err)
	} else {
		GeneratorCheckerPeriod = constants.GENERATOR_TRIGGER_PERIOD
	}

	ExclusiveAssignEnvStr := os.Getenv("COLONIES_EXCLUSIVE_ASSIGN")
	if ExclusiveAssignEnvStr != "" {
		ExclusiveAssign, err = strconv.ParseBool(ExclusiveAssignEnvStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_EXCLUSIVE_ASSIGN")
		}
		CheckError(err)
	} else {
		ExclusiveAssign = false
	}

	DBHost = os.Getenv("COLONIES_DB_HOST")

	DBPortStr := os.Getenv("COLONIES_DB_PORT")
	if DBPortStr != "" {
		DBPort, err = strconv.Atoi(DBPortStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_DB_PORT")
		}
		CheckError(err)
	}

	DBUser = os.Getenv("COLONIES_DB_USER")
	DBPassword = os.Getenv("COLONIES_DB_PASSWORD")

	AllowExecutorReregisterStr := os.Getenv("COLONIES_ALLOW_EXECUTOR_REREGISTER")
	if AllowExecutorReregisterStr != "" {
		AllowExecutorReregister, err = strconv.ParseBool(AllowExecutorReregisterStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_ALLOW_EXECUTOR_REREGISTER")
		}
		CheckError(err)
	} else {
		AllowExecutorReregister = false
	}

	timescaleDBEnv := os.Getenv("COLONIES_DB_TIMESCALEDB")
	if timescaleDBEnv == "true" {
		TimescaleDB = true
	} else {
		TimescaleDB = false
	}

	if ColonyName == "" {
		ColonyName = os.Getenv("COLONIES_COLONY_NAME")
	}

	if ColonyName == "" {
		ColonyName = os.Getenv("COLONIES_COLONY_NAME")
	}

	ServerID = os.Getenv("COLONIES_SERVER_ID")
	ServerPrvKey = os.Getenv("COLONIES_SERVER_PRVKEY")
	ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")

	if PrvKey == "" {
		PrvKey = os.Getenv("COLONIES_PRVKEY")
	}

	if ExecutorType == "" {
		ExecutorType = os.Getenv("COLONIES_EXECUTOR_TYPE")
	}

	if ExecutorName == "" {
		ExecutorName = os.Getenv("COLONIES_EXECUTOR_NAME")
	}

	retentionStr := os.Getenv("COLONIES_RETENTION")
	Retention = false
	if retentionStr == "true" {
		Retention = true
	}
	retentionPolicyStr := os.Getenv("COLONIES_RETENTION_POLICY")
	if retentionPolicyStr != "" {
		RetentionPolicy, err = strconv.ParseInt(retentionPolicyStr, 10, 64)
		if err != nil {
			log.Error("Failed to parse COLONIES_RETENTION_POLICY")
		}
		CheckError(err)
	}

	monitorPortStr := os.Getenv("COLONIES_MONITOR_PORT")
	if monitorPortStr != "" {
		MonitorPort, err = strconv.Atoi(monitorPortStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_MONITOR_PORT")
		}
		CheckError(err)
	}

	intervalStr := os.Getenv("COLONIES_MONITOR_INTERVAL")
	if intervalStr != "" {
		MonitorInterval, err = strconv.Atoi(intervalStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_MONITOR_INTERVAL")
		}
		CheckError(err)
	}
}

func checkDevEnv() {
	envErr := false
	if os.Getenv("LANG") == "" {
		log.Error("LANG environmental variable missing, try export LANG=en_US.UTF-8")
		envErr = true
	}

	if os.Getenv("LANGUAGE") == "" {
		log.Error("LANGUAGE environmental variable missing, try export LANGUAGE=en_US.UTF-8")
		envErr = true
	}

	if os.Getenv("LC_ALL") == "" {
		log.Error("LC_ALL environmental variable missing, try export LC_ALL=en_US.UTF-8")
		envErr = true
	}

	if os.Getenv("LC_CTYPE") == "" {
		log.Error("LC_CTYPE environmental variable missing, try export LC_CTYPE=UTF-8")
		envErr = true
	}

	if os.Getenv("TZ") == "" {
		log.Error("TZ environmental variable missing, try export TZ=Europe/Stockholm")
		envErr = true
	}

	if os.Getenv("COLONIES_SERVER_HOST") == "" {
		log.Error("COLONIES_SERVER_HOST environmental variable missing, try export COLONIES_SERVER_HOST=\"localhost\"")
		envErr = true
	}

	if os.Getenv("COLONIES_SERVER_PORT") == "" {
		log.Error("COLONIES_SERVER_PORT environmental variable missing, try export COLONIES_SERVER_PORT=\"50080\"")
		envErr = true
	}

	if os.Getenv("COLONIES_MONITOR_PORT") == "" {
		log.Error("COLONIES_MONITOR_PORT environmental variable missing, try export COLONIES_MONITOR_PORT=\"21120\"")
		envErr = true
	}

	if os.Getenv("COLONIES_MONITOR_INTERVAL") == "" {
		log.Error("COLONIES_MONITOR_INTERVAL environmental variable missing, try export COLONIES_MONITOR_INTERVAL=\"1\"")
		envErr = true
	}

	if os.Getenv("COLONIES_SERVER_PRVKEY") == "" {
		log.Error("COLONIES_SERVER_PRVKEY environmental variable missing, try export COLONIES_SERVER_PRVKEY=\"fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d\"")
		envErr = true
	}

	if os.Getenv("COLONIES_DB_HOST") == "" {
		log.Error("COLONIES_DB_HOST environmental variable missing, try export COLONIES_DB_HOST=\"localhost\"")
		envErr = true
	}

	if os.Getenv("COLONIES_DB_PORT") == "" {
		log.Error("COLONIES_DB_PORT environmental variable missing, try export COLONIES_DB_PORT=\"50070\"")
		envErr = true
	}

	if os.Getenv("COLONIES_DB_USER") == "" {
		log.Error("COLONIES_DB_USER environmental variable missing, try export COLONIES_DB_USER=\"postgres\"")
		envErr = true
	}

	if os.Getenv("COLONIES_DB_PASSWORD") == "" {
		log.Error("COLONIES_DB_PASSWORD environmental variable missing, try export COLONIES_DB_PASSWORD=\"rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7\"")
		envErr = true
	}

	if os.Getenv("COLONIES_COLONY_NAME") == "" {
		log.Error("COLONIES_COLONY_NAME environmental variable missing, try export COLONIES_COLONY_NAME=\"dev\"")
		envErr = true
	}

	if os.Getenv("COLONIES_COLONY_PRVKEY") == "" {
		log.Error("COLONIES_COLONY_PRVKEY environmental variable missing, try export COLONIES_COLONY_PRVKEY=\"ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514\"")
		envErr = true
	}

	if os.Getenv("COLONIES_PRVKEY") == "" {
		log.Error("COLONIES_PRVKEY environmental variable missing, try export COLONIES_PRVKEY=\"ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05\"")
		envErr = true
	}

	if os.Getenv("COLONIES_EXECUTOR_TYPE") == "" {
		log.Error("COLONIES_EXECUTOR_TYPE environmental variable missing, try export COLONIES_EXECUTOR_TYPE=\"cli\"")
		envErr = true
	}

	if envErr {
		log.Error(envErr)
		fmt.Println("\nExample of enironmental variables:")
		envProposal := "export LANG=en_US.UTF-8\n"
		envProposal += "export LANGUAGE=en_US.UTF-8\n"
		envProposal += "export LC_ALL=en_US.UTF-8\n"
		envProposal += "export LC_CTYPE=UTF-8\n"
		envProposal += "export TZ=Europe/Stockholm\n"
		envProposal += "export COLONIES_SERVER_TLS=\"false\"\n"
		envProposal += "export COLONIES_SERVER_HOST=\"localhost\"\n"
		envProposal += "export COLONIES_SERVER_PORT=\"50080\"\n"
		envProposal += "export COLONIES_MONITOR_PORT=\"21120\"\n"
		envProposal += "export COLONIES_SERVER_PRVKEY=\"fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d\"\n"
		envProposal += "export COLONIES_DB_HOST=\"localhost\"\n"
		envProposal += "export COLONIES_DB_USER=\"postgres\"\n"
		envProposal += "export COLONIES_DB_PORT=\"50070\"\n"
		envProposal += "export COLONIES_DB_PASSWORD=\"rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7\"\n"
		envProposal += "export COLONIES_COLONY_NAME=\"dev\"\n"
		envProposal += "export COLONIES_COLONY_PRVKEY=\"ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514\"\n"
		envProposal += "export COLONIES_PRVKEY=\"ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05\"\n"
		envProposal += "export COLONIES_EXECUTOR_TYPE=\"cli\"\n"

		fmt.Println(envProposal)
		os.Exit(-1)
	}
}

func envError() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Missing required environment variables. Did you forget to source your env file?")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  source docker-compose.env")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Required variables:")
	fmt.Fprintln(os.Stderr, "  COLONIES_SERVER_HOST   - Colonies server hostname")
	fmt.Fprintln(os.Stderr, "  COLONIES_SERVER_PORT   - Colonies server port (e.g., 50080)")
	fmt.Fprintln(os.Stderr, "  COLONIES_COLONY_NAME   - Name of the colony")
	fmt.Fprintln(os.Stderr, "  COLONIES_PRVKEY        - Your private key for authentication")
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

func setup() *client.ColoniesClient {
	parseEnv()

	missingVars := []string{}
	if ServerHost == "" {
		missingVars = append(missingVars, "COLONIES_SERVER_HOST")
	}
	if ColonyName == "" {
		missingVars = append(missingVars, "COLONIES_COLONY_NAME")
	}
	if PrvKey == "" {
		missingVars = append(missingVars, "COLONIES_PRVKEY")
	}

	if len(missingVars) > 0 {
		for _, v := range missingVars {
			log.Errorf("%s not set", v)
		}
		envError()
	}

	// Check for multiple backends (new format: COLONIES_CLIENT_BACKENDS="libp2p,http,grpc")
	backendsEnv := os.Getenv("COLONIES_CLIENT_BACKENDS")
	if backendsEnv != "" {
		backendTypes := backends.ParseClientBackendsFromEnv(backendsEnv)
		configs := make([]*backends.ClientConfig, 0, len(backendTypes))

		for _, backendType := range backendTypes {
			switch backendType {
			case backends.LibP2PClientBackendType:
				// LibP2P client configuration - completely separate from server
				libp2pHost := os.Getenv("COLONIES_CLIENT_LIBP2P_HOST")
				if libp2pHost == "" {
					libp2pHost = os.Getenv("COLONIES_SERVER_HOST") // Backward compatibility
					if libp2pHost == "" {
						libp2pHost = ServerHost
					}
				}
				log.WithFields(log.Fields{"Host": libp2pHost, "Backend": "libp2p"}).Debug("Adding LibP2P backend")
				config := backends.CreateLibP2PClientConfig(libp2pHost)

				// LibP2P-specific bootstrap peers
				bootstrapPeers := os.Getenv("COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS")
				if bootstrapPeers == "" {
					bootstrapPeers = os.Getenv("COLONIES_LIBP2P_BOOTSTRAP_PEERS") // Backward compatibility
				}
				config.BootstrapPeers = bootstrapPeers
				configs = append(configs, config)

			case backends.GinClientBackendType:
				// HTTP client configuration - completely separate from server
				httpHost := os.Getenv("COLONIES_CLIENT_HTTP_HOST")
				if httpHost == "" {
					httpHost = os.Getenv("COLONIES_SERVER_HOST") // Backward compatibility
					if httpHost == "" {
						httpHost = ServerHost
					}
				}

				httpPort := 50080 // Default HTTP port
				httpPortStr := os.Getenv("COLONIES_CLIENT_HTTP_PORT")
				if httpPortStr != "" {
					var err error
					httpPort, err = strconv.Atoi(httpPortStr)
					if err != nil {
						log.WithError(err).Error("Failed to parse COLONIES_CLIENT_HTTP_PORT, using default 50080")
						httpPort = 50080
					}
				} else {
					// Backward compatibility: try old variable names
					httpPortStr = os.Getenv("COLONIES_SERVER_HTTP_PORT")
					if httpPortStr != "" {
						var err error
						httpPort, err = strconv.Atoi(httpPortStr)
						if err != nil {
							log.WithError(err).Error("Failed to parse COLONIES_SERVER_HTTP_PORT, using default 50080")
							httpPort = 50080
						}
					} else if ServerPort != -1 {
						httpPort = ServerPort
					}
				}

				// HTTP-specific TLS settings
				httpInsecure := Insecure
				httpInsecureStr := os.Getenv("COLONIES_CLIENT_HTTP_INSECURE")
				if httpInsecureStr != "" {
					httpInsecure = (httpInsecureStr == "true")
				}

				httpSkipTLSVerify := SkipTLSVerify
				httpSkipTLSVerifyStr := os.Getenv("COLONIES_CLIENT_HTTP_SKIP_TLS_VERIFY")
				if httpSkipTLSVerifyStr != "" {
					httpSkipTLSVerify = (httpSkipTLSVerifyStr == "true")
				}

				log.WithFields(log.Fields{"Host": httpHost, "Port": httpPort, "Backend": "http"}).Debug("Adding HTTP backend")
				config := backends.CreateDefaultClientConfig(httpHost, httpPort, httpInsecure, httpSkipTLSVerify)
				configs = append(configs, config)

			case backends.GRPCClientBackendType:
				// gRPC client configuration - completely separate from server
				grpcHost := os.Getenv("COLONIES_CLIENT_GRPC_HOST")
				if grpcHost == "" {
					grpcHost = os.Getenv("COLONIES_SERVER_HOST") // Backward compatibility
					if grpcHost == "" {
						grpcHost = ServerHost
					}
				}

				grpcPortStr := os.Getenv("COLONIES_CLIENT_GRPC_PORT")
				if grpcPortStr == "" {
					// Backward compatibility
					grpcPortStr = os.Getenv("COLONIES_SERVER_GRPC_PORT")
				}
				if grpcPortStr == "" {
					log.Error("COLONIES_CLIENT_GRPC_PORT must be set when using gRPC backend")
					continue
				}
				grpcPort, err := strconv.Atoi(grpcPortStr)
				if err != nil {
					log.WithError(err).Error("Failed to parse COLONIES_CLIENT_GRPC_PORT")
					continue
				}

				// gRPC-specific TLS settings
				grpcInsecure := Insecure
				grpcInsecureStr := os.Getenv("COLONIES_CLIENT_GRPC_INSECURE")
				if grpcInsecureStr != "" {
					grpcInsecure = (grpcInsecureStr == "true")
				}

				grpcSkipTLSVerify := SkipTLSVerify
				grpcSkipTLSVerifyStr := os.Getenv("COLONIES_CLIENT_GRPC_SKIP_TLS_VERIFY")
				if grpcSkipTLSVerifyStr != "" {
					grpcSkipTLSVerify = (grpcSkipTLSVerifyStr == "true")
				}

				log.WithFields(log.Fields{"Host": grpcHost, "Port": grpcPort, "Backend": "grpc"}).Debug("Adding gRPC backend")
				config := backends.CreateGRPCClientConfig(grpcHost, grpcPort, grpcInsecure, grpcSkipTLSVerify)
				configs = append(configs, config)

			case backends.CoAPClientBackendType:
				// CoAP client configuration - completely separate from server
				coapHost := os.Getenv("COLONIES_CLIENT_COAP_HOST")
				if coapHost == "" {
					coapHost = os.Getenv("COLONIES_SERVER_HOST") // Backward compatibility
					if coapHost == "" {
						coapHost = ServerHost
					}
				}

				coapPortStr := os.Getenv("COLONIES_CLIENT_COAP_PORT")
				if coapPortStr == "" {
					// Backward compatibility
					coapPortStr = os.Getenv("COLONIES_SERVER_COAP_PORT")
				}
				if coapPortStr == "" {
					log.Error("COLONIES_CLIENT_COAP_PORT must be set when using CoAP backend")
					continue
				}
				coapPort, err := strconv.Atoi(coapPortStr)
				if err != nil {
					log.WithError(err).Error("Failed to parse COLONIES_CLIENT_COAP_PORT")
					continue
				}

				// CoAP-specific TLS settings (though CoAP typically uses DTLS)
				coapInsecure := Insecure
				coapInsecureStr := os.Getenv("COLONIES_CLIENT_COAP_INSECURE")
				if coapInsecureStr != "" {
					coapInsecure = (coapInsecureStr == "true")
				}

				coapSkipTLSVerify := SkipTLSVerify
				coapSkipTLSVerifyStr := os.Getenv("COLONIES_CLIENT_COAP_SKIP_TLS_VERIFY")
				if coapSkipTLSVerifyStr != "" {
					coapSkipTLSVerify = (coapSkipTLSVerifyStr == "true")
				}

				log.WithFields(log.Fields{"Host": coapHost, "Port": coapPort, "Backend": "coap"}).Debug("Adding CoAP backend")
				config := backends.CreateCoAPClientConfig(coapHost, coapPort, coapInsecure, coapSkipTLSVerify)
				configs = append(configs, config)
			}
		}

		if len(configs) > 0 {
			// Build detailed backend information for logging
			backendDetails := make([]string, 0, len(configs))
			for _, cfg := range configs {
				switch cfg.BackendType {
				case backends.GinClientBackendType:
					backendDetails = append(backendDetails, fmt.Sprintf("HTTP(%s:%d,insecure=%t)", cfg.Host, cfg.Port, cfg.Insecure))
				case backends.GRPCClientBackendType:
					backendDetails = append(backendDetails, fmt.Sprintf("gRPC(%s:%d,insecure=%t)", cfg.Host, cfg.Port, cfg.Insecure))
				case backends.LibP2PClientBackendType:
					backendDetails = append(backendDetails, fmt.Sprintf("LibP2P(%s)", cfg.Host))
				case backends.CoAPClientBackendType:
					backendDetails = append(backendDetails, fmt.Sprintf("CoAP(%s:%d,insecure=%t)", cfg.Host, cfg.Port, cfg.Insecure))
				}
			}
			log.WithFields(log.Fields{
				"backends": strings.Join(backendDetails, " → "),
				"count":    len(configs),
			}).Info("Creating multi-backend client with fallback chain")
			return client.CreateColoniesClientWithMultipleBackends(configs)
		}
	}

	// Fall back to legacy COLONIES_CLIENT_BACKEND for backward compatibility
	clientBackend := os.Getenv("COLONIES_CLIENT_BACKEND")
	if clientBackend == "libp2p" {
		log.WithFields(log.Fields{"ServerHost": ServerHost, "Backend": "libp2p"}).Debug("Starting a Colonies LibP2P client")
		// For LibP2P, ServerHost can be:
		// - A multiaddr like "/ip4/127.0.0.1/tcp/5000/p2p/12D3KooW..."
		// - "dht" for DHT-based discovery
		// - "dht:rendezvous-name" for custom rendezvous point
		config := backends.CreateLibP2PClientConfig(ServerHost)
		// Add bootstrap peers from environment if specified
		config.BootstrapPeers = os.Getenv("COLONIES_LIBP2P_BOOTSTRAP_PEERS")
		return client.CreateColoniesClientWithConfig(config)
	}

	// Default to HTTP/Gin backend
	log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies HTTP client")
	return client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
}

func insertNewLines(s string, interval int) string {
	var result strings.Builder
	count := 0

	// Split the string into words
	words := strings.Fields(s)

	for _, word := range words {
		wordLength := len(word)

		// If adding the next word exceeds the interval and count is not at the beginning of a new line, insert a newline
		if count+wordLength > interval && count > 0 {
			result.WriteString("\n")
			count = 0 // Reset count after inserting newline
		}

		// If it's not the beginning of the line, add a space before the word
		if count > 0 {
			result.WriteString(" ")
			count++ // Increment count for the space
		}

		// Add the word and update the count
		result.WriteString(word)
		count += wordLength
	}

	return result.String()
}

func formatTimestamp(timestamp string) string {
	return strings.Replace(timestamp, "T", " ", 1)
}
