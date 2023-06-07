package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func canAuthGitHubOIDC() bool {
	_, ok1 := os.LookupEnv("ACTIONS_ID_TOKEN_REQUEST_URL")
	_, ok2 := os.LookupEnv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	return ok1 && ok2
}

func authGitHubOIDC(ctx context.Context) (string, error) {
	tokenURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	runtimeToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")

	u, err := url.Parse(tokenURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("audience", API().Endpoint())
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json; api-version=2.0")
	req.Header.Set("Authorization", "Bearer "+runtimeToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	var respBody struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}
	if respBody.Value == "" {
		return "", fmt.Errorf("GitHub returned no token")
	}

	return API().AuthGitHubOIDC(ctx, respBody.Value)
}
