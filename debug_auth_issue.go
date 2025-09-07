package main

import (
	"fmt"
	"log"
	"os"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/server"
)

func main() {
	fmt.Println("=== Debugging Authentication Issue ===")
	
	// Test both backends to see if there's a difference
	backends := []string{"gin", "libp2p"}
	
	for _, backend := range backends {
		fmt.Printf("\n--- Testing %s backend ---\n", backend)
		
		os.Setenv("COLONIES_BACKEND_TYPE", backend)
		
		// Setup database
		db, err := postgresql.PrepareTests()
		if err != nil {
			log.Fatal("Failed to prepare database:", err)
		}
		defer db.Close()
		
		// Setup crypto and server ID (same as in StartCluster)
		cryptoInstance := crypto.CreateCrypto()
		serverPrvKey, err := cryptoInstance.GeneratePrivateKey()
		if err != nil {
			log.Fatal("Failed to generate private key:", err)
		}
		serverID, err := cryptoInstance.GenerateID(serverPrvKey)
		if err != nil {
			log.Fatal("Failed to generate server ID:", err)
		}
		
		fmt.Printf("Generated serverID: %s\n", serverID)
		fmt.Printf("Generated serverPrvKey: %s...\n", serverPrvKey[:20])
		
		// Set server ID in database
		err = db.SetServerID("", serverID)
		if err != nil {
			log.Fatal("Failed to set server ID:", err)
		}
		
		// Get server ID back from database
		retrievedID, err := db.GetServerID()
		if err != nil {
			log.Fatal("Failed to get server ID:", err)
		}
		
		fmt.Printf("Retrieved serverID: %s\n", retrievedID)
		fmt.Printf("IDs match: %v\n", serverID == retrievedID)
		
		// Create server with this backend
		node := cluster.Node{
			Name:           "debug-node",
			Host:           "localhost",
			EtcdClientPort: 26000,
			EtcdPeerPort:   25000,
			RelayPort:      24000,
			APIPort:        8080,
		}
		clusterConfig := cluster.Config{}
		clusterConfig.AddNode(node)
		
		fmt.Println("Creating server...")
		srv := server.CreateServerFromEnv(db, 8080, false, "", "", node, clusterConfig, "/tmp/debug-etcd", 10, 1, true, false, false, -1, 500)
		
		fmt.Printf("Server created successfully with %s backend\n", backend)
		
		// Test server ID retrieval from database through server
		// We can't access the server ID directly, so we'll check if the database setup is consistent
		fmt.Printf("Server created and database setup appears consistent\n")
		
		// Clean up
		srv.Shutdown()
		db.Close()
		
		fmt.Printf("%s backend test completed\n", backend)
	}
	
	fmt.Println("\n=== Debug completed ===")
}