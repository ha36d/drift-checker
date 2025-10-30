package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
)

// Config represents the application configuration
type Config struct {
	Metadata   map[string]string           `yaml:"metadata"`
	Cloud      string                      `yaml:"cloud"`
	Account    string                      `yaml:"account"`
	Region     string                      `yaml:"region"`
	Components []map[string]map[string]any `yaml:"components"`
}

var (
	cfgFile string
	verbose bool
	config  Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "drift-checker",
	Short: "Cloud infrastructure drift checker",
	Long: `drift-checker is an IaC tool built to check drifts in cloud infrastructure,
supporting Terraform and OpenTofu.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Command execution failed: %v", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	must(viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")))
	viper.SetDefault("verbose", false)
}

func setupLogging() {
	if verbose {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp: true,
		})
	}
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		setupDefaultConfigPaths()
	}
	viper.AutomaticEnv()
	if err := loadAndValidateConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	if verbose {
		logConfiguration()
	}
}

func setupDefaultConfigPaths() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".config")
}

func loadAndValidateConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
		log.Warn("No config file found! Using flags and environment variables")
	}
	if err := validateConfigSchema(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	return nil
}

func validateConfigSchema() error {
	schemaPath := filepath.Join(".", "config_schema.json")
	schemaFile, err := os.ReadFile(schemaPath)
	if err != nil {
		// Schema is optional; only validate if present
		log.Debug("No config_schema.json found, skipping schema validation")
		return nil
	}
	schemaLoader := gojsonschema.NewStringLoader(string(schemaFile))
	documentLoader := gojsonschema.NewGoLoader(viper.AllSettings())
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}
	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("invalid configuration: %v", errors)
	}
	return nil
}

func logConfiguration() {
	log.Debug("Using config file:", viper.ConfigFileUsed())
	log.Debug("--- Configuration ---")
	for key, value := range viper.AllSettings() {
		log.Debugf("\t%s = %v", key, value)
	}
	log.Debug("---")
}

// must panics if err is not nil
func must(err error) {
	if err != nil {
		panic(err)
	}
}
