package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ha36d/drift-checker/internal/plan"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	gateInputPath   string
	gateFormat      string
	gateStrict      bool
	gateMaxDeletes  int
	gateMaxReplaces int
	gateList        bool
)

// gateCmd enforces destructive-change policy (delete/replace) on a normal plan JSON.
var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "Enforce destructive-change policy against a plan JSON (not refresh-only)",
	Long: `Reads a standard Terraform/OpenTofu plan JSON (from 'show -json') and enforces a
destructive-change policy. Destructive actions are deletes and replaces.

Exit codes:
  0 = safe
  2 = destructive changes present or thresholds exceeded (when --strict)
  1 = error`,
	Example: `  drift-checker gate --input plan.json --strict
  drift-checker gate --input plan.json --format json --max-deletes 0 --max-replaces 0 --strict
  drift-checker gate --input plan.json --format text --list`,
	RunE: runGate,
}

func init() {
	rootCmd.AddCommand(gateCmd)

	gateCmd.Flags().StringVar(&gateInputPath, "input", "", "path to a normal plan JSON file (from 'show -json') [required]")
	gateCmd.MarkFlagRequired("input")

	gateCmd.Flags().StringVar(&gateFormat, "format", "md", "output format: md|json|text")
	gateCmd.Flags().BoolVar(&gateStrict, "strict", false, "exit with code 2 if destructive changes are present or thresholds exceeded")
	gateCmd.Flags().IntVar(&gateMaxDeletes, "max-deletes", -1, "maximum allowed deletes before failing (negative means unlimited)")
	gateCmd.Flags().IntVar(&gateMaxReplaces, "max-replaces", -1, "maximum allowed replaces before failing (negative means unlimited)")
	gateCmd.Flags().BoolVar(&gateList, "list", false, "include list of destructive resource addresses in the output")
}

type gatePayload struct {
	Updates           int      `json:"updates"`
	Replaces          int      `json:"replaces"`
	Deletes           int      `json:"deletes"`
	DestructiveTotal  int      `json:"destructive_total"`
	Destructive       []string `json:"destructive,omitempty"`
	TotalResourceRefs int      `json:"total"` // equivalent to len(stats.DriftedResources)
}

func runGate(cmd *cobra.Command, args []string) error {
	b, err := os.ReadFile(gateInputPath)
	if err != nil {
		return fmt.Errorf("read --input: %w", err)
	}

	// Reuse the plan parser for counts and total drifted list.
	stats, err := plan.ParseStats(b)
	if err != nil {
		return fmt.Errorf("invalid plan JSON: %w", err)
	}

	// Extract *destructive-only* addresses (delete/replace) for listing/JSON.
	destructiveAddrs, err := extractDestructiveAddresses(b)
	if err != nil {
		return fmt.Errorf("failed to extract destructive addresses: %w", err)
	}

	payload := gatePayload{
		Updates:           stats.Updates,
		Replaces:          stats.Replaces,
		Deletes:           stats.Deletes,
		DestructiveTotal:  stats.Replaces + stats.Deletes,
		Destructive:       nil,
		TotalResourceRefs: len(stats.DriftedResources),
	}
	if gateList {
		payload.Destructive = destructiveAddrs
	}

	// Render per requested format (stdout only). All logs go to stderr.
	switch gateFormat {
	case "", "md", "markdown":
		fmt.Println(renderGateMarkdown(payload))
	case "text", "txt":
		fmt.Println(renderGateText(payload))
	case "json":
		js, err := json.Marshal(payload) // compact, valid JSON
		if err != nil {
			return fmt.Errorf("failed to render json: %w", err)
		}
		fmt.Println(string(js))
	default:
		return fmt.Errorf("unsupported format %q (use md|text|json)", gateFormat)
	}

	// Strict policy gating: destructive present OR thresholds exceeded
	thresholdHit := (gateMaxDeletes >= 0 && stats.Deletes > gateMaxDeletes) ||
		(gateMaxReplaces >= 0 && stats.Replaces > gateMaxReplaces)

	destructivePresent := (stats.Deletes > 0 || stats.Replaces > 0)

	if gateStrict && (destructivePresent || thresholdHit) {
		// Use os.Exit to meet the precise exit-code contract.
		os.Exit(2)
	}

	// Log some context for troubleshooting (stderr).
	log.WithFields(log.Fields{
		"updates":           stats.Updates,
		"replaces":          stats.Replaces,
		"deletes":           stats.Deletes,
		"destructive_total": payload.DestructiveTotal,
		"max_deletes":       gateMaxDeletes,
		"max_replaces":      gateMaxReplaces,
		"strict":            gateStrict,
	}).Debug("gate evaluation")

	return nil
}

func renderGateMarkdown(p gatePayload) string {
	var s string
	s += "## Destructive Change Gate\n\n"
	s += fmt.Sprintf("- **Updates**: %d\n", p.Updates)
	s += fmt.Sprintf("- **Replaces**: %d\n", p.Replaces)
	s += fmt.Sprintf("- **Deletes**: %d\n", p.Deletes)
	s += fmt.Sprintf("- **Destructive total (delete+replace)**: %d\n", p.DestructiveTotal)
	s += fmt.Sprintf("- **Total changed resources in plan**: %d\n", p.TotalResourceRefs)
	if len(p.Destructive) > 0 {
		s += "\n### Destructive Resources\n\n"
		for _, a := range p.Destructive {
			s += fmt.Sprintf("- `%s`\n", a)
		}
	}
	return s
}

func renderGateText(p gatePayload) string {
	var s string
	s += "Destructive Change Gate\n"
	s += fmt.Sprintf("Updates: %d\n", p.Updates)
	s += fmt.Sprintf("Replaces: %d\n", p.Replaces)
	s += fmt.Sprintf("Deletes: %d\n", p.Deletes)
	s += fmt.Sprintf("Destructive total (delete+replace): %d\n", p.DestructiveTotal)
	s += fmt.Sprintf("Total changed resources in plan: %d\n", p.TotalResourceRefs)
	if len(p.Destructive) > 0 {
		s += "\nDestructive Resources:\n"
		for _, a := range p.Destructive {
			s += fmt.Sprintf("- %s\n", a)
		}
	}
	return s
}

// Local minimal structs to extract *destructive* addresses without changing internal/plan API.
type gatePlan struct {
	ResourceChanges []gateRC `json:"resource_changes"`
}
type gateRC struct {
	Address string    `json:"address"`
	Change  gateDelta `json:"change"`
}
type gateDelta struct {
	Actions []string `json:"actions"`
}

func extractDestructiveAddresses(planJSON []byte) ([]string, error) {
	var p gatePlan
	if err := json.Unmarshal(planJSON, &p); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(p.ResourceChanges))
	for _, rc := range p.ResourceChanges {
		acts := rc.Change.Actions
		if isDeleteOnly(acts) || isReplace(acts) {
			out = append(out, rc.Address)
		}
	}
	return out, nil
}

func isDeleteOnly(acts []string) bool {
	return len(acts) == 1 && acts[0] == "delete"
}

func isReplace(acts []string) bool {
	if len(acts) != 2 {
		return false
	}
	// any order of ["create","delete"] is a replace
	first, second := acts[0], acts[1]
	return (first == "create" && second == "delete") || (first == "delete" && second == "create")
}