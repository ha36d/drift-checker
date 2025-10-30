package plan

import (
	"encoding/json"
	"fmt"
	"slices"
)

type tfPlan struct {
	ResourceChanges []resourceChange `json:"resource_changes"`
}

type resourceChange struct {
	Address string       `json:"address"`
	Mode    string       `json:"mode"`
	Type    string       `json:"type"`
	Name    string       `json:"name"`
	Change  changeDetail `json:"change"`
}

type changeDetail struct {
	Actions []string `json:"actions"`
}

type Stats struct {
	Updates          int
	Deletes          int
	Replaces         int
	DriftedResources []string
	TotalResources   int
}

// ParseStats extracts counts from Terraform/OpenTofu plan JSON (from `show -json`).
func ParseStats(planJSON []byte) (Stats, error) {
	var p tfPlan
	if err := json.Unmarshal(planJSON, &p); err != nil {
		return Stats{}, fmt.Errorf("invalid plan JSON: %w", err)
	}

	stats := Stats{
		TotalResources: len(p.ResourceChanges),
	}

	for _, rc := range p.ResourceChanges {
		acts := rc.Change.Actions
		if len(acts) == 0 {
			continue
		}

		// Recognize actions:
		// - ["update"]
		// - ["delete"]
		// - ["create","delete"] or ["delete","create"] => replace
		// (refresh-only plans surface drift via update/delete/replace)
		isReplace := (slices.Contains(acts, "create") && slices.Contains(acts, "delete"))
		isUpdate := len(acts) == 1 && acts[0] == "update"
		isDelete := len(acts) == 1 && acts[0] == "delete"

		switch {
		case isReplace:
			stats.Replaces++
			stats.DriftedResources = append(stats.DriftedResources, rc.Address)
		case isUpdate:
			stats.Updates++
			stats.DriftedResources = append(stats.DriftedResources, rc.Address)
		case isDelete:
			stats.Deletes++
			stats.DriftedResources = append(stats.DriftedResources, rc.Address)
		}
	}

	return stats, nil
}