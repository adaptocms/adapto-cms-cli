package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/config"
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
)

// ErrSessionExpired is returned when a 401 could not be recovered by refreshing.
var ErrSessionExpired = errors.New("session expired — run 'adapto auth login' to sign in again")

// New creates a ClientWithResponses that injects Bearer token and X-Tenant-ID headers.
func New(cfg config.Config) (*client.ClientWithResponses, error) {
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("API URL is required (set ADAPTO_API_URL or --api-url)")
	}

	var base http.RoundTripper = http.DefaultTransport
	if cfg.Verbose {
		base = &verboseTransport{base: base}
	}
	base = &authTransport{base: base, apiURL: cfg.APIURL, canRefresh: cfg.TokenFromCreds}

	opts := []client.ClientOption{
		client.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			if cfg.Token != "" {
				req.Header.Set("Authorization", "Bearer "+cfg.Token)
			}
			if cfg.TenantID != "" {
				req.Header.Set("X-Tenant-ID", cfg.TenantID)
			}
			return nil
		}),
		client.WithHTTPClient(&http.Client{Transport: base}),
	}

	return client.NewClientWithResponses(cfg.APIURL, opts...)
}

// authTransport retries a request once on 401 after refreshing the stored token.
type authTransport struct {
	base       http.RoundTripper
	apiURL     string
	canRefresh bool
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized || !t.canRefresh {
		return resp, err
	}
	// Can only retry if the body is replayable.
	if req.Body != nil && req.GetBody == nil {
		return resp, nil
	}

	newToken, rerr := refreshStoredToken(t.apiURL)
	if rerr != nil {
		drain(resp)
		return nil, ErrSessionExpired
	}
	drain(resp)

	retryReq := req.Clone(req.Context())
	if req.GetBody != nil {
		body, berr := req.GetBody()
		if berr != nil {
			return nil, berr
		}
		retryReq.Body = body
	}
	retryReq.Header.Set("Authorization", "Bearer "+newToken)
	return t.base.RoundTrip(retryReq)
}

func refreshStoredToken(apiURL string) (string, error) {
	creds, err := credentials.Load()
	if err != nil || creds.RefreshToken == "" {
		return "", fmt.Errorf("no refresh token available")
	}

	c, err := client.NewClientWithResponses(apiURL)
	if err != nil {
		return "", err
	}
	resp, err := c.RefreshAccessTokenAuthRefreshPostWithResponse(context.Background(), client.RefreshTokenRequest{
		RefreshToken: creds.RefreshToken,
	})
	if err != nil {
		return "", err
	}
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return "", fmt.Errorf("refresh failed: %d", resp.StatusCode())
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return "", err
	}
	newAccess, _ := data["access_token"].(string)
	if newAccess == "" {
		return "", fmt.Errorf("refresh returned no access token")
	}
	creds.AccessToken = newAccess
	if newRefresh, ok := data["refresh_token"].(string); ok && newRefresh != "" {
		creds.RefreshToken = newRefresh
	}
	_ = credentials.Save(creds)
	return newAccess, nil
}

func drain(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

type verboseTransport struct {
	base http.RoundTripper
}

func (t *verboseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Fprintf(os.Stderr, ">>> REQUEST\n%s\n", dump)

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	dump, _ = httputil.DumpResponse(resp, true)
	fmt.Fprintf(os.Stderr, "<<< RESPONSE\n%s\n", dump)
	return resp, nil
}
