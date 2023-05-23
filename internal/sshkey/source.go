package sshkey

import (
	"errors"

	"golang.org/x/crypto/ssh"
)

type Source interface {
	Close() error
	Signers() ([]ssh.Signer, error)
}

func AuthMethodSources(sources []Source) ssh.AuthMethod {
	i := 0
	return ssh.RetryableAuthMethod(ssh.PublicKeysCallback(func() (signers []ssh.Signer, err error) {
		if i >= len(sources) {
			return nil, errors.New("no more auth sources")
		}
		signers, err = sources[i].Signers()
		i++
		return signers, err
	}), len(sources))
}
