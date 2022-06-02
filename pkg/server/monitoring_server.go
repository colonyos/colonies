package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type MonitoringServer struct {
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "colonies_processes",
		Help: "The total number of Colonies processes",
	})
)

func CreateMonitoringServer(port int) *MonitoringServer {
	server := &MonitoringServer{}
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)

	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()

	log.WithFields(log.Fields{"Port": port}).Info("Starting Monitoring server")

	return server
}
