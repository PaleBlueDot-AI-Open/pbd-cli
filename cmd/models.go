package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/client"
	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/output"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Browse available models",
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List models available to you",
	Example: `  pbd-cli models list
  pbd-cli models list --output json`,
	RunE: runModelsList,
}

func init() {
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(modelsListCmd)
}

func runModelsList(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()
	c := client.NewClient(cfg)

	data, err := c.Get("/openIntelligence/api/user/models")
	if err != nil {
		return handleClientError(err)
	}

	// Default: raw JSON output
	if !isFormatOutput() {
		fmt.Println(string(data))
		return nil
	}

	var models []string
	if err := json.Unmarshal(data, &models); err != nil {
		var resp struct {
			Data []string `json:"data"`
		}
		json.Unmarshal(data, &resp)
		models = resp.Data
	}

	headers := []string{"MODEL"}
	var rows [][]string
	for _, m := range models {
		rows = append(rows, []string{m})
	}

	output.PrintTable(cmd.OutOrStdout(), headers, rows)
	return nil
}
