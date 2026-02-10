package synthient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RequestOptions configures optional per-request behavior for client calls.
//
// It is passed to request helpers and API methods to override defaults without
// changing the Client itself. When Context is non-nil, it is used for request
// cancellation, deadlines, and timeouts. If Context is nil, the request uses
// context.Background() (or the client/request default).
type RequestOptions struct {
	Context context.Context
}

// IMPORTANT: make sure to close the returned reader
func request(
	options *RequestOptions,
	client *Client,
	req *http.Request,
	expectedStatusCode int,
) (io.ReadCloser, error) {
	if options == nil {
		options = &RequestOptions{Context: context.Background()}
	}

	if strings.TrimSpace(client.Token) == "" {
		return nil, ErrNoToken
	}
	req.Header.Add("Authorization", client.Token)
	req = req.WithContext(options.Context)

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request to %s: %w", req.URL.String(), err)
	}

	fail := func(e error) (io.ReadCloser, error) {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("closing file: %w", errors.Join(e, closeErr))
		}
		return nil, e
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return fail(ErrBadRequest)
	case http.StatusUnauthorized:
		return fail(ErrUnauthorized)
	case http.StatusPaymentRequired:
		return fail(ErrPaymentRequired)
	case http.StatusInternalServerError:
		return fail(ErrInternalServerError)
	}
	if resp.StatusCode != expectedStatusCode {
		err = fmt.Errorf(
			`status of %d "%s" (%d "%s" expected) making request: %w`,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			expectedStatusCode,
			http.StatusText(expectedStatusCode),
			ErrUnexpectedStatusCode,
		)
		return fail(err)
	}

	return resp.Body, nil
}

func requestJSON[T any](
	options *RequestOptions,
	client *Client,
	req *http.Request,
	expectedStatusCode int,
) (T, error) {
	var zero T // to be used as "nil"
	body, err := request(options, client, req, expectedStatusCode)
	if err != nil {
		return zero, fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = body.Close() }()

	var data T
	err = json.NewDecoder(body).Decode(&data)
	if err != nil {
		return zero, fmt.Errorf("parsing json: %w", err)
	}

	return data, nil
}
