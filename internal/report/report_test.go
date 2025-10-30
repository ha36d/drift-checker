package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ha36d/drift-checker/internal/plan"
)

func TestRenderMarkdownAndJSON(t *testing.T) {
	s := plan.Stats{
		Updates:          1,
		Replaces:         1,
		Deletes:          1,
		DriftedResources: []string{"a", "b", "c"},
		TotalResources:   3,
	}

	md := RenderMarkdown(s, plan.RunnerTofu)
	if !strings.Contains(md, "Drift Summary (tofu)") {
		t.Fatalf("markdown missing runner header: %s", md)
	}
	for _, want := range []string{"**Updates**: 1", "**Replaces**: 1", "**Deletes**: 1", "Total changed resources**: 3"} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q in:\n%s", want, md)
		}
	}

	js, err := RenderJSON(s)
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}
	var payload struct {
		Updates  int      `json:"updates"`
		Replaces int      `json:"replaces"`
		Deletes  int      `json:"deletes"`
		Drifted  []string `json:"drifted"`
		Total    int      `json:"total"`
	}
	if err := json.Unmarshal([]byte(js), &payload); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if payload.Updates != 1 || payload.Replaces != 1 || payload.Deletes != 1 || payload.Total != 3 {
		t.Fatalf("unexpected json payload: %+v", payload)
	}
	if len(payload.Drifted) != 3 {
		t.Fatalf("expected 3 drifted, got %d", len(payload.Drifted))
	}
}