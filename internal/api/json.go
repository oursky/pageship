package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type HTTPStatusCodeError struct {
	Status string
	Code   int
}

func (e HTTPStatusCodeError) Error() string {
	return e.Status
}
func (e HTTPStatusCodeError) StatusCode() int {
	return e.Code
}

type ServerError struct {
	Message string
	Code    int
}

func (e ServerError) Error() string {
	return e.Message
}
func (e ServerError) StatusCode() int {
	return e.Code
}

func ErrorStatusCode(err error) (int, bool) {
	var e interface{ StatusCode() int }
	if errors.As(err, &e) {
		return e.StatusCode(), true
	}
	return 0, false
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

func decodeJSONResponse[T any](resp *http.Response) (result T, err error) {
	type response struct {
		Error  *string `json:"error"`
		Result T       `json:"result"`
	}
	if resp.StatusCode != http.StatusOK && (resp.StatusCode < 400 || resp.StatusCode >= 500) {
		err = HTTPStatusCodeError{Status: resp.Status, Code: resp.StatusCode}
		return
	}

	var v response
	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		if resp.StatusCode != http.StatusOK {
			err = HTTPStatusCodeError{Status: resp.Status, Code: resp.StatusCode}
		}
		return
	}

	if v.Error != nil {
		err = ServerError{Message: *v.Error, Code: resp.StatusCode}
		return
	}
	result = v.Result
	return result, nil
}
