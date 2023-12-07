package api

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
	"golang.org/x/net/websocket"
)

type Client struct {
	endpoint  string
	client    *http.Client
	TokenFunc func(r *http.Request) (string, error)
}

func NewClientWithTransport(endpoint string, transport http.RoundTripper) *Client {
	return &Client{
		endpoint: endpoint,
		client:   &http.Client{Transport: transport},
	}
}

func NewClient(endpoint string) *Client {
	return NewClientWithTransport(endpoint, http.DefaultTransport)
}

func (c *Client) Endpoint() string { return c.endpoint }

func (c *Client) attachToken(r *http.Request) error {
	if c.TokenFunc == nil {
		return nil
	}

	token, err := c.TokenFunc(r)
	if err != nil {
		return err
	}

	r.Header.Add("Authorization", "Bearer "+token)
	return nil
}

func (c *Client) GetManifest(ctx context.Context) (*APIManifest, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "manifest")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIManifest](resp)
}

func (c *Client) CreateApp(ctx context.Context, appID string) (*APIApp, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps")
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, "POST", endpoint, map[string]any{
		"id": appID,
	})
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIApp](resp)
}

func (c *Client) ListApps(ctx context.Context) ([]APIApp, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[[]APIApp](resp)
}

func (c *Client) GetApp(ctx context.Context, id string) (*APIApp, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", id)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIApp](resp)
}

func (c *Client) ListUsers(ctx context.Context, appID string) ([]APIUser, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "users")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[[]APIUser](resp)
}

func (c *Client) AddUser(ctx context.Context, appID string, userID string) error {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "users")
	if err != nil {
		return err
	}

	req, err := newJSONRequest(ctx, "POST", endpoint, map[string]any{
		"userID": userID,
	})
	if err != nil {
		return err
	}
	if err := c.attachToken(req); err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = decodeJSONResponse[struct{}](resp)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteUser(ctx context.Context, appID string, userID string) error {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "users", userID)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	if err := c.attachToken(req); err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = decodeJSONResponse[struct{}](resp)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ConfigureApp(ctx context.Context, appID string, conf *config.AppConfig) (*APIApp, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "config")
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, "PUT", endpoint, map[string]any{
		"config": conf,
	})
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIApp](resp)
}

func (c *Client) CreateSite(ctx context.Context, appID string, siteName string) (*APISite, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "sites")
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, "POST", endpoint, map[string]any{
		"name": siteName,
	})
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APISite](resp)
}

func (c *Client) ListSites(ctx context.Context, appID string) ([]APISite, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "sites")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[[]APISite](resp)
}

func (c *Client) UpdateSite(
	ctx context.Context,
	appID string,
	siteName string,
	patch *SitePatchRequest,
) (*APISite, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "sites", siteName)
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, "PATCH", endpoint, patch)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APISite](resp)
}

func (c *Client) GetDeployment(ctx context.Context, appID string, deploymentName string) (*APIDeployment, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "deployments", deploymentName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIDeployment](resp)
}

func (c *Client) ListDeployments(ctx context.Context, appID string) ([]APIDeployment, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "deployments")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[[]APIDeployment](resp)
}

func (c *Client) SetupDeployment(
	ctx context.Context,
	appID string,
	name string,
	files []models.FileEntry,
	siteConfig *config.SiteConfig,
) (*models.Deployment, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "deployments")
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, "POST", endpoint, map[string]any{
		"name":        name,
		"files":       files,
		"site_config": siteConfig,
	})
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*models.Deployment](resp)
}

func (c *Client) UploadDeploymentTarball(
	ctx context.Context,
	appID string,
	deploymentName string,
	tarball io.Reader,
	size int64,
) (*models.Deployment, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "deployments", deploymentName, "tarball")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, tarball)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}
	req.ContentLength = size

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*models.Deployment](resp)
}

func (c *Client) ListDomains(ctx context.Context, appID string) ([]APIDomain, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "domains")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[[]APIDomain](resp)
}

func (c *Client) ActivateDomain(ctx context.Context, appID string, domainName string, replaceApp string) (*APIDomain, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "domains", domainName, "activate")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if replaceApp != "" {
		req.URL.RawQuery = url.Values{
			"replaceApp": []string{replaceApp},
		}.Encode()
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIDomain](resp)
}

func (c *Client) DeleteDomain(ctx context.Context, appID string, domainName string) (*APIDomain, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "domains", domainName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIDomain](resp)
}

func (c *Client) OpenAuthGitHubSSH(ctx context.Context) (*websocket.Conn, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "auth", "github-ssh")
	if err != nil {
		return nil, err
	}

	endpoint = strings.Replace(endpoint, "http", "ws", 1)
	config, err := websocket.NewConfig(endpoint, "/")
	if err != nil {
		return nil, err
	}
	config.Dialer = &net.Dialer{Cancel: ctx.Done()}

	ws, err := websocket.DialConfig(config)
	return ws, err
}

func (c *Client) AuthGitHubOIDC(ctx context.Context, oidcToken string) (string, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "auth", "github-oidc")
	if err != nil {
		return "", err
	}

	var body struct {
		Token string `json:"token"`
	}
	body.Token = oidcToken
	req, err := newJSONRequest(ctx, "POST", endpoint, body)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[string](resp)
}

func (c *Client) GetMe(ctx context.Context) (*APIUser, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "auth", "me")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := c.attachToken(req); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[*APIUser](resp)
}
