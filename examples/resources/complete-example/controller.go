package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
)

// ExecutorDeploymentController reconciles ExecutorDeployment resources
type ExecutorDeploymentController struct {
	coloniesClient *client.ColoniesClient
	executorPrvKey string
	colonyName     string
	executorName   string
	deployedExecutors map[string][]string // resource UID -> deployed executor IDs
}

// NewExecutorDeploymentController creates a new controller
func NewExecutorDeploymentController(
	coloniesClient *client.ColoniesClient,
	executorPrvKey string,
	colonyName string,
	executorName string,
) *ExecutorDeploymentController {
	return &ExecutorDeploymentController{
		coloniesClient: coloniesClient,
		executorPrvKey: executorPrvKey,
		colonyName:     colonyName,
		executorName:   executorName,
		deployedExecutors: make(map[string][]string),
	}
}

// Run starts the controller's reconciliation loop
func (c *ExecutorDeploymentController) Run(ctx context.Context) error {
	log.Printf("Starting ExecutorDeployment controller for colony: %s", c.colonyName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Controller shutting down...")
			return ctx.Err()
		default:
			// Try to get work from ColonyOS
			if err := c.processNext(); err != nil {
				log.Printf("Error processing: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// processNext gets the next process and handles it
func (c *ExecutorDeploymentController) processNext() error {
	// Assign a process with our executor type
	process, err := c.coloniesClient.AssignWithContext(
		c.colonyName,
		5, // timeout in seconds
		c.executorPrvKey,
	)

	if err != nil {
		// No work available or error
		return err
	}

	log.Printf("Assigned process: %s (label: %s)", process.ID, process.FunctionSpec.Label)

	// Extract Resource from process kwargs
	resource, err := c.extractResource(process)
	if err != nil {
		return c.failProcess(process.ID, fmt.Sprintf("Failed to extract resource: %v", err))
	}

	// Reconcile the resource
	if err := c.reconcile(resource, process); err != nil {
		return c.failProcess(process.ID, fmt.Sprintf("Reconciliation failed: %v", err))
	}

	// Mark process as successful
	return c.coloniesClient.Close(process.ID, c.executorPrvKey)
}

// extractResource gets the Resource from the FunctionSpec
func (c *ExecutorDeploymentController) extractResource(process *core.Process) (*core.Resource, error) {
	// The resource is directly attached to the FunctionSpec
	if process.FunctionSpec.Resource == nil {
		return nil, fmt.Errorf("no resource attached to process")
	}

	return process.FunctionSpec.Resource, nil
}

// reconcile performs the actual reconciliation logic
func (c *ExecutorDeploymentController) reconcile(resource *core.Resource, process *core.Process) error {
	log.Printf("Reconciling %s/%s", resource.Kind, resource.Metadata.Name)

	// Extract spec fields
	runtime, _ := resource.GetSpec("runtime")
	replicas, _ := resource.GetSpec("replicas")
	executorType, _ := resource.GetSpec("executorType")
	image, _ := resource.GetSpec("image")

	log.Printf("  Runtime: %v", runtime)
	log.Printf("  Replicas: %v", replicas)
	log.Printf("  ExecutorType: %v", executorType)
	log.Printf("  Image: %v", image)

	// Get current deployment state
	currentReplicas := c.getCurrentReplicas(resource.Metadata.UID)
	desiredReplicas := int(replicas.(float64))

	log.Printf("  Current replicas: %d, Desired: %d", currentReplicas, desiredReplicas)

	// Reconcile to desired state
	switch runtime {
	case "docker":
		return c.reconcileDockerDeployment(resource, desiredReplicas)
	case "kubernetes":
		return c.reconcileKubernetesDeployment(resource, desiredReplicas)
	case "local":
		return c.reconcileLocalDeployment(resource, desiredReplicas)
	default:
		return fmt.Errorf("unsupported runtime: %v", runtime)
	}
}

// reconcileDockerDeployment deploys executors using Docker
func (c *ExecutorDeploymentController) reconcileDockerDeployment(resource *core.Resource, desiredReplicas int) error {
	resourceUID := resource.Metadata.UID
	currentExecutors := c.deployedExecutors[resourceUID]
	currentReplicas := len(currentExecutors)

	image, _ := resource.GetSpec("image")
	executorType, _ := resource.GetSpec("executorType")

	// Extract config
	config, hasConfig := resource.GetSpec("config")
	var envVars map[string]interface{}
	if hasConfig {
		if configMap, ok := config.(map[string]interface{}); ok {
			if env, ok := configMap["env"].(map[string]interface{}); ok {
				envVars = env
			}
		}
	}

	// Scale up
	if currentReplicas < desiredReplicas {
		toAdd := desiredReplicas - currentReplicas
		log.Printf("Scaling up: deploying %d new Docker containers", toAdd)

		for i := 0; i < toAdd; i++ {
			containerName := fmt.Sprintf("%s-%s-%d",
				resource.Metadata.Name,
				resourceUID[:8],
				currentReplicas+i)

			// Build docker run command
			args := []string{"run", "-d", "--name", containerName}

			// Add environment variables
			args = append(args, "-e", fmt.Sprintf("COLONIES_EXECUTOR_TYPE=%v", executorType))
			args = append(args, "-e", fmt.Sprintf("COLONIES_COLONY_NAME=%s", c.colonyName))

			for key, val := range envVars {
				args = append(args, "-e", fmt.Sprintf("%s=%v", key, val))
			}

			// Add image
			args = append(args, fmt.Sprintf("%v", image))

			// Execute docker run
			cmd := exec.Command("docker", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Failed to start container: %v\nOutput: %s", err, output)
				continue
			}

			containerID := string(output)[:12]
			log.Printf("Started container: %s (ID: %s)", containerName, containerID)

			currentExecutors = append(currentExecutors, containerID)
		}
	}

	// Scale down
	if currentReplicas > desiredReplicas {
		toRemove := currentReplicas - desiredReplicas
		log.Printf("Scaling down: removing %d Docker containers", toRemove)

		for i := 0; i < toRemove; i++ {
			containerID := currentExecutors[len(currentExecutors)-1]
			currentExecutors = currentExecutors[:len(currentExecutors)-1]

			// Stop and remove container
			exec.Command("docker", "stop", containerID).Run()
			exec.Command("docker", "rm", containerID).Run()
			log.Printf("Removed container: %s", containerID)
		}
	}

	// Update deployed executors map
	c.deployedExecutors[resourceUID] = currentExecutors

	// Update resource status
	resource.SetStatus("phase", "Running")
	resource.SetStatus("ready", len(currentExecutors))
	resource.SetStatus("available", len(currentExecutors))
	resource.SetStatus("deployedExecutors", currentExecutors)
	resource.SetStatus("lastUpdateTime", time.Now().Format(time.RFC3339))

	log.Printf("Reconciliation complete: %d replicas running", len(currentExecutors))
	return nil
}

// reconcileKubernetesDeployment deploys executors to Kubernetes
func (c *ExecutorDeploymentController) reconcileKubernetesDeployment(resource *core.Resource, desiredReplicas int) error {
	// This would use kubectl or the Kubernetes Go client
	log.Printf("Kubernetes deployment not implemented in this example")

	resource.SetStatus("phase", "Pending")
	resource.SetStatus("message", "Kubernetes deployment not implemented")

	return fmt.Errorf("kubernetes deployment not implemented")
}

// reconcileLocalDeployment deploys executors as local processes
func (c *ExecutorDeploymentController) reconcileLocalDeployment(resource *core.Resource, desiredReplicas int) error {
	resourceUID := resource.Metadata.UID
	currentExecutors := c.deployedExecutors[resourceUID]
	currentReplicas := len(currentExecutors)

	executorType, _ := resource.GetSpec("executorType")

	// Scale up
	if currentReplicas < desiredReplicas {
		toAdd := desiredReplicas - currentReplicas
		log.Printf("Scaling up: starting %d local executor processes", toAdd)

		for i := 0; i < toAdd; i++ {
			executorName := fmt.Sprintf("%s-%s-%d",
				resource.Metadata.Name,
				resourceUID[:8],
				currentReplicas+i)

			// Start executor as background process
			// In a real implementation, this would start an actual executor binary
			cmd := exec.Command("sleep", "infinity")
			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start executor: %v", err)
				continue
			}

			processID := fmt.Sprintf("pid-%d", cmd.Process.Pid)
			log.Printf("Started executor: %s (type: %v, PID: %s)",
				executorName, executorType, processID)

			currentExecutors = append(currentExecutors, processID)
		}
	}

	// Scale down
	if currentReplicas > desiredReplicas {
		toRemove := currentReplicas - desiredReplicas
		log.Printf("Scaling down: stopping %d executor processes", toRemove)

		for i := 0; i < toRemove; i++ {
			processID := currentExecutors[len(currentExecutors)-1]
			currentExecutors = currentExecutors[:len(currentExecutors)-1]

			// Kill the process
			// In real implementation, would gracefully stop the executor
			log.Printf("Stopped executor: %s", processID)
		}
	}

	// Update deployed executors map
	c.deployedExecutors[resourceUID] = currentExecutors

	// Update resource status
	resource.SetStatus("phase", "Running")
	resource.SetStatus("ready", len(currentExecutors))
	resource.SetStatus("available", len(currentExecutors))
	resource.SetStatus("deployedExecutors", currentExecutors)
	resource.SetStatus("lastUpdateTime", time.Now().Format(time.RFC3339))

	log.Printf("Reconciliation complete: %d replicas running", len(currentExecutors))
	return nil
}

// getCurrentReplicas returns the current number of deployed executors for a resource
func (c *ExecutorDeploymentController) getCurrentReplicas(resourceUID string) int {
	if executors, ok := c.deployedExecutors[resourceUID]; ok {
		return len(executors)
	}
	return 0
}

// failProcess marks a process as failed
func (c *ExecutorDeploymentController) failProcess(processID string, errorMsg string) error {
	log.Printf("Failing process %s: %s", processID, errorMsg)
	return c.coloniesClient.Fail(processID, []string{errorMsg}, c.executorPrvKey)
}
