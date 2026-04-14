package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/eggnita/adapto_cms_cli/internal/client"
	"github.com/eggnita/adapto_cms_cli/internal/config"
)

// New creates a ClientWithResponses that injects Bearer token and X-Tenant-ID headers.
func New(cfg config.Config) (*client.ClientWithResponses, error) {
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("API URL is required (set ADAPTO_API_URL or --api-url)")
	}

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
	}

	if cfg.Verbose {
		opts = append(opts, client.WithHTTPClient(&http.Client{
			Transport: &verboseTransport{base: http.DefaultTransport},
		}))
	}

	return client.NewClientWithResponses(cfg.APIURL, opts...)
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
