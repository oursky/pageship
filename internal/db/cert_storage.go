package db

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/caddyserver/certmagic"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/time"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/sha3"
)

var dataEncoding = base64.RawURLEncoding

const (
	prefixPlaintext = "1"
	prefixEncrypted = "1e"
)

type CertStorage struct {
	clock  time.Clock
	mu     sync.Mutex
	locks  map[string]struct{}
	DB     DB
	key    *[32]byte
	locker LockerDB
}

func NewCertStorage(db DB, pass string) *CertStorage {
	var key *[32]byte
	if pass != "" {
		h := sha3.New256()
		h.Write([]byte(pass))
		key = new([32]byte)
		copy(key[:], h.Sum(nil))
	}

	return &CertStorage{
		clock: time.SystemClock,
		locks: make(map[string]struct{}),
		DB:    db,
		key:   key,
	}
}

func (s *CertStorage) Lock(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, locked := s.locks[name]; locked {
		return models.ErrCertificateDataLocked
	}

	if s.locker == nil {
		locker, err := s.DB.Locker(ctx)
		if err != nil {
			return err
		}
		s.locker = locker
	}

	err := s.locker.Lock(ctx, name)
	if err != nil {
		return err
	}

	s.locks[name] = struct{}{}
	return nil
}

func (s *CertStorage) Unlock(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, locked := s.locks[name]; !locked {
		return nil
	}

	if s.locker == nil {
		return nil
	}

	if err := s.locker.Unlock(ctx, name); err != nil {
		return err
	}

	delete(s.locks, name)
	if len(s.locks) == 0 {
		s.locker.Close()
		s.locker = nil
	}

	return nil
}

func (s *CertStorage) encrypt(d []byte) (string, error) {
	prefix := prefixPlaintext
	data := d
	if s.key != nil {
		var nonce [24]byte
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			return "", err
		}

		data = secretbox.Seal(nonce[:], d, &nonce, s.key)
		prefix = prefixEncrypted
	}
	return prefix + ":" + dataEncoding.EncodeToString(data), nil
}

func (s *CertStorage) decrypt(d string) ([]byte, error) {
	prefix, encoded, ok := strings.Cut(d, ":")
	if !ok {
		return nil, fs.ErrNotExist
	}

	data, err := dataEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fs.ErrNotExist
	}

	switch prefix {
	case prefixPlaintext:
		return data, nil

	case prefixEncrypted:
		if s.key == nil {
			return nil, fs.ErrNotExist
		}

		var nonce [24]byte
		copy(nonce[:], data[:24])
		decrypted, ok := secretbox.Open(nil, data[24:], &nonce, s.key)
		if !ok {
			return nil, fs.ErrNotExist
		}
		data = decrypted

	default:
		return nil, fs.ErrNotExist
	}
	return data, nil
}

func (s *CertStorage) Load(ctx context.Context, key string) ([]byte, error) {
	entry, err := s.DB.GetCertDataEntry(ctx, key)
	if errors.Is(err, models.ErrCertificateDataNotFound) {
		return nil, fs.ErrNotExist
	} else if err != nil {
		return nil, err
	}
	return s.decrypt(entry.Value)
}

func (s *CertStorage) Store(ctx context.Context, key string, value []byte) error {
	data, err := s.encrypt(value)
	if err != nil {
		return err
	}

	entry := models.NewCertDataEntry(key, data, s.clock.Now())
	return s.DB.SetCertDataEntry(ctx, entry)
}

func (s *CertStorage) Delete(ctx context.Context, key string) error {
	err := s.DB.DeleteCertificateData(ctx, key)
	if errors.Is(err, models.ErrCertificateDataNotFound) {
		return fs.ErrNotExist
	} else if err != nil {
		return err
	}
	return nil
}

func (s *CertStorage) Exists(ctx context.Context, key string) bool {
	entry, err := s.DB.GetCertDataEntry(ctx, key)
	if err != nil {
		return false
	}
	_, err = s.decrypt(entry.Value)
	return err == nil
}

func (s *CertStorage) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	if recursive {
		return nil, errors.New("recursive is not supported")
	}
	keys, err := s.DB.ListCertificateData(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// Remove nested entry
	n := 0
	for _, key := range keys {
		k := strings.TrimLeft(strings.TrimPrefix(key, prefix), "/")
		if !strings.Contains(k, "/") {
			keys[n] = k
			n++
		}
	}
	keys = keys[:n]

	return keys, nil
}

func (s *CertStorage) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	keys, err := s.DB.ListCertificateData(ctx, key)
	if err != nil {
		return certmagic.KeyInfo{}, err
	}

	dir := key + "/"
	isDir := false
	for _, k := range keys {
		if k == key {
			entry, err := s.DB.GetCertDataEntry(ctx, key)
			if err != nil {
				return certmagic.KeyInfo{}, err
			}
			return certmagic.KeyInfo{
				Key:        key,
				Modified:   entry.UpdatedAt,
				IsTerminal: true,
			}, nil
		} else if strings.HasPrefix(k, dir) {
			isDir = true
		}
	}

	if isDir {
		return certmagic.KeyInfo{
			Key:        key,
			IsTerminal: false,
		}, nil
	} else {
		return certmagic.KeyInfo{}, fs.ErrNotExist
	}
}

var _ certmagic.Storage = &CertStorage{}
