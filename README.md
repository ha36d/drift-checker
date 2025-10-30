# drift-checker

A CLI tool for drift detection and destructive-change gating in Terraform/OpenTofu plans.

## Features
- Detects drift using refresh-only plans
- Enforces destructive-change policies (delete/replace)
- Outputs in Markdown, JSON, or text
- CI-ready with strict exit codes

## Quickstart

```bash
# Detect drift
terraform plan -refresh-only -json > plan.json
./drift-checker scan --path . --format md --strict

# Gate destructive changes
terraform show -json plan.bin | ./drift-checker gate --strict
```

## Project Structure
- `cmd/drift-checker/` - CLI entrypoint
- `internal/plan/` - Plan parsing logic
- `internal/drift/` - Drift detection
- `internal/report/` - Output formatting
- `testdata/` - Test fixtures

## License
Apache-2.0
