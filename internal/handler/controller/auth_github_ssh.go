package controller

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/sshkey"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

var sshHostKey ssh.Signer
var githubKeys *sshkey.GitHubKeys

func init() {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	sshHostKey, err = ssh.NewSignerFromSigner(key)
	if err != nil {
		panic(err)
	}

	githubKeys, err = sshkey.NewGitHubKeys()
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
			pubKeys, err := githubKeys.PublicKey(meta.User())
			if err != nil {
				c.Logger.Warn("cannot get GitHub public key",
					zap.String("user", meta.User()),
					zap.Error(err),
				)
				return nil, err
			}

			if _, ok := pubKeys[string(pubKey.Marshal())]; !ok {
				c.Logger.Debug(
					"user authentication failed",
					zap.String("user", meta.User()),
					zap.String("fingerprint", fingerprint),
				)

				return nil, fmt.Errorf("unknown public key for %q", meta.User())
			}

			c.Logger.Info(
				"user authenticated",
				zap.String("user", meta.User()),
				zap.String("fingerprint", fingerprint),
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
			c.Logger.Debug("invalid req type")
			break
		}

		token, err := c.generateUserToken(
			conn.Request().Context(),
			models.UserCredentialGitHubUser(sshConn.User()),
			&models.UserCredentialData{
				KeyFingerprint: sshConn.Permissions.Extensions["pubkey-fp"],
			})
		if err != nil {
			c.Logger.Warn("failed to generate token", zap.Error(err))
			sshConn.Close()
			return
		}
		req.Reply(true, []byte(token))
	}

	sshConn.Close()
}
