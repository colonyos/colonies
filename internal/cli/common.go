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
	"github.com/colonyos/colonies/pkg/security"
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
		log.WithFields(log.Fields{"Error": err, "BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Error(err.Error())
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
			envError()
		}
	}

	if ServerID == "" {
		ServerID = os.Getenv("COLONIES_SERVER_ID")
	}

	TLSEnv := os.Getenv("COLONIES_TLS")
	if TLSEnv == "true" {
		UseTLS = true
		Insecure = false
	} else if TLSEnv == "false" {
		UseTLS = false
		Insecure = true
	}

	if TLSKey == "" {
		TLSKey = os.Getenv("COLONIES_TLSKEY")
	}

	if TLSCert == "" {
		TLSCert = os.Getenv("COLONIES_TLSCERT")
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
		CheckError(err)
	} else {
		CronCheckerPeriod = server.CRON_TRIGGER_PERIOD
	}

	GeneratorPeriodCheckerEnvStr := os.Getenv("COLONIES_GENERATOR_CHECKER_PERIOD")
	if GeneratorPeriodCheckerEnvStr != "" {
		GeneratorCheckerPeriod, err = strconv.Atoi(GeneratorPeriodCheckerEnvStr)
		CheckError(err)
	} else {
		GeneratorCheckerPeriod = server.GENERATOR_TRIGGER_PERIOD
	}

	ExclusiveAssignEnvStr := os.Getenv("COLONIES_EXCLUSIVE_ASSIGN")
	if ExclusiveAssignEnvStr != "" {
		ExclusiveAssign, err = strconv.ParseBool(ExclusiveAssignEnvStr)
		CheckError(err)
	} else {
		ExclusiveAssign = false
	}

	AllowExecutorReregisterStr := os.Getenv("COLONIES_ALLOW_EXECUTOR_REREGISTER")
	if AllowExecutorReregisterStr != "" {
		AllowExecutorReregister, err = strconv.ParseBool(AllowExecutorReregisterStr)
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

	if ServerID != "" {
		ServerID = os.Getenv("COLONIES_SERVER_ID")
	}

	if ColonyID == "" {
		ColonyID = os.Getenv("COLONIES_COLONY_ID")
	}

	if ExecutorID == "" {
		ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
	}

	keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
	CheckError(err)

	ServerPrvKey, err = keychain.GetPrvKey(ServerID)
	if err != nil {
		ServerPrvKey = os.Getenv("COLONIES_SERVER_PRVKEY")
	}

	ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
	if err != nil {
		ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")
	}

	ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
	if err != nil {
		ExecutorPrvKey = os.Getenv("COLONIES_EXECUTOR_PRVKEY")
	}

	if ExecutorType == "" {
		ExecutorType = os.Getenv("COLONIES_EXECUTOR_TYPE")
	}

	if ExecutorName == "" {
		ExecutorName = os.Getenv("COLONIES_EXECUTOR_NAME")
	}
}

func envError() {
	env := `export COLONIES_TLS="true"
export COLONIES_SERVER_TLS=""
export COLONIES_SERVER_HOST=""
export COLONIES_SERVER_PORT=""
export COLONIES_COLONY_ID=""
export COLONIES_EXECUTOR_ID=""
export COLONIES_EXECUTOR_PRVKEY=""
    `

	envAlt := `export COLONIES_TLS="true"
export COLONIES_SERVER_TLS=""
export COLONIES_SERVER_HOST=""
export COLONIES_SERVER_PORT=""
export COLONIES_COLONY_ID=""
export COLONIES_USER_ID=""
export COLONIES_USER_PRVKEY=""
    `

	log.Error("Please set the following environmental variable: \n\n" + env + "\nor alternatively:\n\n" + envAlt)
	os.Exit(-1)
}

func setup() *client.ColoniesClient {
	parseEnv()

	if (UserID == "" || UserPrvKey == "") || (ExecutorID == "" || ExecutorPrvKey == "") {
		envError()
	}

	if ColonyID == "" || ServerHost == "" {
		envError()
	}

	log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
	return client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
}
