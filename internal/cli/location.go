package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	locationCmd.AddCommand(addLocationCmd)
	locationCmd.AddCommand(listLocationsCmd)
	locationCmd.AddCommand(getLocationCmd)
	locationCmd.AddCommand(removeLocationCmd)
	rootCmd.AddCommand(locationCmd)

	locationCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	locationCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addLocationCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	addLocationCmd.Flags().StringVarP(&LocationName, "name", "", "", "Location name")
	addLocationCmd.MarkFlagRequired("name")
	addLocationCmd.Flags().StringVarP(&LocationDesc, "desc", "", "", "Location description")
	addLocationCmd.Flags().Float64VarP(&Long, "long", "", 0, "Longitude")
	addLocationCmd.Flags().Float64VarP(&Lat, "lat", "", 0, "Latitude")

	getLocationCmd.Flags().StringVarP(&LocationName, "name", "", "", "Location name")
	getLocationCmd.MarkFlagRequired("name")

	removeLocationCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	removeLocationCmd.Flags().StringVarP(&LocationName, "name", "", "", "Location name")
	removeLocationCmd.MarkFlagRequired("name")
}

var locationCmd = &cobra.Command{
	Use:   "location",
	Short: "Manage locations",
	Long:  "Manage locations",
}

var addLocationCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new location",
	Long:  "Add a new location",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if LocationName == "" {
			CheckError(errors.New("Location name must be specified"))
		}

		locationID := core.GenerateRandomID()
		location := core.CreateLocation(locationID, LocationName, ColonyName, LocationDesc, Long, Lat)

		if ColonyPrvKey == "" {
			CheckError(errors.New("You must specify a Colony private key by exporting COLONIES_COLONY_PRVKEY"))
		}

		addedLocation, err := client.AddLocation(location, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ColonyName":  ColonyName,
			"Name":        addedLocation.Name,
			"LocationID":  addedLocation.ID,
			"Description": addedLocation.Description,
			"Long":        addedLocation.Long,
			"Lat":         addedLocation.Lat}).
			Info("Location added")
	},
}

var listLocationsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List locations in a colony",
	Long:  "List locations in a colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		locationsFromServer, err := client.GetLocations(ColonyName, PrvKey)
		CheckError(err)

		if len(locationsFromServer) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No locations found")
			os.Exit(0)
		}

		printLocationsTable(locationsFromServer)
	},
}

var getLocationCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a location",
	Long:  "Get info about a location",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		location, err := client.GetLocation(ColonyName, LocationName, PrvKey)
		CheckError(err)

		if location == nil {
			CheckError(errors.New("Location not found"))
		}

		printLocationTable(location)
	},
}

var removeLocationCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a location from a colony",
	Long:  "Remove a location from a colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if ColonyPrvKey == "" {
			CheckError(errors.New("You must specify a Colony private key by exporting COLONIES_COLONY_PRVKEY"))
		}

		fmt.Print("WARNING!!! Are you sure you want to remove location <" + LocationName + "> in colony <" + ColonyName + ">? (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			err := client.RemoveLocation(ColonyName, LocationName, ColonyPrvKey)
			CheckError(err)

			log.WithFields(log.Fields{
				"ColonyName":   ColonyName,
				"LocationName": LocationName}).
				Info("Location removed")
		} else {
			fmt.Println("Aborting ...")
		}
	},
}
