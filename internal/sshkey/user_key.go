package sshkey

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/oursky/pageship/internal/config"
	"golang.org/x/crypto/ssh"
)

type UserKey struct {
	promptKeyfile    func(def string) (string, error)
	promptPassphrase func() ([]byte, error)
}

func NewUserKey(
	promptKeyfile func(def string) (string, error),
	promptPassphrase func() ([]byte, error),
) *UserKey {
	return &UserKey{
		promptKeyfile,
		promptPassphrase,
	}
}

func (k *UserKey) Close() error { return nil }

func (k *UserKey) getKeyfile() (string, error) {
	conf, err := config.LoadClientConfig()
	if err != nil {
		return "", err
	}

	if conf.SSHKeyFile != "" {
		_, err := os.Stat(conf.SSHKeyFile)
		if err == nil {
			return conf.SSHKeyFile, nil
		}
	}

	homeKeyFile, err := homeSSHKeyFile()
	if err != nil {
		return "", err
	}

	keyFile, err := k.promptKeyfile(homeKeyFile)
	if err != nil {
		return "", err
	}

	conf.SSHKeyFile = keyFile
	_ = conf.Save()

	return keyFile, nil
}

func (k *UserKey) Signers() ([]ssh.Signer, error) {
	file, err := k.getKeyfile()
	if err != nil {
		return nil, err
	}

	pem, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(pem)
	if errors.As(err, new(*ssh.PassphraseMissingError)) {
		passphrase, perr := k.promptPassphrase()
		if perr != nil {
			return nil, perr
		}
		key, err = ssh.ParsePrivateKeyWithPassphrase(pem, passphrase)
	}
	if err != nil {
		return nil, err
	}

	return []ssh.Signer{key}, nil
}

func homeSSHKeyFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	ssh := filepath.Join(home, ".ssh")
	matches, err := filepath.Glob(filepath.Join(ssh, "id_*"))
	if err != nil {
		return "", err

	}

	keyFile := ""
	for _, path := range matches {
		name := filepath.Base(path)
		if name == "id_rsa" || name == "id_ed25519" {
			keyFile = path
			break
		}
	}

	return keyFile, nil
}
