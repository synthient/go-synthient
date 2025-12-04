package synthient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RequestOptions struct {
	Context context.Context
}

func request[T any](
	options *RequestOptions,
	client *Client,
	req *http.Request,
	expectedStatusCode int,
) (T, error) {
	var zeroValue T // to be used as "nil"
	if options == nil {
		options = &RequestOptions{Context: context.Background()}
	}

	if strings.TrimSpace(client.Token) == "" {
		return zeroValue, ErrNoToken
	}
	req.Header.Add("Authorization", client.Token)
	req = req.WithContext(options.Context)

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return zeroValue, fmt.Errorf("%w failed to perform request to %s", err, req.URL.String())
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		err = ErrBadRequest
	case http.StatusUnauthorized:
		err = ErrUnauthorized
	case http.StatusPaymentRequired:
		err = ErrPaymentRequired
	case http.StatusInternalServerError:
		err = ErrInternalServerError
	}
	if err != nil {
		return zeroValue, err
	}
	if resp.StatusCode != expectedStatusCode {
		return zeroValue, fmt.Errorf(
			"%w: expected status code of %d but got a status code of %d",
			ErrUnexpectedStatusCode,
			expectedStatusCode,
			resp.StatusCode,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroValue, fmt.Errorf("%w reading response body failed", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return zeroValue, fmt.Errorf("%w failed to close response body", err)
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		return zeroValue, fmt.Errorf("%w parsing json failed", err)
	}

	return data, nil
}
