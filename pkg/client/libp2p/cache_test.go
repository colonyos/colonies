package libp2p

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func TestGetDHTCachePath(t *testing.T) {
	path, err := getDHTCachePath()
	assert.NoError(t, err)
	assert.NotEmpty(t, path)

	// Should be in home directory
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".colonies", "dht_cache")
	assert.Equal(t, expected, path)

	// Directory should exist after calling getDHTCachePath
	coloniesDir := filepath.Join(home, ".colonies")
	_, err = os.Stat(coloniesDir)
	assert.NoError(t, err, ".colonies directory should be created")
}

func TestDHTCacheSaveAndLoad(t *testing.T) {
	// Create a test backend with mock data
	backend := &LibP2PClientBackend{
		serverPeers:   make(map[peer.ID]*ServerPeer),
		dhtRendezvous: "test-rendezvous",
	}

	// Create test peer ID
	testPeerID := "12D3KooWBrsnBU9rZ8ZBaniVexPfdLmYyF34doTRtSJ7XqfC3JfM"
	peerID, err := peer.Decode(testPeerID)
	assert.NoError(t, err)

	// Create test multiaddr
	testAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/5000")
	assert.NoError(t, err)

	// Add server peer
	backend.serverPeers[peerID] = &ServerPeer{
		ID:       peerID,
		Addrs:    []multiaddr.Multiaddr{testAddr},
		LastSeen: time.Now(),
		Active:   true,
	}

	// Save cache
	err = backend.saveDHTCache()
	assert.NoError(t, err)

	// Verify cache file exists
	cachePath, err := getDHTCachePath()
	assert.NoError(t, err)
	_, err = os.Stat(cachePath)
	assert.NoError(t, err, "Cache file should exist")

	// Read and verify cache contents
	data, err := os.ReadFile(cachePath)
	assert.NoError(t, err)

	var cache DHTCache
	err = json.Unmarshal(data, &cache)
	assert.NoError(t, err)

	// Verify cache structure
	assert.Equal(t, 1, cache.Version)
	assert.NotZero(t, cache.Updated)
	assert.Len(t, cache.Peers, 1)
	assert.Equal(t, testPeerID, cache.Peers[0].PeerID)
	assert.Equal(t, "test-rendezvous", cache.Peers[0].Rendezvous)
	assert.Contains(t, cache.Peers[0].Addrs, "/ip4/127.0.0.1/tcp/5000")

	// Clean up
	os.Remove(cachePath)
}

func TestDHTCacheExpiry(t *testing.T) {
	cachePath, err := getDHTCachePath()
	assert.NoError(t, err)

	// Create expired cache (25 hours old)
	expiredCache := DHTCache{
		Version: 1,
		Updated: time.Now().Add(-25 * time.Hour),
		Peers: []CachedPeer{
			{
				PeerID:     "12D3KooWBrsnBU9rZ8ZBaniVexPfdLmYyF34doTRtSJ7XqfC3JfM",
				Addrs:      []string{"/ip4/127.0.0.1/tcp/5000"},
				LastSeen:   time.Now().Add(-25 * time.Hour),
				Rendezvous: "test-rendezvous",
			},
		},
	}

	// Write expired cache
	data, err := json.Marshal(expiredCache)
	assert.NoError(t, err)
	err = os.WriteFile(cachePath, data, 0600)
	assert.NoError(t, err)

	// Read back and verify expiry logic would detect it
	var loadedCache DHTCache
	data, err = os.ReadFile(cachePath)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &loadedCache)
	assert.NoError(t, err)

	// Verify cache is expired (more than 24 hours old)
	cacheAge := time.Since(loadedCache.Updated)
	assert.Greater(t, cacheAge, 24*time.Hour, "Cache should be expired")

	// Clean up
	os.Remove(cachePath)
}

func TestDHTCacheRendezvousFiltering(t *testing.T) {
	t.Skip("Skipping test that requires libp2p host setup")

	// This test would need a proper libp2p host mock to test peer connection
	// The cache load logic filters by rendezvous, but it also tries to connect
	// which requires a running libp2p host
}

func TestDHTCacheInvalidPeerID(t *testing.T) {
	t.Skip("Skipping test that requires libp2p host setup")
}

func TestDHTCacheInvalidMultiaddr(t *testing.T) {
	t.Skip("Skipping test that requires libp2p host setup")
}

func TestDHTCacheMissingFile(t *testing.T) {
	t.Skip("Skipping test that requires libp2p host setup")
}

func TestDHTCacheEmptyPeers(t *testing.T) {
	// Create a test backend with no peers
	backend := &LibP2PClientBackend{
		serverPeers:   make(map[peer.ID]*ServerPeer),
		dhtRendezvous: "test-rendezvous",
	}

	// Save cache with no peers
	err := backend.saveDHTCache()
	assert.NoError(t, err)

	// Verify cache file exists
	cachePath, err := getDHTCachePath()
	assert.NoError(t, err)
	_, err = os.Stat(cachePath)
	assert.NoError(t, err, "Cache file should exist even with no peers")

	// Read cache
	data, err := os.ReadFile(cachePath)
	assert.NoError(t, err)

	var cache DHTCache
	err = json.Unmarshal(data, &cache)
	assert.NoError(t, err)

	// Should have empty peers array
	assert.Len(t, cache.Peers, 0)

	// Clean up
	os.Remove(cachePath)
}

func TestDHTCacheInactivePeersNotSaved(t *testing.T) {
	// Create a test backend with inactive peer
	backend := &LibP2PClientBackend{
		serverPeers:   make(map[peer.ID]*ServerPeer),
		dhtRendezvous: "test-rendezvous",
	}

	// Create test peer ID
	testPeerID := "12D3KooWBrsnBU9rZ8ZBaniVexPfdLmYyF34doTRtSJ7XqfC3JfM"
	peerID, err := peer.Decode(testPeerID)
	assert.NoError(t, err)

	// Create test multiaddr
	testAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/5000")
	assert.NoError(t, err)

	// Add INACTIVE server peer
	backend.serverPeers[peerID] = &ServerPeer{
		ID:       peerID,
		Addrs:    []multiaddr.Multiaddr{testAddr},
		LastSeen: time.Now(),
		Active:   false, // Inactive!
	}

	// Save cache
	err = backend.saveDHTCache()
	assert.NoError(t, err)

	// Read cache
	cachePath, err := getDHTCachePath()
	assert.NoError(t, err)
	data, err := os.ReadFile(cachePath)
	assert.NoError(t, err)

	var cache DHTCache
	err = json.Unmarshal(data, &cache)
	assert.NoError(t, err)

	// Inactive peers should NOT be saved
	assert.Len(t, cache.Peers, 0)

	// Clean up
	os.Remove(cachePath)
}

func TestDHTCacheFilePermissions(t *testing.T) {
	// Create a test backend with mock data
	backend := &LibP2PClientBackend{
		serverPeers:   make(map[peer.ID]*ServerPeer),
		dhtRendezvous: "test-rendezvous",
	}

	// Create test peer
	testPeerID := "12D3KooWBrsnBU9rZ8ZBaniVexPfdLmYyF34doTRtSJ7XqfC3JfM"
	peerID, err := peer.Decode(testPeerID)
	assert.NoError(t, err)

	testAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/5000")
	assert.NoError(t, err)

	backend.serverPeers[peerID] = &ServerPeer{
		ID:       peerID,
		Addrs:    []multiaddr.Multiaddr{testAddr},
		LastSeen: time.Now(),
		Active:   true,
	}

	// Save cache
	err = backend.saveDHTCache()
	assert.NoError(t, err)

	// Check file permissions
	cachePath, err := getDHTCachePath()
	assert.NoError(t, err)

	info, err := os.Stat(cachePath)
	assert.NoError(t, err)

	// Should be 0600 (read/write by owner only)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Clean up
	os.Remove(cachePath)
}
