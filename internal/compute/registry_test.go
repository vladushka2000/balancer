package compute

import (
	"context"
	"testing"
	"time"
)

func TestRegistryInstanceID(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	if r.InstanceID() == "" {
		t.Fatalf("expected non-empty instance id")
	}
}

func TestRegistryAliveCountNoRedis(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	n, err := r.AliveCount(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if n != 1 {
		t.Fatalf("expected alive count 1 with nil redis, got %d", n)
	}
}

func TestRegistryAliveCountCache(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	ctx := context.Background()
	if _, err := r.AliveCount(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.aliveCache != -1 {
		t.Fatalf("expected cache to remain invalid with nil redis, got %d", r.aliveCache)
	}
}

func TestRegistryHeartbeatNoRedis(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	if err := r.heartbeat(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegistryStopNoRedis(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	r.Stop(context.Background())
}

func TestRegistryStartStop(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	ctx := context.Background()
	r.Start(ctx)
	r.Stop(ctx)
}

func TestRegistryStartStopTwice(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	ctx := context.Background()
	r.Start(ctx)
	r.Stop(ctx)
	r.Stop(ctx)
}

func TestRegistryHeartbeatNilRedis(t *testing.T) {
	r := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	if err := r.heartbeat(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewInstanceIDUnique(t *testing.T) {
	a := newInstanceID()
	b := newInstanceID()
	if a == b {
		t.Fatalf("expected distinct instance ids, got %q twice", a)
	}
	if len(a) != 32 {
		t.Fatalf("expected 32 hex chars, got %d", len(a))
	}
}

func TestRegistryAliveCountWithRedis(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-alive"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	r1 := NewRegistry(rdb, ns, 10*time.Second, 3*time.Second)
	r2 := NewRegistry(rdb, ns, 10*time.Second, 3*time.Second)
	r3 := NewRegistry(rdb, ns, 10*time.Second, 3*time.Second)

	if err := r1.heartbeat(ctx); err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	if err := r2.heartbeat(ctx); err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	if err := r3.heartbeat(ctx); err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}

	n, err := r1.AliveCount(ctx)
	if err != nil {
		t.Fatalf("alive count failed: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected 3 alive instances, got %d", n)
	}
}

func TestRegistryAliveCountCached(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-alive-cache"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	r := NewRegistry(rdb, ns, 10*time.Second, 3*time.Second)
	if err := r.heartbeat(ctx); err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	if _, err := r.AliveCount(ctx); err != nil {
		t.Fatalf("alive count failed: %v", err)
	}
	first := r.aliveCache
	if first != 1 {
		t.Fatalf("expected cached 1, got %d", first)
	}
	if _, err := r.AliveCount(ctx); err != nil {
		t.Fatalf("alive count failed: %v", err)
	}
	if r.aliveCache != first {
		t.Fatalf("expected cache to be reused, got %d", r.aliveCache)
	}
}

func TestRegistryStopDeletesKey(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-alive-stop"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	r := NewRegistry(rdb, ns, 10*time.Second, 3*time.Second)
	if err := r.heartbeat(ctx); err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	r.Stop(ctx)
	n, err := r.AliveCount(ctx)
	if err != nil {
		t.Fatalf("alive count failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 alive after stop, got %d", n)
	}
}
