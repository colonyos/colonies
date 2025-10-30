package gitops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNewGitOpsSync(t *testing.T) {
	sync := NewGitOpsSync("")
	assert.NotNil(t, sync)
	assert.NotEmpty(t, sync.workDir)

	customDir := "/tmp/test-gitops"
	sync = NewGitOpsSync(customDir)
	assert.Equal(t, customDir, sync.workDir)
}

func TestLoadResourcesFromPath(t *testing.T) {
	// Create a temporary directory with test resources
	tempDir := t.TempDir()

	// Create a valid resource file
	resource1 := core.CreateResource("TestKind", "test-resource-1", "test-namespace")
	resource1.SetSpec("key1", "value1")

	resource1JSON, err := json.MarshalIndent(resource1, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "resource1.json"), resource1JSON, 0644)
	assert.NoError(t, err)

	// Create another valid resource file
	resource2 := core.CreateResource("TestKind", "test-resource-2", "test-namespace")
	resource2.SetSpec("key2", "value2")

	resource2JSON, err := json.MarshalIndent(resource2, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "resource2.json"), resource2JSON, 0644)
	assert.NoError(t, err)

	// Create an invalid file (should be skipped)
	err = os.WriteFile(filepath.Join(tempDir, "invalid.json"), []byte("not a resource"), 0644)
	assert.NoError(t, err)

	// Create a text file (should be skipped)
	err = os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("readme"), 0644)
	assert.NoError(t, err)

	// Test loading resources
	sync := NewGitOpsSync("")
	resources, err := sync.loadResourcesFromPath(tempDir, "TestKind")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resources))

	// Test loading with kind filter
	resources, err = sync.loadResourcesFromPath(tempDir, "OtherKind")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(resources))

	// Test loading from non-existent path
	resources, err = sync.loadResourcesFromPath("/non/existent/path", "TestKind")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(resources))
}

func TestHashString(t *testing.T) {
	hash1 := hashString("https://github.com/user/repo.git")
	assert.NotEmpty(t, hash1)

	hash2 := hashString("https://github.com/user/repo.git")
	assert.Equal(t, hash1, hash2)

	hash3 := hashString("https://github.com/user/other-repo.git")
	assert.NotEqual(t, hash1, hash3)

	// Test long URLs are truncated
	longURL := "https://github.com/" + string(make([]byte, 100))
	hash := hashString(longURL)
	assert.LessOrEqual(t, len(hash), 50)
}

func TestSyncResources_NoGitOpsConfig(t *testing.T) {
	sync := NewGitOpsSync("")
	rd := core.CreateResourceDefinition("test", "test.io", "v1", "TestKind", "testkinds", "Namespaced", "test-executor", "reconcile")

	resources, err := sync.SyncResources(rd)
	assert.Error(t, err)
	assert.Nil(t, resources)
	assert.Contains(t, err.Error(), "does not have GitOps configuration")
}

// Note: Testing actual git clone/pull would require a real git repository
// and network access, which is not ideal for unit tests. Integration tests
// should cover those scenarios.
