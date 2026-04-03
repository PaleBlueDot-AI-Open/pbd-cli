package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/config"
)

var (
	cfgFile   string
	outputFmt string
	baseURL   string
	// Version information set by main
	version string
	commit  string
	date    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "pbd-cli",
	Short:   "CLI tool for PaleBlueDot AI platform",
	Long:    "pbd-cli is a command-line tool for managing API tokens,\nquerying usage, and browsing models on the PaleBlueDot AI platform.",
	Example: "  pbd-cli login\n  pbd-cli token list\n  pbd-cli usage balance --fromat",
}

// SetVersion sets the version information from main
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolP("format", "f", false, "format output (default: raw JSON)")
	rootCmd.PersistentFlags().String("base-url", "", "override base URL")

	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		path, err := config.DefaultConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		viper.SetConfigFile(path)
	}
	viper.ReadInConfig()
}

func getConfig() (*config.Config, error) {
	path, err := config.DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	if cfgFile != "" {
		path = cfgFile
	}
	return config.Load(path)
}

func saveConfig(cfg *config.Config) error {
	path, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}
	if cfgFile != "" {
		path = cfgFile
	}
	return config.Save(path, cfg)
}

func isFormatOutput() bool {
	return viper.GetBool("format")
}
