package server

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func TestServerManagerCreation(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	node := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        constants.TESTPORT,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	sm := NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		constants.GENERATOR_TRIGGER_PERIOD,
		constants.CRON_TRIGGER_PERIOD,
	)

	assert.NotNil(t, sm)
	assert.False(t, sm.IsRunning())
}

func TestServerManagerBackendFactoryRegistration(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	node := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        constants.TESTPORT,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	sm := NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		constants.GENERATOR_TRIGGER_PERIOD,
		constants.CRON_TRIGGER_PERIOD,
	)

	// Register gin backend factory
	ginFactory := NewGinBackendFactory()
	err = sm.RegisterBackendFactory(ginFactory)
	assert.Nil(t, err)

	// Try to register the same factory again - should fail
	err = sm.RegisterBackendFactory(ginFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Register libp2p backend factory
	libp2pFactory := NewLibP2PBackendFactory()
	err = sm.RegisterBackendFactory(libp2pFactory)
	assert.Nil(t, err)

	// Cannot register after starting
	sm.running = true
	err = sm.RegisterBackendFactory(NewGinBackendFactory())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot register backend factory while server manager is running")
}

func TestServerManagerConfigManagement(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	node := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        constants.TESTPORT,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	sm := NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		constants.GENERATOR_TRIGGER_PERIOD,
		constants.CRON_TRIGGER_PERIOD,
	)

	// Add gin server config
	ginConfig := &ServerConfig{
		BackendType:             GinBackendType,
		Port:                   constants.TESTPORT + 100,
		TLS:                    false,
		TLSPrivateKeyPath:      "",
		TLSCertPath:            "",
		ExclusiveAssign:        true,
		AllowExecutorReregister: false,
		Retention:              false,
		RetentionPolicy:        -1,
		RetentionPeriod:        500,
		Enabled:                true,
	}

	err = sm.AddServerConfig(ginConfig)
	assert.Nil(t, err)

	// Try to add the same backend type again - should fail
	err = sm.AddServerConfig(ginConfig)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Add disabled config - should succeed but not be tracked
	disabledConfig := &ServerConfig{
		BackendType: LibP2PBackendType,
		Port:       constants.TESTPORT + 200,
		Enabled:    false,
	}
	err = sm.AddServerConfig(disabledConfig)
	assert.Nil(t, err)

	// Cannot add config after starting
	sm.running = true
	err = sm.AddServerConfig(&ServerConfig{
		BackendType: "test",
		Enabled:    true,
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot add server config while server manager is running")
}

func TestServerManagerLifecycle(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	err = db.SetServerID("", serverID)
	assert.Nil(t, err)

	node := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        constants.TESTPORT + 300,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	sm := NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		constants.GENERATOR_TRIGGER_PERIOD,
		constants.CRON_TRIGGER_PERIOD,
	)

	// Register factory and add config
	ginFactory := NewGinBackendFactory()
	err = sm.RegisterBackendFactory(ginFactory)
	assert.Nil(t, err)

	ginConfig := &ServerConfig{
		BackendType:             GinBackendType,
		Port:                   constants.TESTPORT + 300,
		TLS:                    false,
		TLSPrivateKeyPath:      "",
		TLSCertPath:            "",
		ExclusiveAssign:        true,
		AllowExecutorReregister: false,
		Retention:              false,
		RetentionPolicy:        -1,
		RetentionPeriod:        500,
		Enabled:                true,
	}

	err = sm.AddServerConfig(ginConfig)
	assert.Nil(t, err)

	// Start all servers
	err = sm.StartAll()
	assert.Nil(t, err)
	assert.True(t, sm.IsRunning())

	// Wait a moment for servers to start
	time.Sleep(200 * time.Millisecond)

	// Check that server is running
	runningServers := sm.GetRunningServers()
	assert.Len(t, runningServers, 1)
	assert.Contains(t, runningServers, GinBackendType)

	// Get server status
	status := sm.GetStatus()
	assert.Len(t, status, 1)
	ginStatus, exists := status[GinBackendType]
	assert.True(t, exists)
	assert.Equal(t, GinBackendType, ginStatus.BackendType)
	assert.True(t, ginStatus.Running)
	assert.Equal(t, constants.TESTPORT+300, ginStatus.Port)

	// Get specific server
	ginServer, exists := sm.GetServer(GinBackendType)
	assert.True(t, exists)
	assert.Equal(t, GinBackendType, ginServer.GetBackendType())
	assert.True(t, ginServer.IsRunning())

	// Try to start again - should fail
	err = sm.StartAll()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop all servers
	err = sm.StopAll(10 * time.Second)
	assert.Nil(t, err)
	assert.False(t, sm.IsRunning())

	// Check that no servers are running
	runningServers = sm.GetRunningServers()
	assert.Len(t, runningServers, 0)
}

func TestServerManagerMissingFactory(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	node := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        constants.TESTPORT + 400,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	sm := NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		constants.GENERATOR_TRIGGER_PERIOD,
		constants.CRON_TRIGGER_PERIOD,
	)

	// Add config without registering factory
	ginConfig := &ServerConfig{
		BackendType: GinBackendType,
		Port:       constants.TESTPORT + 400,
		Enabled:    true,
	}

	err = sm.AddServerConfig(ginConfig)
	assert.Nil(t, err)

	// Start should fail due to missing factory
	err = sm.StartAll()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no factory registered")
}

func TestServerManagerStopTimeout(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	err = db.SetServerID("", serverID)
	assert.Nil(t, err)

	node := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        constants.TESTPORT + 500,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	sm := NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		constants.GENERATOR_TRIGGER_PERIOD,
		constants.CRON_TRIGGER_PERIOD,
	)

	// Register factory and add config
	ginFactory := NewGinBackendFactory()
	err = sm.RegisterBackendFactory(ginFactory)
	assert.Nil(t, err)

	ginConfig := &ServerConfig{
		BackendType:             GinBackendType,
		Port:                   constants.TESTPORT + 500,
		TLS:                    false,
		TLSPrivateKeyPath:      "",
		TLSCertPath:            "",
		ExclusiveAssign:        true,
		AllowExecutorReregister: false,
		Retention:              false,
		RetentionPolicy:        -1,
		RetentionPeriod:        500,
		Enabled:                true,
	}

	err = sm.AddServerConfig(ginConfig)
	assert.Nil(t, err)

	// Start servers
	err = sm.StartAll()
	assert.Nil(t, err)

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Stop with very short timeout (this should still work since gin server shutdown is fast)
	err = sm.StopAll(1 * time.Millisecond)
	// Note: We don't assert the error here since the actual shutdown might be fast enough
	// The important thing is that it doesn't hang
}

func TestServerManagerHealthCheck(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	err = db.SetServerID("", serverID)
	assert.Nil(t, err)

	node := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		EtcdClientPort: 24100,
		EtcdPeerPort:   23100,
		RelayPort:      25100,
		APIPort:        constants.TESTPORT + 600,
	}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)

	sm := NewServerManager(
		db,
		node,
		clusterConfig,
		"/tmp/colonies/etcd",
		constants.GENERATOR_TRIGGER_PERIOD,
		constants.CRON_TRIGGER_PERIOD,
	)

	// Test health check routine
	sm.ctx, sm.cancel = context.WithCancel(context.Background())

	// Test perform health checks with no servers
	sm.performHealthChecks()

	// Register and start a server
	ginFactory := NewGinBackendFactory()
	err = sm.RegisterBackendFactory(ginFactory)
	assert.Nil(t, err)

	ginConfig := &ServerConfig{
		BackendType:             GinBackendType,
		Port:                   constants.TESTPORT + 600,
		TLS:                    false,
		TLSPrivateKeyPath:      "",
		TLSCertPath:            "",
		ExclusiveAssign:        true,
		AllowExecutorReregister: false,
		Retention:              false,
		RetentionPolicy:        -1,
		RetentionPeriod:        500,
		Enabled:                true,
	}

	err = sm.AddServerConfig(ginConfig)
	assert.Nil(t, err)

	err = sm.StartAll()
	assert.Nil(t, err)

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Test health checks with running server
	sm.performHealthChecks()

	// Cleanup
	err = sm.StopAll(5 * time.Second)
	assert.Nil(t, err)
}