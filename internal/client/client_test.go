package client

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PaleBlueDot-AI-Open/pbd-cli/internal/config"
)

func TestDoAuthenticatedRequest(t *testing.T) {
	var receivedCookie, receivedUserHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCookie = r.Header.Get("Cookie")
		receivedUserHeader = r.Header.Get("New-Api-User")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, Cookie: "session=test-session", UserID: 123}
	c := NewClient(cfg)
	_, err := c.Get("/api/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if receivedCookie != "session=test-session" {
		t.Errorf("Cookie = %s", receivedCookie)
	}
	if receivedUserHeader != "123" {
		t.Errorf("New-Api-User = %s", receivedUserHeader)
	}
}

func TestHandle401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, Cookie: "test", UserID: 1}
	c := NewClient(cfg)

	_, err := c.Get("/api/test")
	if !IsAuthError(err) {
		t.Errorf("expected auth error, got %v", err)
	}
}

func TestHandle403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, Cookie: "test", UserID: 1}
	c := NewClient(cfg)

	_, err := c.Get("/api/test")
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHandle500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, Cookie: "test", UserID: 1}
	c := NewClient(cfg)

	_, err := c.Get("/api/test")
	if !errors.Is(err, ErrServer) {
		t.Errorf("expected ErrServer, got %v", err)
	}
}

func TestPost(t *testing.T) {
	var receivedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, Cookie: "test", UserID: 1}
	c := NewClient(cfg)

	body := map[string]string{"key": "value"}
	_, err := c.Post("/api/test", body)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}

	if string(receivedBody) != `{"key":"value"}` {
		t.Errorf("body = %s", receivedBody)
	}
}

func TestDelete(t *testing.T) {
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, Cookie: "test", UserID: 1}
	c := NewClient(cfg)

	_, err := c.Delete("/api/test")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if method != http.MethodDelete {
		t.Errorf("method = %s", method)
	}
}

func TestGetResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, Cookie: "test", UserID: 1}
	c := NewClient(cfg)

	body, err := c.Get("/api/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(body) != `{"result":"ok"}` {
		t.Errorf("body = %s", body)
	}
}
