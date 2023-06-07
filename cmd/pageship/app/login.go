package app

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := ensureAuth(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to login: %w", err)
		}

		claims := &models.TokenClaims{}
		_, _, err = jwt.NewParser().ParseUnverified(token, claims)
		if err != nil {
			return fmt.Errorf("server returned malformed token: %w", err)
		}
		name := claims.Name

		err = saveToken(token)
		if err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		Info("Logged in as %q. (id: %q)", name, claims.Subject)
		return nil
	},
}
