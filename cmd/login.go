package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/auth"
	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/config"
)

var (
	loginManual bool
	loginPort   int
)

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)

	loginCmd.Flags().BoolVar(&loginManual, "manual", false, "manual login mode (enter cookie manually)")
	loginCmd.Flags().IntVarP(&loginPort, "port", "p", 0, "port for callback server (default: auto-select 8080-8090)")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to PaleBlueDot",
	Long:  "Log in via browser (default) or manually with --manual flag.",
	Example: `  pbd-cli login                  # Browser login (default)
  pbd-cli login --manual         # Manual login
  pbd-cli login --port 8085      # Use specific port`,
	RunE: runLogin,
}

var logoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Log out and clear local session",
	Example: `  pbd-cli logout`,
	RunE:    runLogout,
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Manual mode: use existing flow
	if loginManual {
		return runManualLogin()
	}

	// Auto mode: browser login
	return runBrowserLogin()
}

func runManualLogin() error {
	cfg, _ := getConfig()
	if burl := viper.GetString("base_url"); burl != "" {
		cfg.BaseURL = burl
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Session Cookie: ")
	cookie, _ := reader.ReadString('\n')
	cookie = strings.TrimSpace(cookie)

	if cookie == "" {
		return fmt.Errorf("cookie is required")
	}

	fmt.Print("User ID: ")
	userIDStr, _ := reader.ReadString('\n')
	userIDStr = strings.TrimSpace(userIDStr)

	var userID int
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	cfg.Cookie = cookie
	cfg.UserID = userID

	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("Logged in as user ID: %d\n", cfg.UserID)
	return nil
}

func runBrowserLogin() error {
	cfg, _ := getConfig()
	if burl := viper.GetString("base_url"); burl != "" {
		cfg.BaseURL = burl
	}

	// Use environment-aware base URL for login
	loginBaseURL := config.GetBaseURL()

	// Start callback server
	srv := auth.NewServer(loginPort)
	ctx := context.Background()

	_, err := srv.Start(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Please use --manual mode instead.")
		return err
	}
	defer srv.Shutdown()

	// Build login URL
	loginURL := buildLoginURL(loginBaseURL, srv.CallbackURL())

	// Open browser
	fmt.Printf("Opening browser for login...\n")
	fmt.Printf("If browser doesn't open, visit: %s\n", loginURL)

	if err := browser.OpenURL(loginURL); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to open browser: %v\n", err)
		fmt.Printf("Please open this URL manually: %s\n", loginURL)
	}

	// Wait for callback with 5 minute timeout
	fmt.Println("Waiting for login...")
	result := srv.WaitWithTimeout(5 * time.Minute)

	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", result.Err)
		return result.Err
	}

	// Call pbd-login API to exchange token for session
	fmt.Println("Exchanging token for session...")
	loginResp, err := auth.ExchangeToken(cfg.BaseURL, result.Token, result.UserID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to exchange token: %v\n", err)
		return err
	}

	// Save config - session value needs to be formatted as cookie
	cfg.Cookie = fmt.Sprintf("session=%s", loginResp.Data.Session)
	cfg.UserID = loginResp.Data.ID

	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("Login successful! Logged in as user ID: %d\n", cfg.UserID)
	return nil
}

func buildLoginURL(baseURL, callbackURL string) string {
	encodedCallback := url.QueryEscape(callbackURL)
	return fmt.Sprintf("%s/login?redirect_uri=%s", baseURL, encodedCallback)
}

func runLogout(cmd *cobra.Command, args []string) error {
	path, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}
	if cfgFile != "" {
		path = cfgFile
	}

	if err := config.ClearSession(path); err != nil {
		return err
	}

	fmt.Println("Logged out successfully")
	return nil
}
