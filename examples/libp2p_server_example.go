package main

import (
	"log"
	"time"
	"os"
	"os/signal"
	"syscall"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/server"
)

func main() {
	// Set up database
	db, err := postgresql.PrepareTests()
	if err != nil {
		log.Fatal("Failed to prepare database:", err)
	}
	defer db.Close()

	// Create node and cluster config
	node := cluster.Node{
		Name:           "libp2p-node-1",
		Host:           "localhost", 
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        8080,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	// Create ServerManager
	serverManager := server.NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		10, // generator period
		1,  // cron period
	)

	// Register LibP2P backend factory
	libp2pFactory := server.NewLibP2PBackendFactory()
	err = serverManager.RegisterBackendFactory(libp2pFactory)
	if err != nil {
		log.Fatal("Failed to register LibP2P factory:", err)
	}

	// Add LibP2P server configuration
	libp2pConfig := &server.ServerConfig{
		BackendType:             server.LibP2PBackendType,
		Port:                   8080,
		TLS:                    false,
		ExclusiveAssign:        true,
		AllowExecutorReregister: false,
		Retention:              false,
		RetentionPolicy:        -1,
		RetentionPeriod:        500,
		Enabled:                true,
	}
	err = serverManager.AddServerConfig(libp2pConfig)
	if err != nil {
		log.Fatal("Failed to add LibP2P config:", err)
	}

	// Start all servers
	err = serverManager.StartAll()
	if err != nil {
		log.Fatal("Failed to start servers:", err)
	}

	log.Println("LibP2P server started successfully!")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	err = serverManager.StopAll(30 * time.Second)
	if err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}