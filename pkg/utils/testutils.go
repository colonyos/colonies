package utils

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
)

func CreateTestProcess(colonyID string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpec(colonyID))
}

func CreateTestProcessSpec(colonyID string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string), []string{}, 1)
}

func CreateTestProcessWithType(colonyID string, runtimeType string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithType(colonyID, runtimeType))
}

func CreateTestProcessSpecWithType(colonyID string, runtimeType string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, -1, 3, 1000, 10, 1, make(map[string]string), []string{}, 1)
}

func CreateTestProcessWithEnv(colonyID string, env map[string]string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithEnv(colonyID, env))
}

func CreateTestProcessSpecWithEnv(colonyID string, env map[string]string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, env, []string{}, 1)
}

func CreateTestProcessWithTargets(colonyID string, targetRuntimeIDs []string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithTargets(colonyID, targetRuntimeIDs))
}

func CreateTestProcessSpecWithTargets(colonyID string, targetRuntimeIDs []string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, targetRuntimeIDs, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string), []string{}, 1)
}

func CreateTestRuntime(colonyID string) *core.Runtime {
	return core.CreateRuntime(core.GenerateRandomID(), "test_runtime_type", "test_runtime_name", colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestRuntimeWithType(colonyID string, runtimeType string) *core.Runtime {
	return core.CreateRuntime(core.GenerateRandomID(), runtimeType, "test_runtime_name", colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestRuntimeWithID(colonyID string, runtimeID string) *core.Runtime {
	return core.CreateRuntime(runtimeID, "test_runtime_type", "test_runtime_name", colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestRuntimeWithKey(colonyID string) (*core.Runtime, string, error) {
	crypto := crypto.CreateCrypto()
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", err
	}

	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	if err != nil {
		return nil, "", err
	}

	return core.CreateRuntime(runtimeID, "test_runtime_type", "test_runtime_name", colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now()), runtimePrvKey, nil
}

func CreateTestColonyWithKey() (*core.Colony, string, error) {
	crypto := crypto.CreateCrypto()

	colonyPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", err
	}

	colonyID, err := crypto.GenerateID(colonyPrvKey)
	if err != nil {
		return nil, "", err
	}
	return core.CreateColony(colonyID, "test_colony_name"), colonyPrvKey, nil
}
