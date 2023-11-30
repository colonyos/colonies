package cli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/validate"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

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
	ServerHostEnv := os.Getenv("COLONIES_SERVER_HOST")
	if ServerHostEnv != "" {
		ServerHost = ServerHostEnv
	}

	ServerPortEnvStr := os.Getenv("COLONIES_SERVER_PORT")
	if ServerPortEnvStr != "" {
		if ServerPort == -1 {
			ServerPort, err = strconv.Atoi(ServerPortEnvStr)
			if err != nil {
				log.Error("Failed to parse COLONIES_SERVER_PORT")
			}
			CheckError(err)
		}
	}

	TLSEnv := os.Getenv("COLONIES_SERVER_TLS")
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
		CronCheckerPeriod = server.CRON_TRIGGER_PERIOD
	}

	GeneratorPeriodCheckerEnvStr := os.Getenv("COLONIES_GENERATOR_CHECKER_PERIOD")
	if GeneratorPeriodCheckerEnvStr != "" {
		GeneratorCheckerPeriod, err = strconv.Atoi(GeneratorPeriodCheckerEnvStr)
		if err != nil {
			log.Error("Failed to parse COLONIES_GENERATOR_CHECKER_PERIOD")
		}
		CheckError(err)
	} else {
		GeneratorCheckerPeriod = server.GENERATOR_TRIGGER_PERIOD
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
	if DBHost != "" {
		DBPort, err = strconv.Atoi(os.Getenv("COLONIES_DB_PORT"))
		if err != nil {
			log.Error("COLONIES_DB_PORT")
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
		ColonyName = os.Getenv("COLONIES_COLONY_ID")
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
		envProposal += "export COLONIES_TLS=\"false\"\n"
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
	env := `export COLONIES_SERVER_TLS="true"
export COLONIES_SERVER_HOST=""
export COLONIES_SERVER_PORT=""
export COLONIES_COLONY_NAME=""
export COLONIES_PRVKEY=""
    `

	log.Error("Please set the following environmental variable: \n\n" + env)
	os.Exit(-1)
}

func setup() *client.ColoniesClient {
	parseEnv()

	if ColonyName == "" {
		log.Error("COLONIES_COLONY_NAME not set")
		envError()
	}

	if PrvKey == "" {
		log.Error("COLONIES_PRVKEY not set")
		envError()
	}

	if ServerHost == "" {
		log.Error("COLONIES_SERVER_HOST not set")
		envError()
	}

	if ColonyName == "" {
		log.Error("COLONIES_COLONY_NAME not set")
		envError()
	}

	log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
	return client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
}
