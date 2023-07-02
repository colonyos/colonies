package cli

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/core"
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
