# go-synthient

synthient sdk for golang

## Installation

You can add this sdk to your project with the following terminal command:

```bash
go get -u github.com/synthient/go-synthient
```

## Basic Usage

### Creating a [`synthient.Client`](https://pkg.go.dev/github.com/synthient/go-synthient#Client)

When using this SDK all requests are made with a `synthient.Client`. You can create this struct manually or with the `synthient.NewClient` function:

```go
package main

import "github.com/synthient/go-synthient"

func main() {
    client := synthient.NewClient("SECRET TOKEN")
}
```

### Getting IP data

One of the first things you can do with your new client is get IP data. Here is an example of getting IP data for a given IP:

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

`client.GetIp` returns a [`synthient.IP`](https://pkg.go.dev/github.com/synthient/go-synthient#IP) value, along with the error if there is one, of course.

### Anonymizer Feed Data

You can also get feed data and stream it into a file as seen here:

```go
func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	n, err := client.DownloadAnonymizersFeed(synthient.AnonymizersQuery{}, "feed.csv", nil)
	if err != nil {
		log.Fatalf("failed to download anonymizer feed: %s", err)
	}
	fmt.Printf("%d bytes downloaded\n", n)
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
