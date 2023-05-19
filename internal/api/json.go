package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type HTTPStatusCodeError string

func (e HTTPStatusCodeError) Error() string {
	return string(e)
}

type BadRequestError string

func (e BadRequestError) Error() string {
	return fmt.Sprintf("bad request: %s", string(e))
}

func newJSONRequest(ctx context.Context, method string, endpoint string, v any) (*http.Request, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(v)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func decodeJSONResponse[T any](resp *http.Response) (*T, error) {
	type response struct {
		Error  *string `json:"error"`
		Result *T      `json:"result"`
	}
	if resp.StatusCode != http.StatusOK && (resp.StatusCode < 400 || resp.StatusCode >= 500) {
		return nil, HTTPStatusCodeError(resp.Status)
	}

	var v response
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		if resp.StatusCode == http.StatusBadRequest {
			return nil, HTTPStatusCodeError(resp.Status)
		}
		return nil, err
	}

	if v.Error != nil {
		return nil, errors.New(*v.Error)
	}
	return v.Result, nil
}
