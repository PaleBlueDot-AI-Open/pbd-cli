package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/client"
	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/output"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Query usage and billing",
}

var usageBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show balance and subscription status",
	Example: `  pbd-cli usage balance
  pbd-cli usage balance -f`,
	RunE: runUsageBalance,
}

var (
	logsLimit int
	logsPage  int
	logsModel string
	logsToken string
)

func init() {
	rootCmd.AddCommand(usageCmd)
	usageCmd.AddCommand(usageBalanceCmd)
}

func runUsageBalance(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()
	c := client.NewClient(cfg)

	userData, err := c.Get("/openIntelligence/api/user/self")
	if err != nil {
		return handleClientError(err)
	}

	var userResp struct {
		Data struct {
			Quota     int64 `json:"quota"`
			UsedQuota int64 `json:"used_quota"`
		} `json:"data"`
	}
	json.Unmarshal(userData, &userResp)

	subData, subErr := c.Get("/openIntelligence/api/subscription/self")

	// Default: raw JSON output
	if !isFormatOutput() {
		result := map[string]interface{}{
			"user": json.RawMessage(userData),
		}
		if subErr == nil {
			result["subscription"] = json.RawMessage(subData)
		} else {
			result["subscription"] = nil
		}
		output.PrintJSON(cmd.OutOrStdout(), result)
		return nil
	}

	// Formatted output with -f flag
	remaining := userResp.Data.Quota - userResp.Data.UsedQuota
	fmt.Printf("Quota:     %d\n", userResp.Data.Quota)
	fmt.Printf("Used:      %d\n", userResp.Data.UsedQuota)
	fmt.Printf("Remaining: %d\n", remaining)

	if subErr != nil {
		fmt.Println("Subscription: (unavailable)")
	} else {
		var subResp struct {
			Data []struct {
				PlanName string `json:"plan_name"`
				Status   int    `json:"status"`
			} `json:"data"`
		}
		json.Unmarshal(subData, &subResp)
		if len(subResp.Data) == 0 {
			fmt.Println("Subscription: none")
		} else {
			status := "active"
			if subResp.Data[0].Status != 1 {
				status = "inactive"
			}
			fmt.Printf("Subscription: %s (%s)\n", subResp.Data[0].PlanName, status)
		}
	}

	return nil
}

var usageLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show usage logs",
	Example: `  pbd-cli usage logs
  pbd-cli usage logs --limit 50 --model gpt-4o`,
	RunE: runUsageLogs,
}

func init() {
	usageCmd.AddCommand(usageLogsCmd)
	usageLogsCmd.Flags().IntVar(&logsLimit, "limit", 20, "page size")
	usageLogsCmd.Flags().IntVar(&logsPage, "page", 1, "page number")
	usageLogsCmd.Flags().StringVar(&logsModel, "model", "", "filter by model")
	usageLogsCmd.Flags().StringVar(&logsToken, "token", "", "filter by token")
}

func runUsageLogs(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()
	c := client.NewClient(cfg)

	path := fmt.Sprintf("/openIntelligence/api/log/self/?p=%d&page_size=%d", logsPage, logsLimit)
	if logsModel != "" {
		path += "&model_name=" + logsModel
	}
	if logsToken != "" {
		path += "&token_name=" + logsToken
	}

	data, err := c.Get(path)
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
		Data []struct {
			CreatedAt int64  `json:"created_at"`
			ModelName string `json:"model_name"`
			Quota     int64  `json:"quota"`
			TokenName string `json:"token_name"`
		} `json:"data"`
	}
	json.Unmarshal(data, &resp)

	headers := []string{"TIME", "MODEL", "TOKENS", "TOKEN NAME"}
	var rows [][]string
	for _, log := range resp.Data {
		rows = append(rows, []string{
			output.FormatTime(log.CreatedAt),
			log.ModelName,
			strconv.FormatInt(log.Quota, 10),
			log.TokenName,
		})
	}

	output.PrintTable(cmd.OutOrStdout(), headers, rows)
	return nil
}
