package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/sshkey"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login user",
	Run: func(cmd *cobra.Command, args []string) {
		var sources []sshkey.Source
		defer func() {
			for _, s := range sources {
				s.Close()
			}
		}()

		agent, err := sshkey.NewAgent()
		if err != nil {
			Debug("Failed to connect to SSH agent: %s", err)
		} else {
			sources = append(sources, agent)
		}

		sources = append(sources, sshkey.NewUserKey(
			func(def string) (string, error) {
				prompt := promptui.Prompt{
					Label:   "SSH key file",
					Default: def,
					Validate: func(s string) error {
						_, err := os.Stat(s)
						if err != nil {
							return fmt.Errorf("invalid keyfile: %w", err)
						}
						return nil
					},
				}
				return prompt.Run()
			},
			func() ([]byte, error) {
				prompt := promptui.Prompt{
					Label:       "SSH key passphrase",
					Mask:        '*',
					HideEntered: true,
				}
				result, err := prompt.Run()
				if err != nil {
					return nil, err
				}
				return []byte(result), nil
			},
		))

		conf, err := config.LoadClientConfig()
		if err != nil {
			Error("Failed to load config: %s", err)
			return
		}
		userName := conf.GitHubUsername

		if userName == "" {
			prompt := promptui.Prompt{
				Label: "GitHub user name",
				Validate: func(s string) error {
					if len(s) == 0 {
						return errors.New("must enter user name")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			if err != nil {
				Info("Cancelled.")
				return
			}

			userName = result

			conf.GitHubUsername = userName
			_ = conf.Save()
		}

		sshConf := &ssh.ClientConfig{
			User: userName,
			Auth: []ssh.AuthMethod{
				sshkey.AuthMethodSources(sources),
			},
			// This connection is tunneled through WebSocket (with TLS),
			// so no need verify host key.
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		ws, err := apiClient.OpenAuthGitHubSSH(cmd.Context())
		if err != nil {
			Error("Failed to connect to server: %s", err)
			return
		}

		sshConn, _, _, err := ssh.NewClientConn(ws, "", sshConf)
		if err != nil {
			Error("Failed to authenticate with server: %s", err)
			return
		}
		defer sshConn.Close()

		ok, reply, err := sshConn.SendRequest("pageship", true, nil)
		if !ok || err != nil {
			Error("Failed to request token: %s", err)
			return
		}

		token := string(reply)
		claims := &models.TokenClaims{}
		_, _, err = jwt.NewParser().ParseUnverified(token, claims)
		if err != nil {
			Error("server returned malformed token")
			return
		}
		userName = claims.Username

		conf, err = config.LoadClientConfig()
		if err != nil {
			Error("Failed to load config: %s", err)
			return
		}

		conf.AuthToken = token
		err = conf.Save()
		if err != nil {
			Error("Failed to save auth token")
			return
		}

		Info("Logged in as %q.", userName)
	},
}
