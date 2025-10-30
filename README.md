# drift-checker

A CLI tool for drift detection and destructive-change gating in Terraform/OpenTofu plans.

## Features
- Detects drift using **refresh-only** plans (OpenTofu preferred, Terraform fallback)
- Parses plan JSON and counts **updates / deletes / replaces**
- Outputs **Markdown** (default), **text**, or **json** summary (counts + resource addresses)
- Flags: `--path` (default `.`), `--format md|text|json` (default `md`), `--strict` (exit code 2 if drift detected)
- **Gate** subcommand to enforce **destructive-change policy** (delete/replace) for normal plan JSON
- Works with only Terraform **or** only OpenTofu installed
- CI-ready with strict exit codes

---

## Quickstart (Drift Detection)

> Copy-paste friendly. Replace `.` with your IaC directory if needed.

### 1) Install / Build
```bash
# With Go installed (Go 1.22+):
go install github.com/ha36d/drift-checker@latest

# Or build from the repo root:
go build -o drift-checker .
````

### 2) Prepare your working directory

```bash
# In your IaC directory
tofu init || terraform init
```

### 3) Create a refresh-only plan (optional but useful for verification)

You don’t need to run this manually for `drift-checker` to work—`drift-checker` generates and reads a plan internally.

If you want to sanity-check your runner first:

```bash
# OpenTofu (preferred if available)
tofu plan -refresh-only -out=drift-checker.plan
tofu show -json drift-checker.plan | jq . > /dev/null

# Terraform (fallback)
terraform plan -refresh-only -out=drift-checker.plan
terraform show -json drift-checker.plan | jq . > /dev/null
```

### 4) Scan for drift (plan + parse in one step)

```bash
# Markdown output (default), exits 0 if clean; use --strict to exit 2 on drift
drift-checker scan --path . --strict
```

Alternative formats:

```bash
# Plain text
drift-checker scan --path . --format text

# Machine-readable JSON (stdout is **only** the JSON)
drift-checker scan --path . --format json
```

---

## NEW: Destructive-change Gate (normal plan JSON)

Use `gate` to enforce a policy on **destructive actions** (deletes/replaces) from a standard plan JSON (output of `show -json`, *not* refresh-only).

### Typical usage

```bash
# Fail CI if any delete/replace appears
drift-checker gate --input plan.json --strict

# Fail if deletes or replaces exceed thresholds
drift-checker gate --input plan.json --strict --max-deletes 0 --max-replaces 0

# Output formats: md|text|json; include a list of destructive addresses
drift-checker gate --input plan.json --format json --list
```

### Exit code behavior

`drift-checker gate` uses standard shell exit codes so you can gate CI jobs.

| Scenario                                                    | Exit code |
| ----------------------------------------------------------- | --------- |
| Safe (no destructive and thresholds not exceeded)           | `0`       |
| Destructive present **or** thresholds exceeded (`--strict`) | `2`       |
| Any other error (I/O, invalid JSON, etc.)                   | `1`       |

> Tip: Use `--strict` in CI to fail the job on destructive changes: exit code `2` clearly distinguishes policy violations from other errors.

### JSON schema (gate output on stdout)

```json
{
  "updates": 0,
  "replaces": 0,
  "deletes": 0,
  "destructive_total": 0,
  "destructive": ["module.db.aws_db_instance.main"],
  "total": 0
}
```

* `destructive` is included only with `--list`.
* `destructive_total` = `replaces + deletes`.
* `total` equals the count of all changed resource addresses in the plan JSON (updates + deletes + replaces).

### What counts as *destructive*?

We look at `resource_changes[*].change.actions`:

* `["delete"]` → **destructive**
* `["create","delete"]` or `["delete","create"]` → **replace** → **destructive**
* `["update"]` → **non-destructive**

---

## Quickstart (General)

```bash
# In your IaC directory
tofu init || terraform init

# Drift detection
drift-checker scan --path . --strict

# Gate using a normal plan JSON you've already generated
terraform show -json drift-checker.plan > plan.json
drift-checker gate --input plan.json --strict --format md --list
```

## Implementation Notes

* `scan` prefers `tofu`, falls back to `terraform`, and uses a reliable two-step JSON flow.
* `gate` **does not** run a plan: it reads a **provided** plan JSON (normal `show -json` file).
* All logs are written to **stderr**; **stdout** is reserved for the user-selected output (Markdown/Text/JSON).

## Project Structure

* `cmd/` – CLI commands (Cobra)

  * `scan.go` – refresh-only scan
  * `gate.go` – destructive-change policy gate
* `internal/plan/` – Runner selection, plan execution, and JSON parsing
* `internal/report/` – Output formatting for scan summaries
* `internal/drift/` – Orchestration for scan

## Tests

These tests reuse your existing fixtures and patterns:

- `plan_drift.json` (has 1 update, 1 delete, 1 replace) → triggers exit code `2` in strict mode.
- New `plan_updates_only.json` → safe (exit `0` even with `--strict`).
- Threshold test sets `--max-deletes 0 --max-replaces 0` to trip exit `2`.

You can run them with your current CI (`go test -v ./...`).

## License

Apache-2.0
