package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/client"
	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/output"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens",
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tokens",
	Example: `  pbd-cli token list
  pbd-cli token list -f`,
	RunE: runTokenList,
}

var (
	tokenName    string
	tokenQuota   int64
	tokenExpires int64
	tokenModels  string
)

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(tokenListCmd)
}

// Token represents an API token from the platform.
type Token struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	RemainQuota    int64  `json:"remain_quota"`
	UsedQuota      int64  `json:"used_quota"`
	UnlimitedQuota bool   `json:"unlimited_quota"`
	ModelLimits    string `json:"model_limits"`
	Status         int    `json:"status"`
}

func runTokenList(cmd *cobra.Command, args []string) error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	c := client.NewClient(cfg)
	data, err := c.Get("/openIntelligence/api/token/")
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
		Data []Token `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	headers := []string{"ID", "NAME", "QUOTA", "USED", "MODELS", "STATUS"}
	var rows [][]string
	for _, t := range resp.Data {
		quota := output.FormatQuota(t.UnlimitedQuota, t.RemainQuota)
		models := t.ModelLimits
		if models == "" {
			models = "*"
		}
		status := "enabled"
		if t.Status != 1 {
			status = "disabled"
		}
		rows = append(rows, []string{
			strconv.Itoa(t.ID),
			t.Name,
			quota,
			strconv.FormatInt(t.UsedQuota, 10),
			models,
			status,
		})
	}

	output.PrintTable(cmd.OutOrStdout(), headers, rows)
	return nil
}

func handleClientError(err error) error {
	if client.IsAuthError(err) {
		return fmt.Errorf("session expired — please run 'pbd-cli login'")
	}
	return err
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create --name <name>",
	Short: "Create a new token",
	Example: `  pbd-cli token create --name dev-key
  pbd-cli token create --name prod-key --quota 100000 --models gpt-4o,claude-3-5`,
	RunE: runTokenCreate,
}

func init() {
	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCreateCmd.Flags().StringVar(&tokenName, "name", "", "token name (required)")
	tokenCreateCmd.Flags().Int64Var(&tokenQuota, "quota", 0, "remaining quota")
	tokenCreateCmd.Flags().Int64Var(&tokenExpires, "expires", -1, "expiry timestamp")
	tokenCreateCmd.Flags().StringVar(&tokenModels, "models", "", "model allowlist")
	tokenCreateCmd.MarkFlagRequired("name")
}

func runTokenCreate(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()
	c := client.NewClient(cfg)

	reqBody := map[string]interface{}{"name": tokenName}

	if tokenQuota > 0 {
		reqBody["remain_quota"] = tokenQuota
		reqBody["unlimited_quota"] = false
	} else {
		reqBody["unlimited_quota"] = true
	}

	reqBody["expired_time"] = tokenExpires

	if tokenModels != "" {
		reqBody["model_limits_enabled"] = true
		reqBody["model_limits"] = tokenModels
	}

	_, err := c.Post("/openIntelligence/api/token/", reqBody)
	if err != nil {
		return handleClientError(err)
	}

	fmt.Println("Token created successfully.")
	fmt.Printf("Name: %s\n", tokenName)
	fmt.Println("Run 'pbd-cli token list' to find the token ID, then 'pbd-cli token get-key <id>' to retrieve the key.")
	return nil
}

var tokenDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a token",
	Example: `  pbd-cli token delete 5`,
	Args:    cobra.ExactArgs(1),
	RunE:    runTokenDelete,
}

var tokenGetKeyCmd = &cobra.Command{
	Use:     "get-key <id>",
	Short:   "Get plaintext key for a token",
	Example: `  pbd-cli token get-key 5`,
	Args:    cobra.ExactArgs(1),
	RunE:    runTokenGetKey,
}

func init() {
	tokenCmd.AddCommand(tokenDeleteCmd)
	tokenCmd.AddCommand(tokenGetKeyCmd)
}

func runTokenDelete(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()
	c := client.NewClient(cfg)

	_, err := c.Delete("/openIntelligence/api/token/" + args[0])
	if err != nil {
		return handleClientError(err)
	}

	fmt.Printf("Token %s deleted\n", args[0])
	return nil
}

func runTokenGetKey(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()
	c := client.NewClient(cfg)

	data, err := c.Post("/openIntelligence/api/token/"+args[0]+"/key", nil)
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
		Data struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	fmt.Printf("Name: %s\n", resp.Data.Name)
	fmt.Printf("Key:  %s\n", resp.Data.Key)
	return nil
}
