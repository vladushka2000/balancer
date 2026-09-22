package compute

import "testing"

func TestCorrKey(t *testing.T) {
	if got := corrKey("pii", "id1"); got != "pii:corr:id1" {
		t.Fatalf("expected pii:corr:id1, got %q", got)
	}
}

func TestSystemKey(t *testing.T) {
	if got := systemKey("pii", "s1"); got != "pii:system:s1" {
		t.Fatalf("expected pii:system:s1, got %q", got)
	}
}

func TestSystemsSet(t *testing.T) {
	if got := systemsSet("pii"); got != "pii:systems" {
		t.Fatalf("expected pii:systems, got %q", got)
	}
}

func TestConfigEpochKey(t *testing.T) {
	if got := configEpochKey("pii"); got != "pii:control:config_epoch" {
		t.Fatalf("expected pii:control:config_epoch, got %q", got)
	}
}

func TestAliveKey(t *testing.T) {
	if got := aliveKey("pii", "inst1"); got != "pii:alive:inst1" {
		t.Fatalf("expected pii:alive:inst1, got %q", got)
	}
}

func TestStatsKey(t *testing.T) {
	if got := statsKey("pii", "inst1"); got != "pii:stats:inst1" {
		t.Fatalf("expected pii:stats:inst1, got %q", got)
	}
}

func TestAliveSet(t *testing.T) {
	if got := aliveSet("pii"); got != "pii:alive" {
		t.Fatalf("expected pii:alive, got %q", got)
	}
}

func TestStatsSet(t *testing.T) {
	if got := statsSet("pii"); got != "pii:stats" {
		t.Fatalf("expected pii:stats, got %q", got)
	}
}
