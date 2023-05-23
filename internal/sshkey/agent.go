package sshkey

import (
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Agent struct {
	conn   net.Conn
	client agent.ExtendedAgent
}

func NewAgent() (*Agent, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}

	client := agent.NewClient(conn)

	return &Agent{conn: conn, client: client}, nil
}

func (a *Agent) Close() error {
	return a.conn.Close()
}

func (a *Agent) Signers() ([]ssh.Signer, error) {
	return a.client.Signers()
}
