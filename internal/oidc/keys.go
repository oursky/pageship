package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/oursky/pageship/internal/cache"
	"golang.org/x/time/rate"
)

type Key struct {
	Issuer string
	JWKS   *keyfunc.JWKS
}

type Keys struct {
	ctx    context.Context
	l      *rate.Limiter
	client *http.Client
	cache  *cache.Cache[*Key]
}

func NewKeys(ctx context.Context) (*Keys, error) {
	keys := &Keys{
		ctx:    ctx,
		l:      rate.NewLimiter(rate.Every(time.Minute), 10),
		client: &http.Client{},
	}

	cache, err := cache.NewCache(100, time.Hour, keys.load)
	if err != nil {
		return nil, err
	}
	keys.cache = cache

	return keys, nil
}

func (k *Keys) Get(issuer string) (*Key, error) {
	return k.cache.Load(issuer)
}

func (k *Keys) load(issuer string) (*Key, error) {
	ctx, cancel := context.WithTimeout(k.ctx, time.Second*10)
	defer cancel()

	u, err := url.Parse(issuer)
	if err != nil {
		return nil, err
	}
	u = u.JoinPath(".well-known", "openid-configuration")

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get openid config: %w", err)
	}
	defer resp.Body.Close()

	var conf struct {
		Issuer  string `json:"issuer"`
		JWKSUri string `json:"jwks_uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&conf); err != nil {
		return nil, fmt.Errorf("parse openid config: %w", err)
	}

	jwks, err := keyfunc.Get(conf.JWKSUri, keyfunc.Options{
		Ctx:            ctx,
		Client:         k.client,
		RefreshTimeout: time.Second * 10,
	})
	if err != nil {
		return nil, fmt.Errorf("get jwks: %w", err)
	}

	return &Key{Issuer: conf.Issuer, JWKS: jwks}, nil
}
