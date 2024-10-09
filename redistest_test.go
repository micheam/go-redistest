package redistest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	cleanup, err := Start(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m.Run()

	// Don't defer cleanup() because os.Exit() doesn't run deferred functions.
	// https://golang.org/pkg/os/#Exit
	if err := cleanup(); err != nil {
		fmt.Println(err)
	}
}

func Test_Client(t *testing.T) {
	ctx := context.Background()
	client, err := Client(ctx)
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

	val := uuid.NewString()
	if err := client.Set(ctx, "key", val, 0).Err(); err != nil {
		t.Fatal(err)
	}
	got, err := client.Get(ctx, "key").Result()
	if err != nil {
		t.Fatal(err)
	}
	if got != val {
		t.Errorf("unexpected value: %s", got)
	}
}
