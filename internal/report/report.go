package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ha36d/drift-checker/internal/plan"
)

func RenderMarkdown(s plan.Stats, runner plan.RunnerKind) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Drift Summary (%s)\n\n", runner)
	fmt.Fprintf(&b, "- **Updates**: %d\n", s.Updates)
	fmt.Fprintf(&b, "- **Replaces**: %d\n", s.Replaces)
	fmt.Fprintf(&b, "- **Deletes**: %d\n", s.Deletes)
	fmt.Fprintf(&b, "- **Total changed resources**: %d\n\n", len(s.DriftedResources))

	if len(s.DriftedResources) > 0 {
		b.WriteString("### Drifted Resources\n\n")
		for _, addr := range s.DriftedResources {
			fmt.Fprintf(&b, "- `%s`\n", addr)
		}
		b.WriteString("\n")
	} else {
		b.WriteString("_No drift detected._\n")
	}
	return b.String()
}

func RenderText(s plan.Stats, runner plan.RunnerKind) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Drift Summary (%s)\n", runner)
	fmt.Fprintf(&b, "Updates: %d\n", s.Updates)
	fmt.Fprintf(&b, "Replaces: %d\n", s.Replaces)
	fmt.Fprintf(&b, "Deletes: %d\n", s.Deletes)
	fmt.Fprintf(&b, "Total changed resources: %d\n", len(s.DriftedResources))

	if len(s.DriftedResources) > 0 {
		b.WriteString("\nDrifted Resources:\n")
		for _, addr := range s.DriftedResources {
			fmt.Fprintf(&b, "- %s\n", addr)
		}
	} else {
		b.WriteString("\nNo drift detected.\n")
	}
	return b.String()
}

// RenderJSON outputs the minimal, machine-readable summary.
// Fields mirror Markdown/Text modes (total == number of drifted resources).
// No extra data is included.
func RenderJSON(s plan.Stats) (string, error) {
	payload := struct {
		Updates  int      `json:"updates"`
		Replaces int      `json:"replaces"`
		Deletes  int      `json:"deletes"`
		Drifted  []string `json:"drifted"`
		Total    int      `json:"total"`
	}{
		Updates:  s.Updates,
		Replaces: s.Replaces,
		Deletes:  s.Deletes,
		Drifted:  s.DriftedResources,
		Total:    len(s.DriftedResources),
	}

	b, err := json.Marshal(payload) // compact valid JSON (no extra whitespace)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
