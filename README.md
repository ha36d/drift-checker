# drift-checker

A CLI tool for drift detection and destructive-change gating in Terraform/OpenTofu plans.

## Features

- Detects drift using **refresh-only** plans (OpenTofu preferred, Terraform fallback)
- Parses plan JSON and counts **updates / deletes / replaces**
- Outputs **Markdown** (default) or **text** summary (counts + drifted resources)
- Flags: `--path` (default `.`), `--format md|text` (default `md`), `--strict` (exit code 2 if drift detected)
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
````

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

## Project Structure

* `cmd/` – CLI commands (Cobra)
* `internal/plan/` – Runner selection, plan execution, and JSON parsing
* `internal/report/` – Output formatting
* `internal/drift/` – Orchestration

## License

Apache-2.0

```

---

### Notes / Next steps (already unblocked for you)

1. **CLI wiring** is complete: `--path`, `--format`, `--strict` work, exit code `2` on drift when strict.
2. **Plan parsing** counts updates/deletes/replaces and lists resource addresses.
3. Future hardening:
   - Add unit tests with `testdata/*.json` plans.
   - Consider supporting streaming `plan -json` when available (optional optimization).
   - Optionally include counts per module/type.

If you want, I can add tests + sample plan fixtures next.
```
