package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/client"
)

var chatCmd = &cobra.Command{
	Use:   "chat --model <model> --message <message>",
	Short: "Send a chat completion request",
	Long: `Send a single-turn chat message.

Requires a Bearer token (API key), not session auth.
Use --token or set api_key in config.`,
	Example: `  pbd-cli chat --model gpt-4o --message "Hello"
  pbd-cli chat --model claude-3-5-sonnet --message "Hi" --token sk-xxx`,
	RunE: runChat,
}

var (
	chatModel   string
	chatMessage string
	chatToken   string
)

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().StringVar(&chatModel, "model", "", "model name (required)")
	chatCmd.Flags().StringVar(&chatMessage, "message", "", "message (required)")
	chatCmd.Flags().StringVar(&chatToken, "token", "", "API token")
	chatCmd.MarkFlagRequired("model")
	chatCmd.MarkFlagRequired("message")
}

func runChat(cmd *cobra.Command, args []string) error {
	cfg, _ := getConfig()

	token := chatToken
	if token == "" {
		token = cfg.APIKey
	}
	if token == "" {
		return fmt.Errorf("API token required: use --token or set api_key in config")
	}

	reqBody := map[string]interface{}{
		"model": chatModel,
		"messages": []map[string]string{
			{"role": "user", "content": chatMessage},
		},
	}

	c := client.NewTokenClient(cfg, token)
	data, err := c.Post("/v1/chat/completions", reqBody)
	if err != nil {
		return err
	}

	// Default: raw JSON output
	if !isFormatOutput() {
		fmt.Println(string(data))
		return nil
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(data, &resp)

	if len(resp.Choices) > 0 {
		fmt.Println(resp.Choices[0].Message.Content)
	}
	return nil
}
