package api

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
)

type Client struct {
	endpoint string
	client   *http.Client
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

func (c *Client) SetupDeployment(
	ctx context.Context,
	appID string,
	siteName string,
	files []models.FileEntry,
	siteConfig *config.SiteConfig,
) (*models.Deployment, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "sites", siteName, "deployments")
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, "POST", endpoint, map[string]any{
		"files":       files,
		"site_config": siteConfig,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[models.Deployment](resp)
}

func (c *Client) UploadDeploymentTarball(
	ctx context.Context,
	appID string,
	siteName string,
	deploymentID string,
	tarball io.Reader,
) (*models.Deployment, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "sites", siteName, "deployments", deploymentID, "tarball")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, tarball)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[models.Deployment](resp)
}

func (c *Client) PatchDeployment(
	ctx context.Context,
	appID string,
	siteName string,
	deploymentID string,
	patch *DeploymentPatchRequest,
) (*models.Deployment, error) {
	endpoint, err := url.JoinPath(c.endpoint, "api", "v1", "apps", appID, "sites", siteName, "deployments", deploymentID)
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, "PATCH", endpoint, patch)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeJSONResponse[models.Deployment](resp)
}
