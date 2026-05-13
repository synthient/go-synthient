package synthient

import (
	"fmt"
	"net/http"
	"net/url"
)

// Account represents the JSON response returned by the Synthient account endpoint.
//
// It groups data into three major sections:
//
//   - Identity: first/last name and email for the authenticated user.
//   - Organization: the organization the user belongs to, including its ID, name, and the user's relation to it.
//   - LookupQuota: remaining lookup credits and seconds until the quota resets.
//
// Fields and nested structs map 1:1 to the API's JSON payload via struct tags.
//
// Commonly used fields include Account.Email, Organization.Name, and LookupQuota.Credits.
type Account struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	Organization struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Relation string `json:"relation"`
	} `json:"organization"`
	Scopes      []string `json:"scopes"`
	LookupQuota struct {
		Credits  int `json:"credits"`
		ResetsIn int `json:"resets_in"`
	} `json:"lookup_quota"`
}

// GetAccount retrieves profile and quota details for the authenticated user.
//
// It performs an HTTP GET request to the Synthient account endpoint and
// unmarshals the JSON response into an Account value. The request is expected to
// return http.StatusOK; non-OK responses are returned as errors.
//
// options can be used to customize request behavior (timeouts, headers, etc.).
//
// Example:
//
//	account, err := client.GetAccount(nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("%+v\n", account)
func (client *Client) GetAccount(options *RequestOptions) (Account, error) {
	path, err := url.JoinPath(client.BaseAPI.String(), "account", "me")
	if err != nil {
		return Account{}, fmt.Errorf("creating path for account request: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return Account{}, fmt.Errorf("making request for account: %w", err)
	}

	resp, err := requestJSON[Account](options, client, req, http.StatusOK)
	if err != nil {
		return Account{}, fmt.Errorf("requesting JSON data: %w", err)
	}

	return resp, nil
}
