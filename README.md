# drift-checker

A CLI tool for drift detection and destructive-change gating in Terraform/OpenTofu plans.

## Features
- Detects drift using **refresh-only** plans (OpenTofu preferred, Terraform fallback)
- Parses plan JSON and counts **updates / deletes / replaces**
- Outputs **Markdown** (default), **text**, or **json** summary (counts + drifted resources)
- Flags: `--path` (default `.`), `--format md|text|json` (default `md`), `--strict` (exit code 2 if drift detected)
- Works with only Terraform **or** only OpenTofu installed
- CI-ready with strict exit codes

## Quickstart
```bash
# In your IaC directory
tofu init || terraform init

# Run scan (Markdown output by default)
drift-checker scan --path . --strict

# Plain text output
drift-checker scan --path . --format text

# Machine-readable JSON output (stdout is **only** the JSON)
drift-checker scan --path . --format json
````

### JSON schema (output on stdout)

```json
{
  "updates":  0,
  "replaces": 0,
  "deletes":  0,
  "drifted":  ["module.vpc.aws_subnet.public[0]"],
  "total":    1
}
```

### What counts as drift?

We look at `resource_changes[*].change.actions` from the plan JSON (`show -json`):

* `["update"]` → **update**
* `["delete"]` → **delete**
* `["create","delete"]` (any order) → **replace**

Any of the above increments drift counts and lists the resource address.

### Exit codes

* `0` – No drift detected
* `2` – Drift detected **and** `--strict` used
* `1` – Other error

## Implementation Notes

* Prefers `tofu`, falls back to `terraform`
* Uses reliable two-step JSON flow:

  1. `<runner> plan -refresh-only -out=drift-checker.plan`
  2. `<runner> show -json drift-checker.plan`
* No dependency on streaming `plan -json` support.
* All logs are written to **stderr**; **stdout** is reserved for the report (Markdown/Text/JSON).

## Project Structure

* `cmd/` – CLI commands (Cobra)
* `internal/plan/` – Runner selection, plan execution, and JSON parsing
* `internal/report/` – Output formatting
* `internal/drift/` – Orchestration

## License

Apache-2.0

---

### Notes / Next steps

1. Add unit tests with `testdata/*.json` plans.
2. Optionally support streaming `plan -json` when available.
3. Consider per-module/type breakdowns.

```

**Notes on behavior**
- `--format json` prints **only** the JSON object to stdout.
- All logs (info/warn/error) are explicitly sent to **stderr** via `log.SetOutput(os.Stderr)` in `cmd/root.go`.
- JSON fields: `updates`, `replaces`, `deletes`, `drifted` (list of resource addresses), `total` (length of `drifted`). Matches the semantics of Markdown/Text modes’ “Total changed resources”.
```
