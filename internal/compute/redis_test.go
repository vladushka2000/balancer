package compute

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const testRedisAddr = "localhost:16379"

// testRedis returns a redis client for integration tests, skipping if unavailable.
func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	if os.Getenv("PII_TEST_REDIS") == "0" {
		t.Skip("redis integration tests disabled")
	}
	rdb := redis.NewClient(&redis.Options{Addr: testRedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unavailable at %s: %v", testRedisAddr, err)
	}
	return rdb
}
