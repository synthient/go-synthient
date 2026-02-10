# go-synthient

[![godoc](https://pkg.go.dev/badge/github.com/synthient/go-synthient?utm_source=godoc)](https://pkg.go.dev/github.com/synthient/go-synthient)
![go.mod version](https://img.shields.io/github/go-mod/go-version/synthient/go-synthient)
[![report card](https://goreportcard.com/badge/github.com/synthient/go-synthient)](https://goreportcard.com/report/github.com/synthient/go-synthient)

synthient sdk for golang

## Installation

You can add this sdk to your project with the following terminal command:

```bash
go get -u github.com/synthient/go-synthient
```

### Creating a [`synthient.Client`](https://pkg.go.dev/github.com/synthient/go-synthient#Client)

When using this SDK all requests are made with a `synthient.Client`. You can create this struct manually or with the [`synthient.NewClient`](https://pkg.go.dev/github.com/synthient/go-synthient#NewClient) function:

```go
package main

import "github.com/synthient/go-synthient"

func main() {
    client := synthient.NewClient("SECRET TOKEN")
}
```

### Getting IP data

One of the first things you can do with your new client is get IP data. Here is an example of getting IP data for a given IP using [`client.GetIP(...)`](https://pkg.go.dev/github.com/synthient/go-synthient#Client.GetIP):

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	resp, err := client.GetIP("213.149.183.127", nil)
	if err != nil {
		log.Fatalf("failed to get ip address: %s", err)
	}
	fmt.Println(resp.IP)
}
```

`client.GetIP` returns a [`synthient.IP`](https://pkg.go.dev/github.com/synthient/go-synthient#IP) value, along with the error if there is one, of course.

### Anonymizer Feed Data

#### Streaming

You can stream the feed using [`client.StreamAnonymizersFeed(...)`](https://pkg.go.dev/github.com/synthient/go-synthient#Client.StreamAnonymizersFeed):

```go
package main

import (
	"io"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	stream, err := client.StreamAnonymizersFeed(synthient.AnonymizersQuery{
		Provider:     "BIRDPROXIES",
		Type:         "RESIDENTIAL_PROXY",
		LastObserved: "7D",
		Format:       "CSV",
		CountryCode:  "US",
		Full:         false,
		Order:        "desc",
	}, nil)
	if err != nil {
		log.Fatalf("failed to stream feed: %s", err)
	}
	defer func() { _ = stream.Close() }() // important! make sure to close stream

	_, err = io.Copy(os.Stdout, stream)
	if err != nil {
		log.Fatalf("failed to read stream: %s", err)
	}
}
```

#### Downloading

Using [`client.DownloadAnonymizersFeed(...)`](https://pkg.go.dev/github.com/synthient/go-synthient#Client.DownloadAnonymizersFeed) you can easily stream a feed to a file. This will save it to a file called `feed.csv` in this example:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	n, err := client.DownloadAnonymizersFeed(synthient.AnonymizersQuery{
		Provider:     "BIRDPROXIES",
		Type:         "RESIDENTIAL_PROXY",
		LastObserved: "7D",
		Format:       "CSV",
		CountryCode:  "US",
		Full:         false,
		Order:        "desc",
	}, "feed.csv", nil)
	if err != nil {
		log.Fatalf("failed to download feed: %s", err)
	}

	fmt.Println(n, "bytes downloaded")
}
```

## Client Customization

The `synthient.Client` can be customized to use a self-hosted endpoint for example. Here is an example:

```go
package main

import "github.com/synthient/go-synthient"

func main() {
    client := synthient.NewClient("SECRET TOKEN")
    client.BaseAPI.Host = "synthient.myserver.com"
}
```
