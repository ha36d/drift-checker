# drift-checker

A CLI tool for drift detection and destructive-change gating in Terraform/OpenTofu plans.

## Features

- Detects drift using **refresh-only** plans (OpenTofu preferred, Terraform fallback)
- Parses plan JSON and counts **updates / deletes / replaces**
- Outputs **Markdown** (default), **text**, or **json** summary (counts + drifted resources)
- Flags: `--path` (default `.`), `--format md|text|json` (default `md`), `--strict` (exit code 2 if drift detected)
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

### Sample Markdown output

#### Clean (no drift)

```markdown
## Drift Summary (tofu)

- **Updates**: 0
- **Replaces**: 0
- **Deletes**: 0
- **Total changed resources**: 0

_No drift detected._
```

#### Drift detected (example)

```markdown
## Drift Summary (terraform)

- **Updates**: 1
- **Replaces**: 1
- **Deletes**: 0
- **Total changed resources**: 2

### Drifted Resources

- `module.vpc.aws_subnet.public[0]`
- `aws_iam_role.app`
```

> Notes:
>
> * Runner label will be `tofu` or `terraform` depending on what’s found in `PATH`.
> * The list shows resource addresses with detected drift.

---

### Exit code behavior

`drift-checker scan` uses standard shell exit codes so you can gate CI jobs.

| Scenario                                     | `--strict` absent | `--strict` present |
| -------------------------------------------- | ----------------- | ------------------ |
| No drift detected                            | `0`               | `0`                |
| Drift detected                               | `0`               | `2`                |
| Any other error (timeouts, bad config, etc.) | `1`               | `1`                |

> Tip: Use `--strict` in CI to fail the job on drift: exit code `2` clearly distinguishes drift from other errors.

---

## Quickstart (General)

```bash
# In your IaC directory
tofu init || terraform init

# Run scan (Markdown output by default)
drift-checker scan --path . --strict

# Plain text output
drift-checker scan --path . --format text

# Machine-readable JSON output (stdout is **only** the JSON)
drift-checker scan --path . --format json
```

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

---

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
