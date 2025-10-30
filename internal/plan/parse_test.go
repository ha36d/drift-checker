package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return b
}

func TestParseStats_Clean(t *testing.T) {
	b := mustRead(t, filepath.Join("testdata", "plan_clean.json"))

	s, err := ParseStats(b)
	if err != nil {
		t.Fatalf("ParseStats error: %v", err)
	}

	if s.Updates != 0 || s.Replaces != 0 || s.Deletes != 0 {
		t.Fatalf("expected all zero counts, got %+v", s)
	}
	if DriftCount(s) != 0 {
		t.Fatalf("DriftCount expected 0, got %d", DriftCount(s))
	}
	if len(s.DriftedResources) != 0 {
		t.Fatalf("expected no drifted resources, got %v", s.DriftedResources)
	}
	if s.TotalResources != 0 {
		t.Fatalf("expected TotalResources 0, got %d", s.TotalResources)
	}
}

func TestParseStats_Drift(t *testing.T) {
	b := mustRead(t, filepath.Join("testdata", "plan_drift.json"))

	s, err := ParseStats(b)
	if err != nil {
		t.Fatalf("ParseStats error: %v", err)
	}

	if s.Updates != 1 {
		t.Fatalf("Updates expected 1, got %d", s.Updates)
	}
	if s.Replaces != 1 {
		t.Fatalf("Replaces expected 1, got %d", s.Replaces)
	}
	if s.Deletes != 1 {
		t.Fatalf("Deletes expected 1, got %d", s.Deletes)
	}
	if DriftCount(s) != 3 {
		t.Fatalf("DriftCount expected 3, got %d", DriftCount(s))
	}
	if got := len(s.DriftedResources); got != 3 {
		t.Fatalf("expected 3 drifted resources, got %d (%v)", got, s.DriftedResources)
	}
	if s.TotalResources != 3 {
		t.Fatalf("expected TotalResources 3, got %d", s.TotalResources)
	}
}