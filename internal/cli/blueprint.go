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

		// Get blueprint
		blueprint, err := client.GetBlueprint(ColonyName, BlueprintName, PrvKey)
		CheckError(err)

		// If force flag is set, bump the generation first
		if Force {
			log.Info("Force flag enabled - bumping blueprint generation to trigger redeployment")

			// Update the blueprint with force generation bump
			updatedBlueprint, err := client.UpdateBlueprintWithForce(blueprint, true, PrvKey)
			CheckError(err)

			log.WithFields(log.Fields{
				"BlueprintName": BlueprintName,
				"OldGeneration": blueprint.Metadata.Generation,
				"NewGeneration": updatedBlueprint.Metadata.Generation,
			}).Info("Blueprint generation bumped")

			// Use the updated blueprint for reconciliation
			blueprint = updatedBlueprint
		}

		// Get the blueprint definition to find the handler
		sd, err := client.GetBlueprintDefinitionByKind(ColonyName, blueprint.Kind, PrvKey)
		CheckError(err)

		if sd == nil || sd.Spec.Handler.ExecutorType == "" {
			CheckError(errors.New("Blueprint kind has no handler defined"))
		}

		// Submit a direct reconciliation process
		funcSpec := core.CreateEmptyFunctionSpec()
		funcSpec.NodeName = "reconcile"
		funcSpec.Conditions.ColonyName = ColonyName
		funcSpec.Conditions.ExecutorType = sd.Spec.Handler.ExecutorType
		funcSpec.FuncName = "reconcile"
		funcSpec.KwArgs = map[string]interface{}{
			"kind": blueprint.Kind,
		}

		// Apply executor targeting if the blueprint has a handler with specific executor names
		if blueprint.Handler != nil {
			if len(blueprint.Handler.ExecutorNames) > 0 {
				funcSpec.Conditions.ExecutorNames = blueprint.Handler.ExecutorNames
			} else if blueprint.Handler.ExecutorName != "" {
				funcSpec.Conditions.ExecutorNames = []string{blueprint.Handler.ExecutorName}
			}
		}

		// Pass force flag to reconciler so it can pull fresh images
		if Force {
			funcSpec.KwArgs["force"] = true
		}

		process, err := client.Submit(funcSpec, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"BlueprintName": BlueprintName,
			"Kind":          blueprint.Kind,
			"Generation":    blueprint.Metadata.Generation,
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
