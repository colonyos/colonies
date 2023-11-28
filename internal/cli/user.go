package cli

import (
	"errors"
	"os"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	userCmd.AddCommand(addUserCmd)
	userCmd.AddCommand(listUsersCmd)
	userCmd.AddCommand(removeUserCmd)
	rootCmd.AddCommand(userCmd)

	userCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	userCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addUserCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	addUserCmd.Flags().StringVarP(&Username, "username", "", "", "Username")
	addUserCmd.MarkFlagRequired("username")
	addUserCmd.Flags().StringVarP(&UserID, "userid", "", "", "User Id")
	addUserCmd.MarkFlagRequired("userid")
	addUserCmd.Flags().StringVarP(&Email, "email", "", "", "Email")
	addUserCmd.Flags().StringVarP(&Phone, "phone", "", "", "Phone")

	removeUserCmd.Flags().StringVarP(&Username, "username", "", "", "Username")
	removeUserCmd.MarkFlagRequired("username")
}

var userCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  "Manage users",
}

var addUserCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user",
	Long:  "Add a new user",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(UserID) != 64 {
			CheckError(errors.New("Invalid Eser Id length"))
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

var listUsersCmd = &cobra.Command{
	Use:   "ls",
	Short: "List users member of a Colony",
	Long:  "List users member of a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		usersFromServer, err := client.GetUsers(ColonyName, PrvKey)
		CheckError(err)

		if len(usersFromServer) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No users found")
			os.Exit(0)
		}

		var data [][]string
		for _, user := range usersFromServer {
			data = append(data, []string{user.Name, user.Email, user.Phone})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Username", "Email", "Phone"})

		for _, v := range data {
			table.Append(v)
		}

		table.Render()
	},
}

var removeUserCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a user from a Colony",
	Long:  "Remove a user from a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.DeleteUser(ColonyName, Username, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ColonyName": ColonyName,
			"Username":   Username}).
			Info("User removed")
	},
}
