package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	drift "github.com/ha36d/drift-checker/internal/drift"
)

var (
	timeout     time.Duration
	forceUpdate bool // kept (unused) to avoid breaking previous flags
	pathFlag    string
	formatFlag  string
	strictFlag  bool
)

// scanCmd represents the infrastructure scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for drift in Terraform/OpenTofu state via refresh-only plan",
	Long: `Runs a refresh-only plan using OpenTofu (preferred) or Terraform, parses the JSON plan,
and reports counts of updates / deletes / replaces, plus the list of drifted resources.

Exit status:
  0 = no drift
  2 = drift detected (when --strict is set)
  1 = other error`,
	Example: `  drift-checker scan
  drift-checker scan --path . --format md --strict
  drift-checker scan --timeout 30m`,
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().DurationVar(&timeout, "timeout", 2*time.Hour, "timeout for the scan operation")
	scanCmd.Flags().BoolVar(&forceUpdate, "force", false, "deprecated: no-op (kept for compatibility)")
	scanCmd.Flags().StringVar(&pathFlag, "path", ".", "working directory containing the Terraform/OpenTofu configuration")
	scanCmd.Flags().StringVar(&formatFlag, "format", "md", "output format: md|text")
	scanCmd.Flags().BoolVar(&strictFlag, "strict", false, "exit with code 2 if drift is detected")
}

func runScan(cmd *cobra.Command, args []string) error {
	// Create cancellable context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Handle OS interrupts
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Warn("Received interrupt signal, initiating graceful shutdown...")
		cancel()
	}()

	log.WithFields(log.Fields{
		"path":    pathFlag,
		"format":  formatFlag,
		"strict":  strictFlag,
		"timeout": timeout,
	}).Info("Scan parameters")

	res, err := drift.CheckDrift(ctx, drift.Options{
		Path:   pathFlag,
		Format: formatFlag,
		Strict: strictFlag,
	})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("scan operation timed out after %v", timeout)
		}
		return fmt.Errorf("scan failed: %w", err)
	}

	// Print report
	fmt.Println(res.RenderedReport)

	// Strict mode: exit 2 if drift
	if strictFlag && res.DriftDetected {
		// Using os.Exit(2) to conform to required contract
		os.Exit(2)
	}

	return nil
}
