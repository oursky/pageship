package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/sshkey"
	"golang.org/x/crypto/ssh"
)

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
			return "", ErrCancelled
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
