package app

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersAddCmd)
	usersCmd.AddCommand(usersDelCmd)

	usersCmd.PersistentFlags().String("app", "", "app ID")
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			Error("App ID is not set")
			return
		}

		users, err := apiClient.ListUsers(cmd.Context(), appID)
		if err != nil {
			Error("Failed to list users: %s", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME")
		for _, u := range users {
			fmt.Fprintf(w, "%s\t%s\n", u.ID, u.Name)
		}
		w.Flush()
	},
}

var usersAddCmd = &cobra.Command{
	Use:   "add user-id",
	Short: "Add user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			Error("App ID is not set")
			return
		}

		userID := args[0]

		err := apiClient.AddUser(cmd.Context(), appID, userID)
		if err != nil {
			Error("Failed to add users: %s", err)
			return
		}

		Info("Done!")
	},
}

var usersDelCmd = &cobra.Command{
	Use:   "delete user-id",
	Short: "Delete user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			Error("App ID is not set")
			return
		}

		userID := args[0]

		err := apiClient.DeleteUser(cmd.Context(), appID, userID)
		if err != nil {
			Error("Failed to delete users: %s", err)
			return
		}

		Info("Done!")
	},
}
