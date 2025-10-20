package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
)

func main() {
	// Command line flags
	mode := flag.String("mode", "", "Mode: register-crd, controller, submit, or demo")
	rdFile := flag.String("crd", "executor-deployment-definition.json", "Path to CRD file")
	resourceFile := flag.String("resource", "ml-executor-deployment.json", "Path to resource file")
	flag.Parse()

	if *mode == "" {
		log.Fatal("Please specify -mode: register-crd, controller, submit, or demo")
	}

	// Initialize ColonyOS client
	coloniesClient := createColoniesClient()

	switch *mode {
	case "register-crd":
		registerResourceDefinition(coloniesClient, *rdFile)
	case "controller":
		runController(coloniesClient)
	case "submit":
		submitResource(coloniesClient, *resourceFile)
	case "demo":
		runDemo(coloniesClient, *rdFile, *resourceFile)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

// createColoniesClient creates a ColonyOS client from environment variables
func createColoniesClient() *client.ColoniesClient {
	serverHost := getEnvOrDefault("COLONIES_SERVER_HOST", "localhost")
	serverPort := getEnvOrDefault("COLONIES_SERVER_PORT", "50080")
	insecure := getEnvOrDefault("COLONIES_INSECURE", "true") == "true"

	return client.CreateColoniesClient(serverHost, atoi(serverPort), insecure, false)
}

// registerResourceDefinition registers a ResourceDefinition with ColonyOS
func registerResourceDefinition(coloniesClient *client.ColoniesClient, rdFile string) {
	log.Printf("Registering CRD from file: %s", rdFile)

	// Read CRD file
	data, err := ioutil.ReadFile(rdFile)
	if err != nil {
		log.Fatalf("Failed to read CRD file: %v", err)
	}

	// Parse CRD
	crd, err := core.ConvertJSONToResourceDefinition(string(data))
	if err != nil {
		log.Fatalf("Failed to parse CRD: %v", err)
	}

	// Validate CRD
	if err := crd.Validate(); err != nil {
		log.Fatalf("CRD validation failed: %v", err)
	}

	log.Printf("CRD Details:")
	log.Printf("  Name: %s", crd.Metadata.Name)
	log.Printf("  Group: %s", crd.Spec.Group)
	log.Printf("  Version: %s", crd.Spec.Version)
	log.Printf("  Kind: %s", crd.Spec.Names.Kind)
	log.Printf("  Handler Type: %s", crd.Spec.Handler.ExecutorType)
	log.Printf("  Handler Function: %s", crd.Spec.Handler.FunctionName)

	// In a real implementation, this would store the CRD in the database
	// For now, we just print it
	log.Println("✓ CRD registered successfully")
	log.Println("\nNote: In a full implementation, the CRD would be stored in the ColonyOS database")
}

// runController starts the executor deployment controller
func runController(coloniesClient *client.ColoniesClient) {
	colonyName := getEnvOrDefault("COLONIES_COLONY_NAME", "dev")
	executorName := getEnvOrDefault("COLONIES_EXECUTOR_NAME", "executor-deployment-controller")
	executorPrvKey := getEnvOrDefault("COLONIES_EXECUTOR_PRVKEY", "")

	if executorPrvKey == "" {
		log.Println("No COLONIES_EXECUTOR_PRVKEY provided, generating new key pair...")
		crypto := crypto.CreateCrypto()
		executorPrvKey, _ = crypto.GeneratePrivateKey()
		executorPubKey, _ := crypto.GetPublicKey(executorPrvKey)
		log.Printf("Generated executor private key: %s", executorPrvKey)
		log.Printf("Generated executor public key: %s", executorPubKey)
	}

	log.Printf("Starting controller:")
	log.Printf("  Colony: %s", colonyName)
	log.Printf("  Executor Name: %s", executorName)
	log.Printf("  Executor Type: executor-deployment-controller")

	// Create and start controller
	controller := NewExecutorDeploymentController(
		coloniesClient,
		executorPrvKey,
		colonyName,
		executorName,
	)

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("\nReceived shutdown signal...")
		cancel()
	}()

	// Run controller
	log.Println("Controller is running. Press Ctrl+C to stop.")
	if err := controller.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Controller error: %v", err)
	}

	log.Println("Controller stopped")
}

// submitResource submits a Resource to ColonyOS
func submitResource(coloniesClient *client.ColoniesClient, resourceFile string) {
	log.Printf("Submitting Resource from file: %s", resourceFile)

	// Read resource file
	data, err := ioutil.ReadFile(resourceFile)
	if err != nil {
		log.Fatalf("Failed to read resource file: %v", err)
	}

	// Parse resource
	resource, err := core.ConvertJSONToResource(string(data))
	if err != nil {
		log.Fatalf("Failed to parse resource: %v", err)
	}

	// Validate resource
	if err := resource.Validate(); err != nil {
		log.Fatalf("Resource validation failed: %v", err)
	}

	log.Printf("Resource Details:")
	log.Printf("  API Version: %s", resource.APIVersion)
	log.Printf("  Kind: %s", resource.Kind)
	log.Printf("  Name: %s", resource.Metadata.Name)
	log.Printf("  Namespace: %s", resource.Metadata.Namespace)

	// Print spec
	log.Printf("  Spec:")
	for key, val := range resource.Spec {
		log.Printf("    %s: %v", key, val)
	}

	// In a real implementation, the ColonyOS server would:
	// 1. Look up the CRD for this resource type
	log.Println("\n✓ Resource ready for submission")
	log.Println("Note: The server will:")
	log.Println("  1. Look up the CRD by apiVersion+kind")
	log.Println("  2. Validate the Resource against the CRD schema")
	log.Println("  3. Convert to a Process using attach Resource to FunctionSpec")
	log.Println("  4. Submit the Process to the appropriate controller")
}

// runDemo runs a complete demonstration
func runDemo(coloniesClient *client.ColoniesClient, rdFile, resourceFile string) {
	log.Println("=== ExecutorDeployment CRD Demo ===\n")

	// Step 1: Register CRD
	log.Println("Step 1: Registering CRD")
	log.Println("─────────────────────────")
	registerResourceDefinition(coloniesClient, rdFile)
	time.Sleep(2 * time.Second)

	// Step 2: Show resource
	log.Println("\n\nStep 2: Loading Resource")
	log.Println("─────────────────────────────────")

	data, _ := ioutil.ReadFile(resourceFile)
	resource, _ := core.ConvertJSONToResource(string(data))

	log.Printf("Resource: %s/%s", resource.Kind, resource.Metadata.Name)
	log.Printf("Desired State:")
	runtime, _ := resource.GetSpec("runtime")
	replicas, _ := resource.GetSpec("replicas")
	executorType, _ := resource.GetSpec("executorType")
	log.Printf("  - Runtime: %v", runtime)
	log.Printf("  - Replicas: %v", replicas)
	log.Printf("  - Executor Type: %v", executorType)

	time.Sleep(2 * time.Second)

	// Step 3: Validate Resource Against CRD
	log.Println("\n\nStep 3: Validate Resource Against CRD")
	log.Println("──────────────────────────────────────")

	// Load the CRD again for validation
	crdData, _ := ioutil.ReadFile(rdFile)
	crd, _ := core.ConvertJSONToResourceDefinition(string(crdData))

	log.Println("Validating Resource against CRD schema...")
	if err := resource.ValidateAgainstRD(crd); err != nil {
		log.Printf("  ✗ Validation failed: %v", err)
	} else {
		log.Println("  ✓ Validation passed")
		log.Println("  ✓ API version matches")
		log.Println("  ✓ Kind matches")
		log.Println("  ✓ All required fields present")
		log.Println("  ✓ Field types correct")
	}

	time.Sleep(2 * time.Second)

	// Step 4: Server converts to Process
	log.Println("\n\nStep 4: Server Converts to Process")
	log.Println("───────────────────────────────────")

	log.Println("ColonyOS server would:")
	log.Println("  1. Look up CRD for ExecutorDeployment")
	log.Println("  2. Validate Resource against CRD schema")
	log.Println("  3. Create FunctionSpec from Resource")
	log.Println("  4. Create Process with executor type: executor-deployment-controller")
	log.Println("  5. Function to call: reconcile_executor_deployment")
	log.Println("  6. Submit Process to queue")

	time.Sleep(2 * time.Second)

	// Step 5: Simulate reconciliation
	log.Println("\n\nStep 5: Simulating Reconciliation")
	log.Println("────────────────────────────────────")

	log.Println("Controller would:")
	log.Println("  1. Receive the Process")
	log.Println("  2. Extract Resource from kwargs")
	log.Println("  3. Compare desired state vs current state")
	log.Println("  4. Deploy/scale executors accordingly")
	log.Println("  5. Update resource status")
	log.Println("  6. Mark Process as complete")

	// Simulate status update
	resource.SetStatus("phase", "Running")
	resource.SetStatus("ready", replicas)
	resource.SetStatus("available", replicas)
	resource.SetStatus("deployedExecutors", []string{
		"exec-ml-001",
		"exec-ml-002",
		"exec-ml-003",
	})
	resource.SetStatus("lastUpdateTime", time.Now().Format(time.RFC3339))

	log.Println("\n  Updated Status:")
	statusJSON, _ := json.MarshalIndent(resource.Status, "  ", "  ")
	fmt.Printf("  %s\n", statusJSON)

	// Summary
	log.Println("\n\n=== Summary ===")
	log.Println("This demo showed:")
	log.Println("  ✓ CRD registration and validation")
	log.Println("  ✓ Resource creation")
	log.Println("  ✓ Schema validation against CRD")
	log.Println("  ✓ Conversion to ColonyOS Process")
	log.Println("  ✓ Controller reconciliation logic")
	log.Println("  ✓ Status updates")
	log.Println("\nTo run a real deployment:")
	log.Println("  1. Start ColonyOS server")
	log.Println("  2. Run: go run . -mode register-crd")
	log.Println("  3. Run: go run . -mode controller (in separate terminal)")
	log.Println("  4. Run: go run . -mode submit")
}

// createExecutorDeploymentCRD creates a CRD programmatically
func createExecutorDeploymentCRD() *core.ResourceDefinition {
	return core.CreateResourceDefinition(
		"executordeployments.compute.colonies.io",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"executor-deployment-controller",
		"reconcile_executor_deployment",
	)
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func atoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
