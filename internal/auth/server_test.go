package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestFindAvailablePort(t *testing.T) {
	port, err := FindAvailablePort(8080, 8090)
	if err != nil {
		t.Fatalf("expected to find available port, got error: %v", err)
	}
	if port < 8080 || port > 8090 {
		t.Fatalf("expected port in range 8080-8090, got %d", port)
	}
}

func TestFindAvailablePortInvalidRange(t *testing.T) {
	_, err := FindAvailablePort(8090, 8080)
	if err == nil {
		t.Fatal("expected error when start > end, got nil")
	}
}

func TestParseCallbackParams(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantToken  string
		wantUserID string
		wantErr    bool
	}{
		{
			name:       "valid params",
			url:        "/callback?token=test_token&userId=123",
			wantToken:  "test_token",
			wantUserID: "123",
			wantErr:    false,
		},
		{
			name:       "token with special chars",
			url:        "/callback?token=abc%3D123%26xyz&userId=456",
			wantToken:  "abc=123&xyz",
			wantUserID: "456",
			wantErr:    false,
		},
		{
			name:    "missing token",
			url:     "/callback?userId=123",
			wantErr: true,
		},
		{
			name:    "missing userId",
			url:     "/callback?token=test",
			wantErr: true,
		},
		{
			name:    "empty token",
			url:     "/callback?token=&userId=123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, userID, err := ParseCallbackParams(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("token = %q, want %q", token, tt.wantToken)
			}
			if userID != tt.wantUserID {
				t.Errorf("userID = %q, want %q", userID, tt.wantUserID)
			}
		})
	}
}

func TestServerStartAndWait(t *testing.T) {
	srv := NewServer(0)

	ctx := context.Background()
	port, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Shutdown()

	// Simulate callback
	callbackURL := fmt.Sprintf("http://localhost:%d/callback?token=test_token&userId=123", port)
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("callback returned status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Wait for result with timeout
	result := srv.WaitWithTimeout(5 * time.Second)
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}
	if result.Token != "test_token" {
		t.Errorf("token = %q, want %q", result.Token, "test_token")
	}
	if result.UserID != "123" {
		t.Errorf("userID = %q, want %q", result.UserID, "123")
	}
}

func TestServerTimeout(t *testing.T) {
	srv := NewServer(0)

	ctx := context.Background()
	_, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Shutdown()

	// Don't send callback, wait for timeout
	result := srv.WaitWithTimeout(100 * time.Millisecond)
	if result.Err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestServerConcurrentCallbackAndTimeout(t *testing.T) {
	srv := NewServer(0)

	ctx := context.Background()
	port, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Shutdown()

	// Start a goroutine that will send a callback after a delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		callbackURL := fmt.Sprintf("http://localhost:%d/callback?token=test_token&userId=123", port)
		http.Get(callbackURL) //nolint:errcheck
	}()

	// Wait with a timeout that may race with the callback
	result := srv.WaitWithTimeout(100 * time.Millisecond)

	// Either success or timeout is acceptable, but no panic should occur
	if result.Err != nil && result.Token == "" {
		// Timeout is fine - the callback lost the race
	} else if result.Token == "test_token" {
		// Callback won the race
	} else {
		t.Errorf("unexpected result: %+v", result)
	}
}
