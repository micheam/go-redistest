package redistest

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ory/dockertest"
	"github.com/redis/go-redis/v9"
)

const DefaultImageTag = "7.4.1-alpine"

var redisVersion = DefaultImageTag

// SetImageTag sets the image tag for the redis server.
// See https://hub.docker.com/_/redis/tags for available tags.
func SetImageTag(version string) {
	redisVersion = version
}

var (
	once     sync.Once
	resource *dockertest.Resource
	addr     string
	cleanup  func() error

	maxWait = 60 * time.Second
)

// Addr returns the address of the redis server.
// e.g: localhost:6379
func Addr(ctx context.Context) string {
	if addr != "" {
		return addr
	}
	timeouter := time.After(3 * time.Second) // "3秒間だけ待ってやる"
	for {
		select {
		case <-ctx.Done():
			return ""
		case <-timeouter:
			return addr
		}
	}
}

// Client returns a redis client connected to the redis server.
// This will block until the connection is established.
func Client(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     Addr(ctx),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// Wait for the connection to be established.
	bf := backoff.NewExponentialBackOff()
	bf.InitialInterval = 1 * time.Second
	bf.MaxInterval = 10 * time.Second
	bf.MaxElapsedTime = 30 * time.Second
	op := func() error {
		_, err := client.Ping(ctx).Result()
		return err
	}
	err := backoff.Retry(op, backoff.WithContext(bf, ctx))
	if err != nil {
		return nil, fmt.Errorf("[redistest] failed to connect to redis server: %v", err)
	}
	return client, nil
}

// Start starts a redis server using dockertest.
func Start(ctx context.Context) (func() error, error) {
	logger := slog.Default().With("module", "redistest")

	var err error
	var pool *dockertest.Pool
	pool, err = dockertest.NewPool("") // use default on windows (tcp/http) and linux/osx (socket)
	pool.MaxWait = maxWait
	if err != nil {
		return nil, fmt.Errorf("can't connect to docker: %v", err)
	}
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// start minio server and create a bucket
	once.Do(func() {
		resource, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "redis",
			Tag:        redisVersion,
		})
		if err != nil {
			err = fmt.Errorf("can't start redis server: %v", err)
			return
		}

		addr = fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp"))
		logger.DebugContext(ctx, "redis server started", "addr", addr)

		cleanup = func() error {
			if pool != nil {
				return pool.Purge(resource)
			}
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return cleanup, nil
}
