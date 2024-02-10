# Go MongoDB Client

The MongoDB supported client for Go.

## Requirements

- Go 1.20 or higher
- MongoDB 3.6 and higher

## Installation
The recommended way to get started using the Go MongoDB Client is by using Go modules to install the dependency in 
your project. This can be done either by importing packages from `github.com/phoenixTW/go-mongodb-client` and having 
the build step install the dependency or by explicitly running

```bash
go get github.com/phoenixTW/go-mongodb-client
```

## Usage

### New Client

To get started with the client, import the `mongo` package and create a `mongodb.New`:

```go
package example

import (
	"context"
	"github.com/phoenixTW/go-mongodb-client/mongodb"
	"time"
)

func example() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := mongodb.New(ctx, "mongodb://localhost:27017", "example", nil)
}
```

Make sure to defer a call to `Disconnect` after instantiating your client:

```go
defer func() {
    if err = client.Disconnect(ctx); err != nil {
        panic(err)
    }
}()
```

For more advanced configuration and authentication, see the [documentation for mongo.Connect](https://pkg.go.dev/go.
mongodb.org/mongo-driver/mongo#Connect).

### Mongo Storage

To initiate storage for mongodb, import the `mongostorage` package and create a `mongostorage.New`:

```go
	storage := mongostorage.New(client.Database("example-database"))
```

### Retry Storage

To initiate retry storage for mongodb, import the `mongostorage` package and create a `mongostorage.NewRetry`:

```go
    storage := mongostorage.NewRetry(client.Database("example-database"), 3, 3*time.Second)
```

# License

The MongoDB Go Driver is licensed under the [Apache License](./LICENSE).
