package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	userCmd.AddCommand(addUserCmd)
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

	getUserCmd.Flags().StringVarP(&Username, "name", "", "", "Username")
	getUserCmd.MarkFlagRequired("name")

	removeUserCmd.Flags().StringVarP(&Username, "name", "", "", "Username")
	removeUserCmd.MarkFlagRequired("name")

}

var userCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage Users",
	Long:  "Manage Users",
}

var addUserCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new User",
	Long:  "Add a new User",
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

var getUserCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a User",
	Long:  "Get info about a User",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		user, err := client.GetUser(ColonyName, Username, PrvKey)
		CheckError(err)

		userData := [][]string{
			[]string{"Name", user.Name},
			[]string{"ID", user.ID},
			[]string{"ColonyName", user.ColonyName},
			[]string{"Email", user.Email},
			[]string{"Phone", user.Phone},
		}

		userTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range userData {
			userTable.Append(v)
		}
		userTable.SetAlignment(tablewriter.ALIGN_LEFT)
		userTable.Render()
	},
}

var removeUserCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a User from a Colony",
	Long:  "Remove a User from a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		fmt.Print("WARNING!!! Are you sure you want to delete the User <" + Username + "> from Colony <" + ColonyName + "> ! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			err := client.DeleteUser(ColonyName, Username, ColonyPrvKey)
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
