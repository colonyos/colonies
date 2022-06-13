package server

import (
	"net/http"
	"strconv"

	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type MonitoringServer struct {
	coloniesClient         *client.ColoniesClient
	coloniesProcessesGauge prometheus.Gauge
	serverPrvKey           string
	coloniesServerHost     string
	coloniesServerPort     int
	insecure               bool
	skipTLSVerify          bool
	stat                   *core.Statistics
	pullIntervall          int
}

func CreateMonitoringServer(port int,
	coloniesServerHost string,
	coloniesServerPort int,
	insecure bool,
	skipTLSVerify bool,
	serverPrvKey string,
	pullIntervall int) *MonitoringServer {
	server := &MonitoringServer{}

	server.coloniesClient = client.CreateColoniesClient(coloniesServerHost, coloniesServerPort, insecure, skipTLSVerify)
	server.serverPrvKey = serverPrvKey
	server.coloniesServerHost = coloniesServerHost
	server.coloniesServerPort = coloniesServerPort
	server.insecure = insecure
	server.skipTLSVerify = skipTLSVerify
	server.pullIntervall = pullIntervall
	server.stat = nil

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":"+strconv.Itoa(port), nil)
	}()

	go func() {
		for {
			server.stat = server.fetchStat()
			time.Sleep(time.Duration(server.pullIntervall) * time.Second)
		}
	}()

	server.registerGauges()

	return server
}

func (server *MonitoringServer) fetchStat() *core.Statistics {
	stat, err := server.coloniesClient.Statistics(server.serverPrvKey)
	if err != nil {
		log.Error(err)
		return nil
	}

	return stat
}

func (server *MonitoringServer) registerGauges() {
	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_colonies",
			Help:      "Number of colonies",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.Colonies)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_runtimes",
			Help:      "Number of runtimes",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.Runtimes)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_processes_waiting",
			Help:      "Number of waiting processes",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.WaitingProcesses)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_processes_running",
			Help:      "Number of running processes",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.RunningProcesses)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_processes_successful",
			Help:      "Number of successful processes",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.SuccessfulProcesses)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_processes_failed",
			Help:      "Number of failed processes",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.FailedProcesses)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_workflows_waiting",
			Help:      "Number of waiting workflows",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.WaitingWorkflows)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_workflows_running",
			Help:      "Number of running workflows",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.RunningWorkflows)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_workflows_successful",
			Help:      "Number of successful workflows",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.SuccessfulWorkflows)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "colonies",
			Name:      "server_workflows_failed",
			Help:      "Number of failed workflows",
		},
		func() float64 {
			if server.stat == nil {
				log.WithFields(log.Fields{
					"ColoniesServerHost": server.coloniesServerHost,
					"ColoniesServerPort": server.coloniesServerPort,
					"Insecure":           server.insecure,
					"SkipTLSVerify":      server.skipTLSVerify}).
					Error("Failed to fetch Colonies server statistics (stat is nil)")
				return -1.0
			}
			return float64(server.stat.FailedWorkflows)
		},
	)); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register Prometheus metrics")
	}

	wait := make(chan struct{})
	<-wait
}
