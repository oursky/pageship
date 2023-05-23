package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/sshkey"
	"golang.org/x/crypto/ssh"
)

const reauthThreshold time.Duration = time.Minute * 10

func ensureAuth(ctx context.Context) (string, error) {
	conf, err := config.LoadClientConfig()
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}

	token := conf.AuthToken
	if isTokenValid(token) == nil {
		return token, nil
	}

	token, err = authGitHubSSH(ctx)

	if err == nil {
		_ = saveToken(token)
	}

	return token, err
}

func saveToken(token string) error {
	conf, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	conf.AuthToken = token
	err = conf.Save()
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

func isTokenValid(token string) error {
	claims := &models.TokenClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return err
	}
	if claims.ExpiresAt.Sub(time.Now()) < reauthThreshold {
		// Expires soon, need reauth
		return models.ErrInvalidCredentials
	}
	return nil
}

func authGitHubSSH(ctx context.Context) (string, error) {
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
		return "", fmt.Errorf("load config: %w", err)
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
			return "", nil
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

	ws, err := apiClient.OpenAuthGitHubSSH(ctx)
	if err != nil {
		return "", fmt.Errorf("connect to server: %w", err)
	}

	sshConn, _, _, err := ssh.NewClientConn(ws, "", sshConf)
	if err != nil {
		return "", fmt.Errorf("authenticate with server: %w", err)
	}
	defer sshConn.Close()

	ok, reply, err := sshConn.SendRequest("pageship", true, nil)
	if err != nil {
		return "", fmt.Errorf("request token: %w", err)
	} else if !ok {
		return "", fmt.Errorf("request token failed")
	}

	token := string(reply)
	return token, nil
}
