package httpclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/adaptocms/adapto-cms-cli/internal/config"
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
	"github.com/adaptocms/adapto-cms-cli/internal/httpclient"
)

type recordedRequest struct {
	Path  string
	Token string
}

type mockAPI struct {
	mu       sync.Mutex
	requests []recordedRequest
	refresh  func(w http.ResponseWriter)
	me       func(token string, w http.ResponseWriter)
}

func (m *mockAPI) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		if auth := r.Header.Get("Authorization"); len(auth) > 7 {
			token = auth[7:]
		}
		m.mu.Lock()
		m.requests = append(m.requests, recordedRequest{Path: r.URL.Path, Token: token})
		m.mu.Unlock()

		switch r.URL.Path {
		case "/auth/refresh":
			m.refresh(w)
		case "/auth/me":
			m.me(token, w)
		default:
			http.Error(w, `{"detail":"Not Found"}`, http.StatusNotFound)
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func TestRefreshRetryOn401(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := credentials.Save(&credentials.Credentials{AccessToken: "expired", RefreshToken: "refresh-1"}); err != nil {
		t.Fatal(err)
	}

	api := &mockAPI{
		refresh: func(w http.ResponseWriter) {
			writeJSON(w, 200, map[string]string{"access_token": "fresh", "refresh_token": "refresh-2"})
		},
		me: func(token string, w http.ResponseWriter) {
			if token != "fresh" {
				writeJSON(w, 401, map[string]string{"detail": "expired"})
				return
			}
			writeJSON(w, 200, map[string]any{"user": map[string]any{"id": "u1", "email": "a@b.c", "status": "active", "is_email_verified": true}})
		},
	}
	server := httptest.NewServer(api.handler())
	defer server.Close()

	c, err := httpclient.New(config.Config{APIURL: server.URL, Token: "expired", TokenFromCreds: true})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetMeAuthMeGetWithResponse(context.Background())
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode() != 200 {
		t.Fatalf("status = %d, want 200 after refresh retry", resp.StatusCode())
	}

	wantSequence := []recordedRequest{
		{Path: "/auth/me", Token: "expired"},
		{Path: "/auth/refresh", Token: ""},
		{Path: "/auth/me", Token: "fresh"},
	}
	if len(api.requests) != len(wantSequence) {
		t.Fatalf("requests = %+v, want %+v", api.requests, wantSequence)
	}
	for i, want := range wantSequence {
		if api.requests[i] != want {
			t.Fatalf("request[%d] = %+v, want %+v", i, api.requests[i], want)
		}
	}

	creds, err := credentials.Load()
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessToken != "fresh" || creds.RefreshToken != "refresh-2" {
		t.Fatalf("credentials not rotated: %+v", creds)
	}
}

func TestSessionExpiredWhenRefreshFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := credentials.Save(&credentials.Credentials{AccessToken: "expired", RefreshToken: "dead"}); err != nil {
		t.Fatal(err)
	}

	api := &mockAPI{
		refresh: func(w http.ResponseWriter) {
			writeJSON(w, 401, map[string]string{"detail": "refresh token expired"})
		},
		me: func(token string, w http.ResponseWriter) {
			writeJSON(w, 401, map[string]string{"detail": "expired"})
		},
	}
	server := httptest.NewServer(api.handler())
	defer server.Close()

	c, err := httpclient.New(config.Config{APIURL: server.URL, Token: "expired", TokenFromCreds: true})
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.GetMeAuthMeGetWithResponse(context.Background())
	if !errors.Is(err, httpclient.ErrSessionExpired) {
		t.Fatalf("err = %v, want ErrSessionExpired", err)
	}
}

func TestNoRefreshForExplicitToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := credentials.Save(&credentials.Credentials{AccessToken: "stored", RefreshToken: "refresh-1"}); err != nil {
		t.Fatal(err)
	}

	api := &mockAPI{
		me: func(token string, w http.ResponseWriter) {
			writeJSON(w, 401, map[string]string{"detail": "bad token"})
		},
	}
	server := httptest.NewServer(api.handler())
	defer server.Close()

	c, err := httpclient.New(config.Config{APIURL: server.URL, Token: "explicit", TokenFromCreds: false})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetMeAuthMeGetWithResponse(context.Background())
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode() != 401 {
		t.Fatalf("status = %d, want 401 passed through", resp.StatusCode())
	}
	if len(api.requests) != 1 {
		t.Fatalf("expected exactly 1 request (no refresh), got %+v", api.requests)
	}
}

func TestNewRequiresAPIURL(t *testing.T) {
	if _, err := httpclient.New(config.Config{}); err == nil {
		t.Fatal("expected error for missing API URL")
	}
}
