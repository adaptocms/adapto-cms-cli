package cmdutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/eggnita/adapto_cms_cli/internal/client"
	"github.com/eggnita/adapto_cms_cli/internal/config"
	apierrors "github.com/eggnita/adapto_cms_cli/internal/errors"
	"github.com/eggnita/adapto_cms_cli/internal/httpclient"
	"github.com/spf13/cobra"
)

// NewClient creates an API client from current config, validating that token is set.
func NewClient() (*client.ClientWithResponses, config.Config, error) {
	cfg := config.Load()
	c, err := httpclient.New(cfg)
	if err != nil {
		return nil, cfg, err
	}
	return c, cfg, nil
}

// NewClientWithAuth creates an API client and requires a token.
func NewClientWithAuth() (*client.ClientWithResponses, config.Config, error) {
	c, cfg, err := NewClient()
	if err != nil {
		return nil, cfg, err
	}
	if cfg.Token == "" {
		return nil, cfg, fmt.Errorf("authentication required: set ADAPTO_TOKEN or --token")
	}
	return c, cfg, nil
}

// Ctx returns a background context.
func Ctx() context.Context {
	return context.Background()
}

// CheckErr checks an HTTP response for errors using the body bytes.
func CheckErr(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	// Create a minimal http.Response for the error checker
	return apierrors.CheckHTTP(statusCode, body)
}

// StringPtr returns a pointer to a string, or nil if empty.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IntPtr returns a pointer to an int, or nil if zero.
func IntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// StringSlicePtr returns a pointer to a string slice parsed from comma-separated input, or nil if empty.
func StringSlicePtr(s string) *[]string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return &parts
}

// AddListFlags adds common pagination/sorting flags to a command.
func AddListFlags(cmd *cobra.Command) {
	cmd.Flags().String("keyword", "", "Search keyword")
	cmd.Flags().String("language", "", "Filter by language")
	cmd.Flags().String("field", "", "Sort field")
	cmd.Flags().String("order", "", "Sort order (asc/desc)")
	cmd.Flags().Int("page", 0, "Page number")
	cmd.Flags().Int("limit", 0, "Items per page")
}
