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
	rootCmd.AddCommand(serviceCmd)

	// ServiceDefinition commands
	serviceCmd.AddCommand(serviceDefinitionCmd)
	serviceDefinitionCmd.AddCommand(addServiceDefinitionCmd)
	serviceDefinitionCmd.AddCommand(getServiceDefinitionCmd)
	serviceDefinitionCmd.AddCommand(listServiceDefinitionsCmd)
	serviceDefinitionCmd.AddCommand(removeServiceDefinitionCmd)

	// Service commands
	serviceCmd.AddCommand(addServiceCmd)
	serviceCmd.AddCommand(getServiceCmd)
	serviceCmd.AddCommand(listServicesCmd)
	serviceCmd.AddCommand(updateServiceCmd)
	serviceCmd.AddCommand(setServiceCmd)
	serviceCmd.AddCommand(removeServiceCmd)
	serviceCmd.AddCommand(historyServiceCmd)

	// ServiceDefinition flags
	addServiceDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key (colony owner)")
	addServiceDefinitionCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	addServiceDefinitionCmd.MarkFlagRequired("spec")

	getServiceDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getServiceDefinitionCmd.Flags().StringVarP(&ServiceDefinitionName, "name", "", "", "ServiceDefinition name")
	getServiceDefinitionCmd.MarkFlagRequired("name")

	listServiceDefinitionsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")

	removeServiceDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key (colony owner)")
	removeServiceDefinitionCmd.Flags().StringVarP(&ServiceDefinitionName, "name", "", "", "ServiceDefinition name")
	removeServiceDefinitionCmd.MarkFlagRequired("name")

	// Service flags
	addServiceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	addServiceCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	addServiceCmd.MarkFlagRequired("spec")

	getServiceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getServiceCmd.Flags().StringVarP(&ServiceName, "name", "", "", "Service name")
	getServiceCmd.MarkFlagRequired("name")

	listServicesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listServicesCmd.Flags().StringVarP(&Kind, "kind", "", "", "Filter by service kind")

	updateServiceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	updateServiceCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	updateServiceCmd.MarkFlagRequired("spec")

	setServiceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	setServiceCmd.Flags().StringVarP(&ServiceName, "name", "", "", "Service name")
	setServiceCmd.Flags().StringVarP(&Key, "key", "", "", "Field key (use dot notation for nested fields, e.g., 'spec.replicas')")
	setServiceCmd.Flags().StringVarP(&Value, "value", "", "", "New value for the field")
	setServiceCmd.MarkFlagRequired("name")
	setServiceCmd.MarkFlagRequired("key")
	setServiceCmd.MarkFlagRequired("value")

	removeServiceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	removeServiceCmd.Flags().StringVarP(&ServiceName, "name", "", "", "Service name")
	removeServiceCmd.MarkFlagRequired("name")

	historyServiceCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	historyServiceCmd.Flags().StringVarP(&ServiceName, "name", "", "", "Service name")
	historyServiceCmd.Flags().IntVarP(&Count, "limit", "l", 10, "Limit number of history entries")
	historyServiceCmd.Flags().IntVarP(&Generation, "generation", "g", -1, "Show details for specific generation")
	historyServiceCmd.MarkFlagRequired("name")
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
	Long:  "Manage custom services and service definitions",
}

var serviceDefinitionCmd = &cobra.Command{
	Use:   "definition",
	Short: "Manage service definitions",
	Long:  "Manage custom service definitions (CRDs)",
}

var addServiceDefinitionCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a ServiceDefinition",
	Long:  "Add a ServiceDefinition (requires colony owner privileges)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonBytes, err := os.ReadFile(SpecFile)
		CheckError(err)

		var sd core.ServiceDefinition
		err = json.Unmarshal(jsonBytes, &sd)
		CheckError(err)

		// Set colony name if not specified
		if sd.Metadata.Namespace == "" {
			sd.Metadata.Namespace = ColonyName
		}

		addedSD, err := client.AddServiceDefinition(&sd, ColonyPrvKey)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				CheckError(errors.New("ServiceDefinition with name '" + sd.Metadata.Name + "' already exists in colony '" + sd.Metadata.Namespace + "'"))
			}
			CheckError(err)
		}

		log.WithFields(log.Fields{
			"ServiceDefinitionID": addedSD.ID,
			"Name":                 addedSD.Metadata.Name,
			"Kind":                 addedSD.Spec.Names.Kind,
			"Group":                addedSD.Spec.Group,
			"Version":              addedSD.Spec.Version,
			"ColonyName":           addedSD.Metadata.Namespace,
		}).Info("ServiceDefinition added")

	},
}

var getServiceDefinitionCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a ServiceDefinition",
	Long:  "Get a ServiceDefinition by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		sd, err := client.GetServiceDefinition(ColonyName, ServiceDefinitionName, PrvKey)
		CheckError(err)

		if sd == nil {
			CheckError(errors.New("ServiceDefinition not found"))
		}

		log.WithFields(log.Fields{
			"ServiceDefinitionID": sd.ID,
			"Name":                 sd.Metadata.Name,
			"Kind":                 sd.Spec.Names.Kind,
			"ColonyName":           sd.Metadata.Namespace,
		}).Info("ServiceDefinition retrieved")

		if JSON {
			jsonString, err := sd.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
		} else {
			printServiceDefinitionTable(sd)
		}
	},
}

var listServiceDefinitionsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List ServiceDefinitions",
	Long:  "List all ServiceDefinitions in the colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		sds, err := client.GetServiceDefinitions(ColonyName, PrvKey)
		CheckError(err)

		if len(sds) == 0 {
			log.Info("No service definitions found")
			return
		}

		log.WithFields(log.Fields{
			"Count":      len(sds),
			"ColonyName": ColonyName,
		}).Info("ServiceDefinitions retrieved")

		if JSON {
			// Print as JSON array
			jsonBytes, err := json.MarshalIndent(sds, "", "  ")
			CheckError(err)
			fmt.Println(string(jsonBytes))
		} else {
			// Print as table
			printServiceDefinitionsTable(sds)
		}
	},
}

var removeServiceDefinitionCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a ServiceDefinition",
	Long:  "Remove a ServiceDefinition by name (requires colony owner privileges)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveServiceDefinition(ColonyName, ServiceDefinitionName, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"Name":       ServiceDefinitionName,
			"ColonyName": ColonyName,
		}).Info("ServiceDefinition removed")
	},
}

var addServiceCmd = &cobra.Command{
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

		addedService, err := client.AddService(&service, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ServiceID": addedService.ID,
			"Name":       addedService.Metadata.Name,
			"Kind":       addedService.Kind,
			"Namespace":  addedService.Metadata.Namespace,
		}).Info("Service added")

	},
}

var getServiceCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a Service",
	Long:  "Get a service by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		service, err := client.GetService(ColonyName, ServiceName, PrvKey)
		CheckError(err)

		if service == nil {
			CheckError(errors.New("Service not found"))
		}

		log.WithFields(log.Fields{
			"ServiceID": service.ID,
			"Name":       service.Metadata.Name,
			"Kind":       service.Kind,
			"Namespace":  service.Metadata.Namespace,
		}).Info("Service retrieved")

		if JSON {
			jsonString, err := service.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
		} else {
			printServiceTable(client, service)
		}
	},
}

var listServicesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Services",
	Long:  "List all services in the colony (optionally filtered by kind)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		services, err := client.GetServices(ColonyName, Kind, PrvKey)
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
			printServicesTable(services)
		}
	},
}

var updateServiceCmd = &cobra.Command{
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

		updatedService, err := client.UpdateService(&service, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ServiceID": updatedService.ID,
			"Name":       updatedService.Metadata.Name,
			"Kind":       updatedService.Kind,
			"Generation": updatedService.Metadata.Generation,
		}).Info("Service updated")

	},
}

var setServiceCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a field value in a Service",
	Long:  "Set a specific field value in a service using dot notation (e.g., 'replicas' or 'env.TZ')",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		// Get the existing service
		service, err := client.GetService(ColonyName, ServiceName, PrvKey)
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
		updatedService, err := client.UpdateService(service, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ServiceID": updatedService.ID,
			"Name":       updatedService.Metadata.Name,
			"Kind":       updatedService.Kind,
			"Key":        Key,
			"Value":      Value,
		}).Info("Service field updated")

	},
}

var removeServiceCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Service",
	Long:  "Remove a service by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveService(ColonyName, ServiceName, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"Name":       ServiceName,
			"ColonyName": ColonyName,
		}).Info("Service removed")
	},
}

var historyServiceCmd = &cobra.Command{
	Use:   "history",
	Short: "Show service history",
	Long:  "Display the history of changes to a service",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		// Get the service to find its ID
		service, err := client.GetService(ColonyName, ServiceName, PrvKey)
		CheckError(err)

		// Get the service history
		histories, err := client.GetServiceHistory(service.ID, Count, PrvKey)
		CheckError(err)

		if len(histories) == 0 {
			log.Info("No history found for this service")
			return
		}

		// If generation is specified, show detailed view of that generation
		if Generation >= 0 {
			var selectedHistory *core.ServiceHistory
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
				printServiceHistoryDetail(selectedHistory)
			}
		} else {
			// Print history table
			if JSON {
				jsonString, err := core.ConvertServiceHistoryArrayToJSON(histories)
				CheckError(err)
				fmt.Println(jsonString)
			} else {
				printServiceHistoryTable(client, histories)
			}
		}
	},
}
