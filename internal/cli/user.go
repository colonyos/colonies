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
	userCmd.AddCommand(addUserCmd)
	userCmd.AddCommand(chUserIDCmd)
	userCmd.AddCommand(listUsersCmd)
	userCmd.AddCommand(getUserCmd)
	userCmd.AddCommand(removeUserCmd)
	rootCmd.AddCommand(userCmd)

	userCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	userCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addUserCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	addUserCmd.Flags().StringVarP(&Username, "name", "", "", "Username")
	addUserCmd.MarkFlagRequired("name")
	addUserCmd.Flags().StringVarP(&UserID, "userid", "", "", "User Id")
	addUserCmd.MarkFlagRequired("userid")
	addUserCmd.Flags().StringVarP(&Email, "email", "", "", "Email")
	addUserCmd.Flags().StringVarP(&Phone, "phone", "", "", "Phone")

	chUserIDCmd.Flags().StringVarP(&UserID, "userid", "", "", "User Id")
	chUserIDCmd.MarkFlagRequired("userid")

	getUserCmd.Flags().StringVarP(&Username, "name", "", "", "Username")
	getUserCmd.MarkFlagRequired("name")

	removeUserCmd.Flags().StringVarP(&Username, "name", "", "", "Username")
	removeUserCmd.MarkFlagRequired("name")
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  "Manage users",
}

var addUserCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new User",
	Long:  "Add a new User",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(UserID) != 64 {
			CheckError(errors.New("Invalid user Id length"))
		}

		if Username == "" {
			CheckError(errors.New("Username must be specified"))
		}

		user := core.CreateUser(ColonyName, UserID, Username, Email, Phone)

		if ColonyPrvKey == "" {
			CheckError(errors.New("You must specify a Colony private key by exporting COLONIES_COLONY_PRVKEY"))
		}

		_, err := client.AddUser(user, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ColonyName": ColonyName,
			"UserId":     UserID,
			"Username":   Username,
			"Email":      Email,
			"Phone":      Phone}).
			Info("User added")
	},
}

var chUserIDCmd = &cobra.Command{
	Use:   "chid",
	Short: "Change user Id",
	Long:  "Change user Id",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(UserID) != 64 {
			CheckError(errors.New("Invalid user Id length"))
		}

		err := client.ChangeUserID(ColonyName, UserID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ColonyName": ColonyName,
			"UserId":     UserID}).
			Info("Changed user Id")
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Users member of a Colony",
	Long:  "List Users member of a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		usersFromServer, err := client.GetUsers(ColonyName, PrvKey)
		CheckError(err)

		if len(usersFromServer) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No users found")
			os.Exit(0)
		}

		printUsersTable(usersFromServer)
	},
}

var getUserCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a User",
	Long:  "Get info about a User",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		user, err := client.GetUser(ColonyName, Username, PrvKey)
		CheckError(err)

		if user == nil {
			CheckError(errors.New("User not found"))
		}

		printUserTable(user)
	},
}

var removeUserCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a User from a Colony",
	Long:  "Remove a User from a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		fmt.Print("WARNING!!! Are you sure you want to remove user <" + Username + "> in colony <" + ColonyName + "> ! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			err := client.RemoveUser(ColonyName, Username, ColonyPrvKey)
			CheckError(err)
		} else {
			fmt.Println("Aborting ...")
		}

		log.WithFields(log.Fields{
			"ColonyName": ColonyName,
			"Username":   Username}).
			Info("User removed")
	},
}
