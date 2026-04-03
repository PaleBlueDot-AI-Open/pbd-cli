package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:     "set --base-url <url>",
	Short:   "Set configuration values",
	Example: `  pbd-cli config set --base-url https://www.palebluedot.ai`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := getConfig()
		if err != nil {
			return err
		}

		if burl := viper.GetString("base_url"); burl != "" {
			cfg.BaseURL = burl
		}

		if err := saveConfig(cfg); err != nil {
			return err
		}

		fmt.Printf("Base URL set to: %s\n", cfg.BaseURL)
		return nil
	},
}

var configViewCmd = &cobra.Command{
	Use:     "view",
	Short:   "View current configuration",
	Example: `  pbd-cli config view`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := getConfig()
		if err != nil {
			return err
		}

		fmt.Printf("Base URL: %s\n", cfg.BaseURL)
		if cfg.UserID != 0 {
			fmt.Printf("User ID:  %d\n", cfg.UserID)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configViewCmd)
}
