package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/client"
)

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Show wallet balance",
	Example: `  pbd-cli wallet
  pbd-cli wallet -f`,
	RunE: runWallet,
}

func init() {
	rootCmd.AddCommand(walletCmd)
}

func runWallet(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()
	c := client.NewClient(cfg)

	data, err := c.Get("/openIntelligence/api/user/self/balance")
	if err != nil {
		return handleClientError(err)
	}

	// Default: raw JSON output
	if !isFormatOutput() {
		fmt.Println(string(data))
		return nil
	}

	// Formatted output with -f flag
	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Balance     float64 `json:"balance"`
			GiftBalance float64 `json:"giftBalance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	fmt.Printf("Balance:      %.6f\n", resp.Data.Balance)
	fmt.Printf("Gift Balance: %.6f\n", resp.Data.GiftBalance)

	return nil
}
