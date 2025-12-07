package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(blueprintCmd)

	// BlueprintDefinition commands
	blueprintCmd.AddCommand(blueprintDefinitionCmd)
	blueprintDefinitionCmd.AddCommand(addBlueprintDefinitionCmd)
	blueprintDefinitionCmd.AddCommand(getBlueprintDefinitionCmd)
	blueprintDefinitionCmd.AddCommand(listBlueprintDefinitionsCmd)
	blueprintDefinitionCmd.AddCommand(removeBlueprintDefinitionCmd)

	// Blueprint commands
	blueprintCmd.AddCommand(addBlueprintCmd)
	blueprintCmd.AddCommand(getBlueprintCmd)
	blueprintCmd.AddCommand(listBlueprintsCmd)
	blueprintCmd.AddCommand(updateBlueprintCmd)
	blueprintCmd.AddCommand(setBlueprintCmd)
	blueprintCmd.AddCommand(removeBlueprintCmd)
	blueprintCmd.AddCommand(reconcileBlueprintCmd)
	blueprintCmd.AddCommand(historyBlueprintCmd)
	blueprintCmd.AddCommand(logBlueprintCmd)
	blueprintCmd.AddCommand(doctorBlueprintCmd)

	// BlueprintDefinition flags
	addBlueprintDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key (colony owner)")
	addBlueprintDefinitionCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	addBlueprintDefinitionCmd.MarkFlagRequired("spec")

	getBlueprintDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getBlueprintDefinitionCmd.Flags().StringVarP(&BlueprintDefinitionName, "name", "", "", "BlueprintDefinition name")
	getBlueprintDefinitionCmd.MarkFlagRequired("name")

	listBlueprintDefinitionsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")

	removeBlueprintDefinitionCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key (colony owner)")
	removeBlueprintDefinitionCmd.Flags().StringVarP(&BlueprintDefinitionName, "name", "", "", "BlueprintDefinition name")
	removeBlueprintDefinitionCmd.MarkFlagRequired("name")

	// Blueprint flags
	addBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	addBlueprintCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	addBlueprintCmd.MarkFlagRequired("spec")

	getBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getBlueprintCmd.Flags().StringVarP(&BlueprintName, "name", "", "", "Blueprint name")
	getBlueprintCmd.MarkFlagRequired("name")

	listBlueprintsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listBlueprintsCmd.Flags().StringVarP(&Kind, "kind", "", "", "Filter by blueprint kind")

	updateBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	updateBlueprintCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification file")
	updateBlueprintCmd.MarkFlagRequired("spec")

	setBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	setBlueprintCmd.Flags().StringVarP(&BlueprintName, "name", "", "", "Blueprint name")
	setBlueprintCmd.Flags().StringVarP(&Key, "key", "", "", "Field key (use dot notation for nested fields, e.g., 'spec.replicas')")
	setBlueprintCmd.Flags().StringVarP(&Value, "value", "", "", "New value for the field")
	setBlueprintCmd.MarkFlagRequired("name")
	setBlueprintCmd.MarkFlagRequired("key")
	setBlueprintCmd.MarkFlagRequired("value")

	removeBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	removeBlueprintCmd.Flags().StringVarP(&BlueprintName, "name", "", "", "Blueprint name")
	removeBlueprintCmd.MarkFlagRequired("name")

	reconcileBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	reconcileBlueprintCmd.Flags().StringVarP(&BlueprintName, "name", "", "", "Blueprint name")
	reconcileBlueprintCmd.Flags().BoolVarP(&Force, "force", "f", false, "Force recreation of all containers (restarts with fresh image)")
	reconcileBlueprintCmd.MarkFlagRequired("name")

	historyBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	historyBlueprintCmd.Flags().StringVarP(&BlueprintName, "name", "", "", "Blueprint name")
	historyBlueprintCmd.Flags().IntVarP(&Count, "limit", "l", 10, "Limit number of history entries")
	historyBlueprintCmd.Flags().IntVarP(&Generation, "generation", "g", -1, "Show details for specific generation")
	historyBlueprintCmd.MarkFlagRequired("name")

	logBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	logBlueprintCmd.Flags().StringVarP(&BlueprintName, "name", "", "", "Blueprint name (optional, shows all reconciler logs if omitted)")
	logBlueprintCmd.Flags().IntVarP(&Count, "count", "c", 100, "Number of log messages to fetch")
	logBlueprintCmd.Flags().BoolVarP(&Follow, "follow", "f", false, "Follow logs in real-time")

	doctorBlueprintCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	doctorBlueprintCmd.Flags().StringVarP(&BlueprintName, "name", "", "", "Blueprint name (optional, checks all blueprints if omitted)")
}

var blueprintCmd = &cobra.Command{
	Use:   "blueprint",
	Short: "Manage blueprints",
	Long:  "Manage custom blueprints and blueprint definitions",
}

var blueprintDefinitionCmd = &cobra.Command{
	Use:   "definition",
	Short: "Manage blueprint definitions",
	Long:  "Manage custom blueprint definitions (CRDs)",
}

var addBlueprintDefinitionCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a BlueprintDefinition",
	Long:  "Add a BlueprintDefinition (requires colony owner privileges)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonBytes, err := os.ReadFile(SpecFile)
		CheckError(err)

		var sd core.BlueprintDefinition
		err = json.Unmarshal(jsonBytes, &sd)
		CheckError(err)

		// Set colony name if not specified
		if sd.Metadata.ColonyName == "" {
			sd.Metadata.ColonyName = ColonyName
		}

		addedSD, err := client.AddBlueprintDefinition(&sd, ColonyPrvKey)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				CheckError(errors.New("BlueprintDefinition with name '" + sd.Metadata.Name + "' already exists in colony '" + sd.Metadata.ColonyName + "'"))
			}
			CheckError(err)
		}

		log.WithFields(log.Fields{
			"BlueprintDefinitionID": addedSD.ID,
			"Name":                 addedSD.Metadata.Name,
			"Kind":                 addedSD.Spec.Names.Kind,
			"Group":                addedSD.Spec.Group,
			"Version":              addedSD.Spec.Version,
			"ColonyName":           addedSD.Metadata.ColonyName,
		}).Info("BlueprintDefinition added")

	},
}

var getBlueprintDefinitionCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a BlueprintDefinition",
	Long:  "Get a BlueprintDefinition by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		sd, err := client.GetBlueprintDefinition(ColonyName, BlueprintDefinitionName, PrvKey)
		CheckError(err)

		if sd == nil {
			CheckError(errors.New("BlueprintDefinition not found"))
		}

		log.WithFields(log.Fields{
			"BlueprintDefinitionID": sd.ID,
			"Name":                 sd.Metadata.Name,
			"Kind":                 sd.Spec.Names.Kind,
			"ColonyName":           sd.Metadata.ColonyName,
		}).Info("BlueprintDefinition retrieved")

		if JSON {
			jsonString, err := sd.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
		} else {
			printBlueprintDefinitionTable(sd)
		}
	},
}

var listBlueprintDefinitionsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List BlueprintDefinitions",
	Long:  "List all BlueprintDefinitions in the colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		sds, err := client.GetBlueprintDefinitions(ColonyName, PrvKey)
		CheckError(err)

		if len(sds) == 0 {
			log.Info("No blueprint definitions found")
			return
		}

		log.WithFields(log.Fields{
			"Count":      len(sds),
			"ColonyName": ColonyName,
		}).Info("BlueprintDefinitions retrieved")

		if JSON {
			// Print as JSON array
			jsonBytes, err := json.MarshalIndent(sds, "", "  ")
			CheckError(err)
			fmt.Println(string(jsonBytes))
		} else {
			// Print as table
			printBlueprintDefinitionsTable(sds)
		}
	},
}

var removeBlueprintDefinitionCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a BlueprintDefinition",
	Long:  "Remove a BlueprintDefinition by name (requires colony owner privileges)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveBlueprintDefinition(ColonyName, BlueprintDefinitionName, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"Name":       BlueprintDefinitionName,
			"ColonyName": ColonyName,
		}).Info("BlueprintDefinition removed")
	},
}

var addBlueprintCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a Blueprint",
	Long:  "Add a custom blueprint instance",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonBytes, err := os.ReadFile(SpecFile)
		CheckError(err)

		var blueprint core.Blueprint
		err = json.Unmarshal(jsonBytes, &blueprint)
		CheckError(err)

		// Set namespace (colony name) if not specified
		if blueprint.Metadata.ColonyName == "" {
			blueprint.Metadata.ColonyName = ColonyName
		}

		addedBlueprint, err := client.AddBlueprint(&blueprint, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"BlueprintID": addedBlueprint.ID,
			"Name":       addedBlueprint.Metadata.Name,
			"Kind":       addedBlueprint.Kind,
			"Namespace":  addedBlueprint.Metadata.ColonyName,
		}).Info("Blueprint added")

	},
}

var getBlueprintCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a Blueprint",
	Long:  "Get a blueprint by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		blueprint, err := client.GetBlueprint(ColonyName, BlueprintName, PrvKey)
		CheckError(err)

		if blueprint == nil {
			CheckError(errors.New("Blueprint not found"))
		}

		log.WithFields(log.Fields{
			"BlueprintID": blueprint.ID,
			"Name":       blueprint.Metadata.Name,
			"Kind":       blueprint.Kind,
			"Namespace":  blueprint.Metadata.ColonyName,
		}).Info("Blueprint retrieved")

		if JSON {
			jsonString, err := blueprint.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
		} else {
			printBlueprintTable(client, blueprint)
		}
	},
}

var listBlueprintsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Blueprints",
	Long:  "List all blueprints in the colony (optionally filtered by kind)",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		blueprints, err := client.GetBlueprints(ColonyName, Kind, PrvKey)
		CheckError(err)

		if len(blueprints) == 0 {
			log.Info("No blueprints found")
			return
		}

		log.WithFields(log.Fields{
			"Count":      len(blueprints),
			"ColonyName": ColonyName,
			"Kind":       Kind,
		}).Info("Blueprints retrieved")

		if JSON {
			// Print as JSON array
			jsonBytes, err := json.MarshalIndent(blueprints, "", "  ")
			CheckError(err)
			fmt.Println(string(jsonBytes))
		} else {
			// Print as table with client for runtime information
			printBlueprintsTableWithClient(client, blueprints)
		}
	},
}

var updateBlueprintCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Blueprint",
	Long:  "Update an existing blueprint",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonBytes, err := os.ReadFile(SpecFile)
		CheckError(err)

		var blueprint core.Blueprint
		err = json.Unmarshal(jsonBytes, &blueprint)
		CheckError(err)

		// Set namespace (colony name) if not specified
		if blueprint.Metadata.ColonyName == "" {
			blueprint.Metadata.ColonyName = ColonyName
		}

		updatedBlueprint, err := client.UpdateBlueprint(&blueprint, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"BlueprintID": updatedBlueprint.ID,
			"Name":       updatedBlueprint.Metadata.Name,
			"Kind":       updatedBlueprint.Kind,
			"Generation": updatedBlueprint.Metadata.Generation,
		}).Info("Blueprint updated")

	},
}

var setBlueprintCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a field value in a Blueprint",
	Long:  "Set a specific field value in a blueprint using dot notation (e.g., 'replicas' or 'env.TZ')",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		// Get the existing blueprint
		blueprint, err := client.GetBlueprint(ColonyName, BlueprintName, PrvKey)
		CheckError(err)

		// Parse the blueprint spec into a map for easy manipulation
		specMap := make(map[string]interface{})
		specBytes, err := json.Marshal(blueprint.Spec)
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
		// We don't allow creating new fields to prevent corrupting the blueprint
		current := specMap
		for i := 0; i < len(keyParts)-1; i++ {
			key := keyParts[i]
			if _, ok := current[key]; !ok {
				CheckError(errors.New("Invalid key path: '" + Key + "' (field '" + key + "' does not exist in blueprint spec)"))
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
			CheckError(errors.New("Invalid key: '" + Key + "' (field '" + finalKey + "' does not exist in blueprint spec)"))
		}

		// Try to parse the value as JSON to support numbers, booleans, etc.
		var parsedValue interface{}
		err = json.Unmarshal([]byte(Value), &parsedValue)
		if err != nil {
			// If it's not valid JSON, treat it as a string
			parsedValue = Value
		}

		current[finalKey] = parsedValue

		// Update the blueprint spec
		blueprint.Spec = specMap

		// Update the blueprint in the colony
		updatedBlueprint, err := client.UpdateBlueprint(blueprint, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"BlueprintID": updatedBlueprint.ID,
			"Name":       updatedBlueprint.Metadata.Name,
			"Kind":       updatedBlueprint.Kind,
			"Key":        Key,
			"Value":      Value,
		}).Info("Blueprint field updated")

		// Trigger immediate reconciliation (idempotent operation)
		process, err := client.ReconcileBlueprint(ColonyName, BlueprintName, false, PrvKey)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Warn("Failed to trigger reconciliation")
		} else {
			log.WithFields(log.Fields{
				"ProcessID": process.ID,
			}).Info("Reconciliation triggered")
		}
	},
}

var removeBlueprintCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Blueprint",
	Long:  "Remove a blueprint by name",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveBlueprint(ColonyName, BlueprintName, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"Name":       BlueprintName,
			"ColonyName": ColonyName,
		}).Info("Blueprint removed")
	},
}

var reconcileBlueprintCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Trigger immediate reconciliation of a blueprint",
	Long:  "Trigger immediate reconciliation of a blueprint. Use --force to recreate all containers with fresh images.",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		// Call server-side reconciliation which looks up executor type from the blueprint handler
		process, err := client.ReconcileBlueprint(ColonyName, BlueprintName, Force, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"BlueprintName": BlueprintName,
			"ProcessID":     process.ID,
			"Force":         Force,
		}).Info("Submitted reconciliation process")

		fmt.Printf("Reconciliation process submitted: %s\n", process.ID)
		fmt.Printf("Use 'colonies log get -p %s' to view progress\n", process.ID)
	},
}

var historyBlueprintCmd = &cobra.Command{
	Use:   "history",
	Short: "Show blueprint history",
	Long:  "Display the history of changes to a blueprint",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		// Get the blueprint to find its ID
		blueprint, err := client.GetBlueprint(ColonyName, BlueprintName, PrvKey)
		CheckError(err)

		// Get the blueprint history
		histories, err := client.GetBlueprintHistory(blueprint.ID, Count, PrvKey)
		CheckError(err)

		if len(histories) == 0 {
			log.Info("No history found for this blueprint")
			return
		}

		// If generation is specified, show detailed view of that generation
		if Generation >= 0 {
			var selectedHistory *core.BlueprintHistory
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
				printBlueprintHistoryDetail(selectedHistory)
			}
		} else {
			// Print history table
			if JSON {
				jsonString, err := core.ConvertBlueprintHistoryArrayToJSON(histories)
				CheckError(err)
				fmt.Println(jsonString)
			} else {
				printBlueprintHistoryTable(client, histories)
			}
		}
	},
}

var logBlueprintCmd = &cobra.Command{
	Use:   "log",
	Short: "Show logs from blueprint reconcilers",
	Long:  "Display logs from reconciler executors. Optionally filter by blueprint name to show logs from the specific reconciler handling that blueprint's location.",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		theme, err := table.LoadTheme("solarized-dark")
		CheckError(err)

		var reconcilerNames []string

		if BlueprintName != "" {
			// Get the blueprint to find its location and handler type
			blueprint, err := client.GetBlueprint(ColonyName, BlueprintName, PrvKey)
			CheckError(err)

			locationName := blueprint.Metadata.LocationName
			executorType := ""

			// Get handler type from blueprint or its definition
			if blueprint.Handler != nil && blueprint.Handler.ExecutorType != "" {
				executorType = blueprint.Handler.ExecutorType
			}

			// Find reconciler executors at this location
			executors, err := client.GetExecutors(ColonyName, PrvKey)
			CheckError(err)

			for _, exec := range executors {
				if exec.LocationName == locationName {
					if executorType == "" || exec.Type == executorType {
						reconcilerNames = append(reconcilerNames, exec.Name)
					}
				}
			}

			if len(reconcilerNames) == 0 {
				log.WithFields(log.Fields{
					"BlueprintName": BlueprintName,
					"Location":      locationName,
					"ExecutorType":  executorType,
				}).Warn("No reconciler found for this blueprint's location")
				return
			}

			log.WithFields(log.Fields{
				"BlueprintName": BlueprintName,
				"Location":      locationName,
				"Reconcilers":   reconcilerNames,
			}).Info("Fetching logs from reconcilers")
		} else {
			// Get all reconciler-type executors
			executors, err := client.GetExecutors(ColonyName, PrvKey)
			CheckError(err)

			for _, exec := range executors {
				if strings.Contains(exec.Type, "reconciler") {
					reconcilerNames = append(reconcilerNames, exec.Name)
				}
			}

			if len(reconcilerNames) == 0 {
				log.Info("No reconciler executors found")
				return
			}

			log.WithFields(log.Fields{
				"Reconcilers": reconcilerNames,
			}).Info("Fetching logs from all reconcilers")
		}

		if Follow {
			// Follow mode - continuously fetch new logs
			lastTimestamps := make(map[string]int64)
			for _, name := range reconcilerNames {
				lastTimestamps[name] = 0
			}

			for {
				for _, reconcilerName := range reconcilerNames {
					logs, err := client.GetLogsByExecutorSince(ColonyName, reconcilerName, Count, lastTimestamps[reconcilerName], PrvKey)
					if err != nil {
						continue
					}
					for _, logEntry := range logs {
						prefix := termenv.String("[" + reconcilerName + "] ").Foreground(theme.ColorCyan).String()
						fmt.Print(prefix + formatLogMessage(logEntry.Message, theme))
					}
					if len(logs) > 0 {
						lastTimestamps[reconcilerName] = logs[len(logs)-1].Timestamp
					}
				}
				time.Sleep(1 * time.Second)
			}
		} else {
			// One-shot mode - fetch latest logs from each reconciler
			for _, reconcilerName := range reconcilerNames {
				logs, err := client.GetLogsByExecutorLatest(ColonyName, reconcilerName, Count, PrvKey)
				if err != nil {
					log.WithFields(log.Fields{
						"Reconciler": reconcilerName,
						"Error":      err,
					}).Warn("Failed to fetch logs")
					continue
				}

				if len(logs) > 0 {
					fmt.Println(termenv.String("=== " + reconcilerName + " ===").Foreground(theme.ColorViolet).Bold().String())
					for _, logEntry := range logs {
						fmt.Print(formatLogMessage(logEntry.Message, theme))
					}
					fmt.Println()
				}
			}
		}
	},
}

var doctorBlueprintCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose blueprint configuration issues",
	Long:  "Analyze blueprint configuration and identify potential problems like missing reconcilers, location mismatches, or configuration errors.",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		theme, err := table.LoadTheme("solarized-dark")
		CheckError(err)

		var blueprintsToCheck []*core.Blueprint

		if BlueprintName != "" {
			blueprint, err := client.GetBlueprint(ColonyName, BlueprintName, PrvKey)
			CheckError(err)
			blueprintsToCheck = append(blueprintsToCheck, blueprint)
		} else {
			blueprints, err := client.GetBlueprints(ColonyName, "", PrvKey)
			CheckError(err)
			blueprintsToCheck = blueprints
		}

		// Get all executors for checking
		executors, err := client.GetExecutors(ColonyName, PrvKey)
		CheckError(err)

		// Build location -> reconciler map
		reconcilersByLocation := make(map[string][]*core.Executor)
		reconcilersByType := make(map[string][]*core.Executor)
		for _, exec := range executors {
			if strings.Contains(exec.Type, "reconciler") {
				reconcilersByLocation[exec.LocationName] = append(reconcilersByLocation[exec.LocationName], exec)
				reconcilersByType[exec.Type] = append(reconcilersByType[exec.Type], exec)
			}
		}

		issuesFound := 0

		for _, blueprint := range blueprintsToCheck {
			fmt.Println(termenv.String("=== " + blueprint.Metadata.Name + " ===").Foreground(theme.ColorViolet).Bold().String())

			locationName := blueprint.Metadata.LocationName
			handlerType := ""
			if blueprint.Handler != nil {
				handlerType = blueprint.Handler.ExecutorType
			}

			// Check 1: Does a reconciler exist at this location?
			reconcilersAtLocation := reconcilersByLocation[locationName]
			if len(reconcilersAtLocation) == 0 {
				issuesFound++
				fmt.Print(termenv.String("  [ERROR] ").Foreground(theme.ColorRed).Bold().String())
				fmt.Printf("No reconciler found at location '%s'\n", locationName)

				// Suggest available locations
				if len(reconcilersByLocation) > 0 {
					fmt.Print(termenv.String("          ").String())
					fmt.Print("Available locations with reconcilers: ")
					locs := []string{}
					for loc := range reconcilersByLocation {
						locs = append(locs, loc)
					}
					fmt.Println(strings.Join(locs, ", "))
				}
			} else {
				fmt.Print(termenv.String("  [OK] ").Foreground(theme.ColorGreen).String())
				fmt.Printf("Reconciler found at location '%s': %s\n", locationName, reconcilersAtLocation[0].Name)
			}

			// Check 2: Does the handler type match an available reconciler?
			if handlerType != "" {
				reconcilersOfType := reconcilersByType[handlerType]
				if len(reconcilersOfType) == 0 {
					issuesFound++
					fmt.Print(termenv.String("  [ERROR] ").Foreground(theme.ColorRed).Bold().String())
					fmt.Printf("No executor of type '%s' found\n", handlerType)

					// Suggest available types
					if len(reconcilersByType) > 0 {
						fmt.Print(termenv.String("          ").String())
						fmt.Print("Available reconciler types: ")
						types := []string{}
						for t := range reconcilersByType {
							types = append(types, t)
						}
						fmt.Println(strings.Join(types, ", "))
					}
				} else {
					// Check if any reconciler of this type is at the right location
					matchFound := false
					for _, rec := range reconcilersOfType {
						if rec.LocationName == locationName {
							matchFound = true
							break
						}
					}
					if !matchFound {
						issuesFound++
						fmt.Print(termenv.String("  [ERROR] ").Foreground(theme.ColorRed).Bold().String())
						fmt.Printf("Reconciler type '%s' exists but not at location '%s'\n", handlerType, locationName)
						fmt.Print(termenv.String("          ").String())
						fmt.Print("Reconcilers of this type are at: ")
						locs := []string{}
						for _, rec := range reconcilersOfType {
							locs = append(locs, rec.LocationName)
						}
						fmt.Println(strings.Join(locs, ", "))
					} else {
						fmt.Print(termenv.String("  [OK] ").Foreground(theme.ColorGreen).String())
						fmt.Printf("Handler type '%s' matches reconciler at location\n", handlerType)
					}
				}
			}

			// Check 3: Reconciler heartbeat (is it actively connected?)
			if len(reconcilersAtLocation) > 0 {
				rec := reconcilersAtLocation[0]
				timeSinceHeard := time.Since(rec.LastHeardFromTime)
				if timeSinceHeard > 2*time.Minute {
					issuesFound++
					fmt.Print(termenv.String("  [WARN] ").Foreground(theme.ColorYellow).Bold().String())
					fmt.Printf("Reconciler '%s' last heard from %s ago (may be offline)\n", rec.Name, timeSinceHeard.Round(time.Second))
				} else {
					fmt.Print(termenv.String("  [OK] ").Foreground(theme.ColorGreen).String())
					fmt.Printf("Reconciler '%s' is active (last heard %s ago)\n", rec.Name, timeSinceHeard.Round(time.Second))
				}
			}

			// Check 4: Replica status (for ExecutorDeployment kind)
			if blueprint.Kind == "ExecutorDeployment" {
				desiredReplicas := -1
				if val, ok := blueprint.GetSpec("replicas"); ok {
					if floatVal, ok := val.(float64); ok {
						desiredReplicas = int(floatVal)
					} else if intVal, ok := val.(int); ok {
						desiredReplicas = intVal
					}
				}

				if desiredReplicas > 0 {
					// Get actual running replicas using BlueprintID matching (same as table)
					runningCount := 0
					currentGen := blueprint.Metadata.Generation
					for _, exec := range executors {
						// Only count approved executors
						if !exec.IsApproved() {
							continue
						}
						// Try BlueprintID match first (more reliable)
						if exec.BlueprintID == blueprint.ID {
							if exec.BlueprintGen == currentGen {
								runningCount++
							}
						} else if exec.BlueprintID == "" {
							// Fallback to name-based matching for executors without BlueprintID
							blueprintHyphens := strings.Count(blueprint.Metadata.Name, "-")
							executorParts := strings.Split(exec.Name, "-")
							expectedParts := blueprintHyphens + 3
							if len(executorParts) == expectedParts {
								deploymentName := strings.Join(executorParts[:blueprintHyphens+1], "-")
								if deploymentName == blueprint.Metadata.Name {
									lastPart := executorParts[len(executorParts)-1]
									var gen int
									_, err := fmt.Sscanf(lastPart, "%d", &gen)
									if err == nil {
										if int64(gen) == currentGen {
											runningCount++
										}
									} else {
										runningCount++
									}
								}
							}
						}
					}

					if runningCount < desiredReplicas {
						issuesFound++
						fmt.Print(termenv.String("  [WARN] ").Foreground(theme.ColorYellow).Bold().String())
						fmt.Printf("Only %d/%d replicas running\n", runningCount, desiredReplicas)
						// Add actionable suggestions
						fmt.Print(termenv.String("          ").String())
						fmt.Printf("Run: colonies blueprint log --name %s\n", blueprint.Metadata.Name)
						fmt.Print(termenv.String("          ").String())
						fmt.Printf("Try: colonies blueprint reconcile --name %s --force\n", blueprint.Metadata.Name)
					} else if runningCount == desiredReplicas {
						fmt.Print(termenv.String("  [OK] ").Foreground(theme.ColorGreen).String())
						fmt.Printf("All %d replicas running\n", desiredReplicas)
					}
				}
			}

			fmt.Println()
		}

		// Summary
		if issuesFound == 0 {
			fmt.Println(termenv.String("No issues found!").Foreground(theme.ColorGreen).Bold().String())
		} else {
			fmt.Println(termenv.String(fmt.Sprintf("Found %d issue(s)", issuesFound)).Foreground(theme.ColorRed).Bold().String())
			fmt.Println()
			fmt.Println(termenv.String("Hints:").Bold().String())
			fmt.Println("  - Check reconciler logs: colonies blueprint log --name <blueprint>")
			fmt.Println("  - Force reconciliation:  colonies blueprint reconcile --name <blueprint> --force")
			fmt.Println("  - List all executors:    colonies executor ls")
		}
	},
}
