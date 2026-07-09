package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/config"
	apierrors "github.com/adaptocms/adapto-cms-cli/internal/errors"
	"github.com/adaptocms/adapto-cms-cli/internal/httpclient"
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
		return nil, cfg, fmt.Errorf("authentication required: run 'adapto auth login' (or set ADAPTO_CLI_TOKEN / --token)")
	}
	return c, cfg, nil
}

// Ctx returns a background context.
func Ctx() context.Context {
	return context.Background()
}

// CheckErr checks an HTTP response for errors using the body bytes.
func CheckErr(httpResp *http.Response, body []byte) error {
	return apierrors.CheckResponse(httpResp, body)
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

func ParseCustomFields(cmd *cobra.Command) (*map[string]client.CustomFieldModel, error) {
	v, _ := cmd.Flags().GetString("custom-fields-json")
	if v == "" {
		return nil, nil
	}
	var cf map[string]client.CustomFieldModel
	dec := json.NewDecoder(strings.NewReader(v))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cf); err != nil {
		return nil, fmt.Errorf("invalid --custom-fields-json: %w", err)
	}
	return &cf, nil
}

// AddCustomFieldsFlag registers the standard --custom-fields-json flag.
func AddCustomFieldsFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().String("custom-fields-json", "", "Custom fields JSON object")
	}
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
