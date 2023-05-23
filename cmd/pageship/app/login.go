package app

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/oursky/pageship/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login user",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := authGitHubSSH(cmd.Context())
		if err != nil {
			Error("Failed to login via SSH: %s", err)
			return
		}

		claims := &models.TokenClaims{}
		_, _, err = jwt.NewParser().ParseUnverified(token, claims)
		if err != nil {
			Error("Server returned malformed token: %s", err)
			return
		}
		userName := claims.Username

		err = saveToken(token)
		if err != nil {
			Error("Failed to save token: %s", err)
			return
		}

		Info("Logged in as %q.", userName)
	},
}
