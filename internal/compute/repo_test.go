package compute

import (
	"context"
	"testing"
	"time"

	"pii/internal/models"
)

func TestRepoPublishReadStats(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-repo"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	repo := NewRepo(rdb, ns)
	snap := Snapshot{RequestsTotal: 100, MaskOK: 60, DemaskOK: 40}
	if err := repo.PublishStats(ctx, "inst1", snap, 10*time.Second); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if err := repo.PublishStats(ctx, "inst2", Snapshot{RequestsTotal: 50, MaskOK: 30, DemaskOK: 20}, 10*time.Second); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	snaps, err := repo.ReadAllStats(ctx)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	agg := Aggregate(snaps)
	if agg.RequestsTotal != 150 {
		t.Fatalf("expected requests_total 150, got %d", agg.RequestsTotal)
	}
	if agg.MaskOK != 90 || agg.DemaskOK != 60 {
		t.Fatalf("expected mask_ok=90 demask_ok=60, got %d/%d", agg.MaskOK, agg.DemaskOK)
	}
}

func TestRepoReadAllStatsEmpty(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-repo-empty"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	repo := NewRepo(rdb, ns)
	snaps, err := repo.ReadAllStats(ctx)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(snaps) != 0 {
		t.Fatalf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestRepoSaveGetSystem(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-sys"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	repo := NewRepo(rdb, ns)
	cfg := models.SystemConfig{SystemID: "sys1", Enabled: true}
	if err := repo.SaveSystem(ctx, cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	got, err := repo.GetSystem(ctx, "sys1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got == nil || got.SystemID != "sys1" {
		t.Fatalf("expected sys1, got %+v", got)
	}
}

func TestRepoGetSystemMissing(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-sys-missing"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	repo := NewRepo(rdb, ns)
	got, err := repo.GetSystem(ctx, "nope")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestRepoListSystems(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-list"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	repo := NewRepo(rdb, ns)
	if err := repo.SaveSystem(ctx, models.SystemConfig{SystemID: "a"}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := repo.SaveSystem(ctx, models.SystemConfig{SystemID: "b"}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	systems, err := repo.ListSystems(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(systems) != 2 {
		t.Fatalf("expected 2 systems, got %d", len(systems))
	}
}

func TestRepoBumpGetControlEpoch(t *testing.T) {
	rdb := testRedis(t)
	ctx := context.Background()
	ns := "pii-test-epoch"
	rdb.FlushDB(ctx)
	defer rdb.FlushDB(ctx)

	repo := NewRepo(rdb, ns)
	if err := repo.BumpConfigEpoch(ctx); err != nil {
		t.Fatalf("bump failed: %v", err)
	}
	if err := repo.BumpConfigEpoch(ctx); err != nil {
		t.Fatalf("bump failed: %v", err)
	}
	epoch, err := repo.GetControlEpoch(ctx)
	if err != nil {
		t.Fatalf("get epoch failed: %v", err)
	}
	if epoch != 2 {
		t.Fatalf("expected epoch 2, got %d", epoch)
	}
}
