package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FindAvailablePort finds an available TCP port in the given range [startPort, endPort].
func FindAvailablePort(startPort, endPort int) (int, error) {
	if startPort > endPort {
		return 0, fmt.Errorf("start port %d is greater than end port %d", startPort, endPort)
	}

	for port := startPort; port <= endPort; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available port in range %d-%d", startPort, endPort)
}

// ParseCallbackParams extracts token and userId from callback URL.
func ParseCallbackParams(urlPath string) (token string, userID string, err error) {
	if !strings.HasPrefix(urlPath, "/callback?") {
		return "", "", fmt.Errorf("invalid callback path: %s", urlPath)
	}

	query := strings.TrimPrefix(urlPath, "/callback?")
	values := parseQuery(query)

	token = values.Get("token")
	if token == "" {
		return "", "", fmt.Errorf("missing token parameter")
	}

	userID = values.Get("userId")
	if userID == "" {
		return "", "", fmt.Errorf("missing userId parameter")
	}

	return token, userID, nil
}

// urlValues is a simple query string parser.
type urlValues map[string][]string

func (v urlValues) Get(key string) string {
	if vs := v[key]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func parseQuery(query string) urlValues {
	values := make(urlValues)
	for _, pair := range strings.Split(query, "&") {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := unescape(kv[0])
		val := unescape(kv[1])
		values[key] = append(values[key], val)
	}
	return values
}

func unescape(s string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			hex := s[i+1 : i+3]
			if b, err := strconv.ParseInt(hex, 16, 32); err == nil {
				result += string(rune(b))
				i += 2
				continue
			}
		}
		result += string(s[i])
	}
	return result
}

// Result holds the callback result.
type Result struct {
	Token  string
	UserID string
	Err    error
}

// Server is a local HTTP server for OAuth callback.
type Server struct {
	port      int
	server    *http.Server
	result    *Result
	done      chan struct{}
	closeOnce sync.Once // Ensures done is closed only once
	mu        sync.Mutex
	tmpl      *template.Template
}

const resultTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>{{.Title}}</title>
  <style>
    body { font-family: system-ui, -apple-system, sans-serif; text-align: center; padding: 50px; }
    .success { color: #22c55e; }
    .error { color: #ef4444; }
    code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; }
  </style>
</head>
<body>
  <h1 class="{{.Class}}">{{.Icon}} {{.Title}}</h1>
  <p>{{.Message}}</p>
  {{if .Hint}}<p>{{.Hint}}</p>{{end}}
  <script>setTimeout(function() { window.close(); }, 2000);</script>
</body>
</html>`

// NewServer creates a new callback server.
// If port is 0, an available port will be auto-assigned.
func NewServer(port int) *Server {
	tmpl := template.Must(template.New("result").Parse(resultTemplate))

	return &Server{
		port: port,
		done: make(chan struct{}),
		tmpl: tmpl,
	}
}

// closeDone safely closes the done channel exactly once.
func (s *Server) closeDone() {
	s.closeOnce.Do(func() {
		close(s.done)
	})
}

// Start starts the HTTP server and returns the actual port.
func (s *Server) Start(ctx context.Context) (int, error) {
	// Find available port if not specified
	if s.port == 0 {
		port, err := FindAvailablePort(8080, 8090)
		if err != nil {
			return 0, err
		}
		s.port = port
	}

	// Create listener
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return 0, fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	// Get actual port (in case port was 0)
	s.port = ln.Addr().(*net.TCPAddr).Port

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)
	s.server = &http.Server{
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.setResult(&Result{Err: err})
		}
	}()

	return s.port, nil
}

// handleCallback handles the OAuth callback request.
func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	token, userID, err := ParseCallbackParams(r.URL.String())
	if err != nil {
		s.renderResult(w, "Login Failed", "error", "✗", err.Error(),
			"Please try again or use pbd-cli login --manual")
		s.setResult(&Result{Err: err})
		s.closeDone()
		return
	}

	s.renderResult(w, "Login Successful", "success", "✓",
		"You can close this page now.", "")
	s.setResult(&Result{Token: token, UserID: userID})
	s.closeDone()
}

// renderResult renders the result HTML page.
func (s *Server) renderResult(w http.ResponseWriter, title, class, icon, message, hint string) {
	data := struct {
		Title   string
		Class   string
		Icon    string
		Message string
		Hint    string
	}{
		Title:   title,
		Class:   class,
		Icon:    icon,
		Message: message,
		Hint:    hint,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	s.tmpl.Execute(w, data)
}

// Wait waits for the callback result.
func (s *Server) Wait() *Result {
	<-s.done
	return s.getResult()
}

// WaitWithTimeout waits for the callback result with a timeout.
func (s *Server) WaitWithTimeout(timeout time.Duration) *Result {
	select {
	case <-s.done:
		return s.getResult()
	case <-time.After(timeout):
		s.closeDone() // Ensure done is closed on timeout too
		return &Result{Err: fmt.Errorf("login timed out after %v", timeout)}
	}
}

// Shutdown shuts down the server.
func (s *Server) Shutdown() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

// CallbackURL returns the callback URL for this server.
func (s *Server) CallbackURL() string {
	return fmt.Sprintf("http://localhost:%d/callback", s.port)
}

func (s *Server) setResult(result *Result) {
	s.mu.Lock()
	s.result = result
	s.mu.Unlock()
}

// getResult returns the current result.
func (s *Server) getResult() *Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result
}

// PbdLoginRequest is the request body for pbd-login API.
type PbdLoginRequest struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
}

// PbdLoginResponse is the response from pbd-login API.
type PbdLoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		DisplayName string `json:"display_name"`
		ID          int    `json:"id"`
		Session     string `json:"cliToken"`
		Username    string `json:"username"`
	} `json:"data"`
}

// ExchangeToken calls pbd-login API to exchange temporary token for session.
func ExchangeToken(baseURL, token, userID string) (*PbdLoginResponse, error) {
	reqBody := PbdLoginRequest{
		Token:  token,
		UserID: userID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/openIntelligence/api/user/pbd-login", baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call pbd-login API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pbd-login failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result PbdLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("pbd-login failed: %s", result.Message)
	}

	if result.Data.ID == 0 {
		return nil, fmt.Errorf("pbd-login returned empty id")
	}

	return &result, nil
}
