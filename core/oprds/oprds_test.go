package oprds

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v9"
)

func TestCheckUnSyncNodeID(t *testing.T) {
	rCli := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	cli := &Client{rds: rCli}
	id := "e_854a5256-4326-40ef-8fb7-6af7e3c5f1d1"

	next, err := cli.CheckUnSyncNodeID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(next)
}
