package drift

import (
	"context"
	"fmt"

	"github.com/ha36d/drift-checker/internal/plan"
	"github.com/ha36d/drift-checker/internal/report"
)

type Options struct {
	Path   string // working dir
	Format string // "md" | "text" | "json"
	Strict bool
}

type Result struct {
	Stats          plan.Stats
	RenderedReport string
	DriftDetected  bool
	Runner         plan.RunnerKind
}

// Test hooks (overridable in tests)
var (
	SelectRunner = plan.SelectRunner
	PlanJSON     = plan.MakeRefreshOnlyPlanJSON
)

func CheckDrift(ctx context.Context, opts Options) (Result, error) {
	runner, err := SelectRunner()
	if err != nil {
		return Result{}, err
	}

	planJSON, err := PlanJSON(ctx, runner, opts.Path)
	if err != nil {
		return Result{}, err
	}

	stats, err := plan.ParseStats(planJSON)
	if err != nil {
		return Result{}, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	var rendered string
	switch opts.Format {
	case "", "md", "markdown":
		rendered = report.RenderMarkdown(stats, runner)
	case "text", "txt":
		rendered = report.RenderText(stats, runner)
	case "json":
		rendered, err = report.RenderJSON(stats)
		if err != nil {
			return Result{}, fmt.Errorf("failed to render json: %w", err)
		}
	default:
		return Result{}, fmt.Errorf("unsupported format %q (use md|text|json)", opts.Format)
	}

	return Result{
		Stats:          stats,
		RenderedReport: rendered,
		DriftDetected:  stats.Updates > 0 || stats.Deletes > 0 || stats.Replaces > 0,
		Runner:         runner,
	}, nil
}
