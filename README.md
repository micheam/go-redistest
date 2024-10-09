# go-redistest

[![Go](https://github.com/micheam/go-redistest/actions/workflows/go.yml/badge.svg)](https://github.com/micheam/go-redistest/actions/workflows/go.yml)

`go-redistest` is a Go module designed to assist in testing applications that rely on Redis databases. It provides utilities for setting up temporary test Redis instances that are automatically initiated when `go test` is executed and disposed of once the tests are complete. This functionality ensures a clean and isolated testing environment, making it particularly beneficial for developers aiming to streamline and enhance their Redis testing workflows.

## Usage

To use `go-redistest`, integrate it into your Go testing environment. The package offers several functions to help manage and assert Redis states during tests.

### Functions

- `Start(ctx context.Context) (func() error, error)`: Starts a new Redis session.
    This function may be called at the beginning of a test to create a new Redis session.
    It is intended to be called from the `TestMain` function.

- `Client(ctx context.Context) (*redis.Client, error)`: Creates a new Redis client.
    This function can be used within a test to connect to the Redis instance and perform operations.

## Examples

Here's a basic example of how to use `go-redistest` in a test:

```go
package mypackage_test

import (
    "context"
    "fmt"
    "testing"

    "github.com/google/uuid"
    redistest "github.com/micheam/go-redistest"
)

func TestMain(m *testing.M) {
    cleanup, err := redistest.Start(context.Background())
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    m.Run()

    // Don't defer cleanup() because os.Exit() doesn't run deferred functions.
    if err := cleanup(); err != nil {
        fmt.Println(err)
    }
}

func Test_Client(t *testing.T) {
    ctx := context.Background()
    client, err := redistest.Client(ctx)
    if err != nil {
        t.Fatal(err)
    }

    pong, err := client.Ping(ctx).Result()
    if err != nil {
        t.Fatal(err)
    }
    if pong != "PONG" {
        t.Errorf("unexpected pong response: %s", pong)
    }
}
```

For more detailed examples, please refer to the [redistest_test.go](redistest_test.go) file.

## Author

This module was developed by [Michito Maeda](https://micheam.com).
Contributions and feedback are welcome.

## Acknowledgements

`go-redistest` relies heavily on the [ory/dockertest](https://github.com/ory/dockertest) and [redis/go-redis](https://github.com/go-redis/redis).
We would like to express our gratitude to the maintainers of these libraries for their excellent work, which makes this module possible.

