package sshkey

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oursky/pageship/internal/cache"
	"golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"
)

type GitHubKeys struct {
	ctx    context.Context
	l      *rate.Limiter
	client *http.Client
	cache  *cache.Cache[map[string]struct{}]
}

func NewGitHubKeys(ctx context.Context) (*GitHubKeys, error) {
	keys := &GitHubKeys{
		ctx:    ctx,
		l:      rate.NewLimiter(rate.Every(time.Minute), 10),
		client: &http.Client{},
	}

	cache, err := cache.NewCache(100, time.Minute, keys.doLoad)
	if err != nil {
		return nil, err
	}
	keys.cache = cache

	return keys, nil
}

func (g *GitHubKeys) PublicKey(username string) (map[string]struct{}, error) {
	pkeys, err := g.cache.Load(strings.ToLower(username))
	if err != nil {
		return nil, err
	}
	return pkeys, nil
}

func (g *GitHubKeys) doLoad(username string) (map[string]struct{}, error) {
	ctx, cancel := context.WithTimeout(g.ctx, time.Second*10)
	defer cancel()

	g.l.Wait(ctx)

	u := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/" + url.PathEscape(username) + ".keys",
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]struct{})
	for {
		pkey, _, _, rest, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			// Assume EOF
			break
		}

		keys[string(pkey.Marshal())] = struct{}{}
		data = rest
	}

	return keys, nil
}
