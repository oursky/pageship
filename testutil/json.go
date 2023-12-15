package testutil

import (
	"encoding/json"
	"net/http"

	"github.com/oursky/pageship/internal/api"
)

func DecodeJSONResponse[T any](resp *http.Response) (result T, err error) {
	type response struct {
		Error  *string `json:"error"`
		Result T       `json:"result"`
	}
	if resp.StatusCode != http.StatusOK && (resp.StatusCode < 400 || resp.StatusCode >= 500) {
		err = api.HTTPStatusCodeError{Status: resp.Status, Code: resp.StatusCode}
		return
	}

	var v response
	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		if resp.StatusCode != http.StatusOK {
			err = api.HTTPStatusCodeError{Status: resp.Status, Code: resp.StatusCode}
		}
		return
	}

	if v.Error != nil {
		err = api.ServerError{Message: *v.Error, Code: resp.StatusCode}
		return
	}
	result = v.Result
	return result, nil
}
