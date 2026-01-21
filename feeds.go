package synthient

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type AnonymizersQuery struct {
	Provider     *string
	Type         *string
	LastObserved *string
	CountryCode  *string
	Format       string
	Full         bool
	Order        string
}

func (client *Client) DownloadAnonymizersFeed(
	query AnonymizersQuery,
	filepath string,
	options *RequestOptions,
) (int64, error) {
	_, err := os.Stat(filepath)
	if !errors.Is(err, fs.ErrNotExist) {
		return 0, fmt.Errorf("creating file at %s: %w", filepath, ErrFileExists)
	}

	params := url.Values{}
	if query.Provider != nil {
		params.Add("provider", *query.Provider)
	} else {

	}
	if query.Type != nil {
		params.Add("type", *query.Type)
	}
	if query.LastObserved != nil {
		params.Add("last_observed", *query.LastObserved)
	}
	if query.CountryCode != nil {
		params.Add("country_code", *query.CountryCode)
	}
	params.Add("full", strconv.FormatBool(query.Full))
	params.Add("format", query.Format)
	params.Add("order", query.Order)

	path, err := url.JoinPath(client.BaseFeeds.String(), "feeds", "anonymizers")
	if err != nil {
		return 0, fmt.Errorf("creating url for anonymizer feed: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request for anonymizer data: %w", err)
	}

	body, err := request(options, client, req, http.StatusOK)
	if err != nil {
		return 0, fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = body.Close() }()

	file, err := os.Create(filepath)
	if err != nil {
		return 0, fmt.Errorf("creating output file (path: %s): %w", filepath, err)
	}
	defer func() { _ = file.Close() }()

	bytes, err := io.Copy(file, body)
	if err != nil {
		return 0, fmt.Errorf("streaming response to file: %w", err)
	}

	err = file.Sync()
	if err != nil {
		return 0, fmt.Errorf("syncing output to file: %w", err)
	}

	return bytes, nil
}
