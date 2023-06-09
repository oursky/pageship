package controller

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

var sshHostKey ssh.Signer

func init() {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	sshHostKey, err = ssh.NewSignerFromSigner(key)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) handleAuthGithubSSH(w http.ResponseWriter, r *http.Request) {
	s := websocket.Server{Handler: c.handleAuthGithubSSHConn}
	s.ServeHTTP(w, r)
}

func (c *Controller) handleAuthGithubSSHConn(conn *websocket.Conn) {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(meta ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			fingerprint := ssh.FingerprintSHA256(pubKey)
			pubKeys, err := c.githubKeys.PublicKey(meta.User())
			if err != nil {
				log(conn.Request()).Warn("cannot get GitHub public key",
					zap.String("user", meta.User()),
					zap.Error(err),
				)
				return nil, err
			}

			if _, ok := pubKeys[string(pubKey.Marshal())]; !ok {
				log(conn.Request()).Debug(
					"user authentication failed",
					zap.String("user", meta.User()),
					zap.String("fingerprint", fingerprint),
				)

				return nil, fmt.Errorf("unknown public key for %q", meta.User())
			}

			if c.Config.UserCredentialsAllowlist != nil {
				list, err := c.Config.UserCredentialsAllowlist.Get(conn.Request().Context())
				if err != nil {
					return nil, fmt.Errorf("access denied")
				}

				cred := models.CredentialGitHubUser(meta.User())
				if !list.IsAllowed(cred) {
					log(conn.Request()).Info(
						"user rejected",
						zap.String("github_user", meta.User()),
						zap.String("ssh_fingerprint", fingerprint),
					)

					return nil, fmt.Errorf("access denied")
				}
			}

			log(conn.Request()).Info(
				"user authenticated",
				zap.String("github_user", meta.User()),
				zap.String("ssh_fingerprint", fingerprint),
			)

			return &ssh.Permissions{
				Extensions: map[string]string{"pubkey-fp": fingerprint},
			}, nil
		},
	}
	config.AddHostKey(sshHostKey)

	sshConn, _, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		conn.Close()
		return
	}

	select {
	case <-c.Clock.After(time.Second * 5):
	case req := <-reqs:
		if req.Type != "pageship" {
			log(conn.Request()).Debug("invalid req type")
			break
		}

		username := sshConn.User()
		token, err := c.generateUserToken(
			conn.Request().Context(),
			username,
			models.CredentialGitHubUser(username),
			&models.UserCredentialData{
				KeyFingerprint: sshConn.Permissions.Extensions["pubkey-fp"],
			})
		if err != nil {
			log(conn.Request()).Warn("failed to generate token", zap.Error(err))
			req.Reply(false, []byte("internal server error"))
			sshConn.Close()
			return
		}
		req.Reply(true, []byte(token))
	}

	sshConn.Close()
}
