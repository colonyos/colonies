package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resourceCmd)

	// ResourceDefinition commands
	resourceCmd.AddCommand(resourceDefinitionCmd)
	resourceDefinitionCmd.AddCommand(addResourceDefinitionCmd)
	resourceDefinitionCmd.AddCommand(getResourceDefinitionCmd)
	resourceDefinitionCmd.AddCommand(listResourceDefinitionsCmd)
	resourceDefinitionCmd.AddCommand(removeResourceDefinitionCmd)

	// Service commands
	resourceCmd.AddCommand(addResourceCmd)
	resourceCmd.AddCommand(getResourceCmd)
	resourceCmd.AddCommand(listResourcesCmd)
	resourceCmd.AddCommand(updateResourceCmd)
	resourceCmd.AddCommand(setResourceCmd)
	resourceCmd.AddCommand(removeResourceCmd)
	resourceCmd.AddCommand(historyResourceCmd)

	// ResourceDefinition flags
	addResourceDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key (colony owner)")
	addResourceDefinitionCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	addResourceDefinitionCmd.MarkFlagRequired("spec")

	getResourceDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getResourceDefinitionCmd.Flags().StringVarP(&ResourceDefinitionName, "name", "", "", "ResourceDefinition name")
	getResourceDefinitionCmd.MarkFlagRequired("name")

	listResourceDefinitionsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")

	removeResourceDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key (colony owner)")
	removeResourceDefinitionCmd.Flags().StringVarP(&ResourceDefinitionName, "name", "", "", "ResourceDefinition name")
	removeResourceDefinitionCmd.MarkFlagRequired("name")

	// Service flags
	addResourceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	addResourceCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	addResourceCmd.MarkFlagRequired("spec")

	getResourceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getResourceCmd.Flags().StringVarP(&ResourceName, "name", "", "", "Service name")
	getResourceCmd.MarkFlagRequired("name")

	listResourcesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listResourcesCmd.Flags().StringVarP(&Kind, "kind", "", "", "Filter by service kind")

	updateResourceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	updateResourceCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	updateResourceCmd.MarkFlagRequired("spec")

	setResourceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	setResourceCmd.Flags().StringVarP(&ResourceName, "name", "", "", "Service name")
	setResourceCmd.Flags().StringVarP(&Key, "key", "", "", "Field key (use dot notation for nested fields, e.g., 'spec.replicas')")
	setResourceCmd.Flags().StringVarP(&Value, "value", "", "", "New value for the field")
	setResourceCmd.MarkFlagRequired("name")
	setResourceCmd.MarkFlagRequired("key")
	setResourceCmd.MarkFlagRequired("value")

	removeResourceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	removeResourceCmd.Flags().StringVarP(&ResourceName, "name", "", "", "Service name")
	removeResourceCmd.MarkFlagRequired("name")

	historyResourceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	historyResourceCmd.Flags().StringVarP(&ResourceName, "name", "", "", "Service name")
	historyResourceCmd.Flags().IntVarP(&Count, "limit", "l", 10, "Limit number of history entries")
	historyResourceCmd.Flags().IntVarP(&Generation, "generation", "g", -1, "Show details for specific generation")
	historyResourceCmd.MarkFlagRequired("name")
}

var resourceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
	Long:  "Manage custom services and service definitions",
}

var resourceDefinitionCmd = &cobra.Command{
	Use:   "definition",
	Short: "Manage service definitions",
	Long:  "Manage custom service definitions (CRDs)",
}

var addResourceDefinitionCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a ResourceDefinition",
	Long:  "Add a ResourceDefinition (requires colony owner privileges)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonBytes, err := os.ReadFile(SpecFile)
		CheckError(err)

		var rd core.ResourceDefinition
		err = json.Unmarshal(jsonBytes, &rd)
		CheckError(err)

		// Set colony name if not specified
		if rd.Metadata.Namespace == "" {
			rd.Metadata.Namespace = ColonyName
		}

		addedRD, err := client.AddResourceDefinition(&rd, ColonyPrvKey)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				CheckError(errors.New("ResourceDefinition with name '" + rd.Metadata.Name + "' already exists in colony '" + rd.Metadata.Namespace + "'"))
			}
			CheckError(err)
		}

		log.WithFields(log.Fields{
			"ResourceDefinitionID": addedRD.ID,
			"Name":                 addedRD.Metadata.Name,
			"Kind":                 addedRD.Spec.Names.Kind,
			"Group":                addedRD.Spec.Group,
			"Version":              addedRD.Spec.Version,
			"ColonyName":           addedRD.Metadata.Namespace,
		}).Info("ResourceDefinition added")

	},
}

var getResourceDefinitionCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a ResourceDefinition",
	Long:  "Get a ResourceDefinition by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		rd, err := client.GetResourceDefinition(ColonyName, ResourceDefinitionName, PrvKey)
		CheckError(err)

		if rd == nil {
			CheckError(errors.New("ResourceDefinition not found"))
		}

		log.WithFields(log.Fields{
			"ResourceDefinitionID": rd.ID,
			"Name":                 rd.Metadata.Name,
			"Kind":                 rd.Spec.Names.Kind,
			"ColonyName":           rd.Metadata.Namespace,
		}).Info("ResourceDefinition retrieved")

		if JSON {
			jsonString, err := rd.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
		} else {
			printResourceDefinitionTable(rd)
		}
	},
}

var listResourceDefinitionsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List ResourceDefinitions",
	Long:  "List all ResourceDefinitions in the colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		rds, err := client.GetResourceDefinitions(ColonyName, PrvKey)
		CheckError(err)

		if len(rds) == 0 {
			log.Info("No service definitions found")
			return
		}

		log.WithFields(log.Fields{
			"Count":      len(rds),
			"ColonyName": ColonyName,
		}).Info("ResourceDefinitions retrieved")

		if JSON {
			// Print as JSON array
			jsonBytes, err := json.MarshalIndent(rds, "", "  ")
			CheckError(err)
			fmt.Println(string(jsonBytes))
		} else {
			// Print as table
			printResourceDefinitionsTable(rds)
		}
	},
}

var removeResourceDefinitionCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a ResourceDefinition",
	Long:  "Remove a ResourceDefinition by name (requires colony owner privileges)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveResourceDefinition(ColonyName, ResourceDefinitionName, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"Name":       ResourceDefinitionName,
			"ColonyName": ColonyName,
		}).Info("ResourceDefinition removed")
	},
}

var addResourceCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a Service",
	Long:  "Add a custom service instance",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonBytes, err := os.ReadFile(SpecFile)
		CheckError(err)

		var service core.Service
		err = json.Unmarshal(jsonBytes, &service)
		CheckError(err)

		// Set namespace (colony name) if not specified
		if service.Metadata.Namespace == "" {
			service.Metadata.Namespace = ColonyName
		}

		addedResource, err := client.AddResource(&service, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ResourceID": addedResource.ID,
			"Name":       addedResource.Metadata.Name,
			"Kind":       addedResource.Kind,
			"Namespace":  addedResource.Metadata.Namespace,
		}).Info("Service added")

	},
}

var getResourceCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a Service",
	Long:  "Get a service by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		service, err := client.GetResource(ColonyName, ResourceName, PrvKey)
		CheckError(err)

		if service == nil {
			CheckError(errors.New("Service not found"))
		}

		log.WithFields(log.Fields{
			"ResourceID": service.ID,
			"Name":       service.Metadata.Name,
			"Kind":       service.Kind,
			"Namespace":  service.Metadata.Namespace,
		}).Info("Service retrieved")

		if JSON {
			jsonString, err := service.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
		} else {
			printResourceTable(client, service)
		}
	},
}

var listResourcesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Services",
	Long:  "List all services in the colony (optionally filtered by kind)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		services, err := client.GetResources(ColonyName, Kind, PrvKey)
		CheckError(err)

		if len(services) == 0 {
			log.Info("No services found")
			return
		}

		log.WithFields(log.Fields{
			"Count":      len(services),
			"ColonyName": ColonyName,
			"Kind":       Kind,
		}).Info("Services retrieved")

		if JSON {
			// Print as JSON array
			jsonBytes, err := json.MarshalIndent(services, "", "  ")
			CheckError(err)
			fmt.Println(string(jsonBytes))
		} else {
			// Print as table
			printResourcesTable(services)
		}
	},
}

var updateResourceCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Service",
	Long:  "Update an existing service",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonBytes, err := os.ReadFile(SpecFile)
		CheckError(err)

		var service core.Service
		err = json.Unmarshal(jsonBytes, &service)
		CheckError(err)

		// Set namespace (colony name) if not specified
		if service.Metadata.Namespace == "" {
			service.Metadata.Namespace = ColonyName
		}

		updatedResource, err := client.UpdateResource(&service, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ResourceID": updatedResource.ID,
			"Name":       updatedResource.Metadata.Name,
			"Kind":       updatedResource.Kind,
			"Generation": updatedResource.Metadata.Generation,
		}).Info("Service updated")

	},
}

var setResourceCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a field value in a Service",
	Long:  "Set a specific field value in a service using dot notation (e.g., 'replicas' or 'env.TZ')",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		// Get the existing service
		service, err := client.GetResource(ColonyName, ResourceName, PrvKey)
		CheckError(err)

		// Parse the service spec into a map for easy manipulation
		specMap := make(map[string]interface{})
		specBytes, err := json.Marshal(service.Spec)
		CheckError(err)
		err = json.Unmarshal(specBytes, &specMap)
		CheckError(err)

		// Remove "spec." prefix if present (for convenience)
		keyPath := Key
		if strings.HasPrefix(keyPath, "spec.") {
			keyPath = strings.TrimPrefix(keyPath, "spec.")
		}

		// Split the key by dots to handle nested fields
		keyParts := strings.Split(keyPath, ".")

		// First, validate that the key path exists in the current spec
		// We don't allow creating new fields to prevent corrupting the service
		current := specMap
		for i := 0; i < len(keyParts)-1; i++ {
			key := keyParts[i]
			if _, ok := current[key]; !ok {
				CheckError(errors.New("Invalid key path: '" + Key + "' (field '" + key + "' does not exist in service spec)"))
			}
			var ok bool
			current, ok = current[key].(map[string]interface{})
			if !ok {
				CheckError(errors.New("Invalid key path: " + Key + " (cannot navigate through non-object at '" + key + "')"))
			}
		}

		// Set the final value
		finalKey := keyParts[len(keyParts)-1]

		// Validate that the final key exists
		if _, ok := current[finalKey]; !ok {
			CheckError(errors.New("Invalid key: '" + Key + "' (field '" + finalKey + "' does not exist in service spec)"))
		}

		// Try to parse the value as JSON to support numbers, booleans, etc.
		var parsedValue interface{}
		err = json.Unmarshal([]byte(Value), &parsedValue)
		if err != nil {
			// If it's not valid JSON, treat it as a string
			parsedValue = Value
		}

		current[finalKey] = parsedValue

		// Update the service spec
		service.Spec = specMap

		// Update the service in the colony
		updatedResource, err := client.UpdateResource(service, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ResourceID": updatedResource.ID,
			"Name":       updatedResource.Metadata.Name,
			"Kind":       updatedResource.Kind,
			"Key":        Key,
			"Value":      Value,
		}).Info("Service field updated")

	},
}

var removeResourceCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Service",
	Long:  "Remove a service by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveResource(ColonyName, ResourceName, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"Name":       ResourceName,
			"ColonyName": ColonyName,
		}).Info("Service removed")
	},
}

var historyResourceCmd = &cobra.Command{
	Use:   "history",
	Short: "Show service history",
	Long:  "Display the history of changes to a service",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		// Get the service to find its ID
		service, err := client.GetResource(ColonyName, ResourceName, PrvKey)
		CheckError(err)

		// Get the service history
		histories, err := client.GetResourceHistory(service.ID, Count, PrvKey)
		CheckError(err)

		if len(histories) == 0 {
			log.Info("No history found for this service")
			return
		}

		// If generation is specified, show detailed view of that generation
		if Generation >= 0 {
			var selectedHistory *core.ResourceHistory
			for _, h := range histories {
				if h.Generation == int64(Generation) {
					selectedHistory = h
					break
				}
			}

			if selectedHistory == nil {
				CheckError(errors.New(fmt.Sprintf("Generation %d not found in history", Generation)))
			}

			if JSON {
				jsonString, err := selectedHistory.ToJSON()
				CheckError(err)
				fmt.Println(jsonString)
			} else {
				printResourceHistoryDetail(selectedHistory)
			}
		} else {
			// Print history table
			if JSON {
				jsonString, err := core.ConvertResourceHistoryArrayToJSON(histories)
				CheckError(err)
				fmt.Println(jsonString)
			} else {
				printResourceHistoryTable(client, histories)
			}
		}
	},
}
