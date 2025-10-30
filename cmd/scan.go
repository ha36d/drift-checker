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
	forceUpdate bool
)

// scanCmd represents the infrastructure scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for the drifts in infrastructure",
	Long: `Scan the for the drifts in infrastructure.
This command will check for the drifts.`,
	Example: `  drift-checker scan
  drift-checker scan --timeout 30m
  drift-checker scan --force`,
	RunE: runBuild,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Add build-specific flags
	scanCmd.Flags().DurationVar(&timeout, "timeout", 2*time.Hour, "timeout for the scan operation")
	scanCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update even if no changes detected")
}

func runBuild(cmd *cobra.Command, args []string) error {
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

	// Validate configuration before proceeding
	if err := validateBuildConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log scan parameters
	logBuildParameters()

	// Execute the build
	log.Info("Starting infrastructure build...")
	startTime := time.Now()

	err := drift.CheckDrift(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("scan operation timed out after %v", timeout)
		}
		return fmt.Errorf("scan failed: %w", err)
	}

	// Log success and duration
	duration := time.Since(startTime)
	log.WithFields(log.Fields{
		"duration": duration.Round(time.Second),
	}).Info("Infrastructure scan completed successfully")

	return nil
}

func validateBuildConfig() error {
	return nil
}

func logBuildParameters() {
	log.WithFields(log.Fields{
		"timeout": timeout,
		"force":   forceUpdate,
	}).Info("Build parameters")
}
