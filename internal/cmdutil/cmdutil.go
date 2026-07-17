package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/config"
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
	apierrors "github.com/adaptocms/adapto-cms-cli/internal/errors"
	"github.com/adaptocms/adapto-cms-cli/internal/httpclient"
	"github.com/adaptocms/adapto-cms-cli/internal/prompt"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// DefaultLanguage is the language a project falls back to when none is chosen.
const DefaultLanguage = "en-US"

var supportedLanguages = []struct {
	code string
	name string
}{
	{"en-US", "English"},
	{"es-ES", "Spanish"},
	{"fr-FR", "French"},
	{"de-DE", "German"},
	{"it-IT", "Italian"},
	{"pt-PT", "Portuguese"},
	{"ro-RO", "Romanian"},
	{"ru-RU", "Russian"},
	{"ja-JP", "Japanese"},
	{"ko-KR", "Korean"},
	{"zh-CN", "Chinese (Simplified)"},
	{"zh-TW", "Chinese (Traditional)"},
	{"ar-SA", "Arabic"},
	{"hi-IN", "Hindi"},
	{"nl-NL", "Dutch"},
	{"sv-SE", "Swedish"},
	{"da-DK", "Danish"},
	{"no-NO", "Norwegian"},
	{"fi-FI", "Finnish"},
	{"pl-PL", "Polish"},
	{"tr-TR", "Turkish"},
}

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

var expirationPresets = []struct {
	label string
	flag  string
	days  int
}{
	{"No expiration", "never", 0},
	{"30 days", "30d", 30},
	{"90 days", "90d", 90},
	{"1 year", "1y", 365},
}

// PromptExpiration resolves an API key expiry from a flag, or an interactive selector. Returns nil for no expiration.
func PromptExpiration(flag string) (*string, error) {
	var days int
	switch {
	case flag != "":
		matched := false
		for _, p := range expirationPresets {
			if p.flag == flag {
				days, matched = p.days, true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("invalid --expires-in %q (use one of: never, 30d, 90d, 1y)", flag)
		}
	case prompt.IsTTY():
		options := make([]huh.Option[string], 0, len(expirationPresets))
		for _, p := range expirationPresets {
			options = append(options, huh.NewOption(p.label, p.flag))
		}
		choice := "never"
		if err := huh.NewSelect[string]().Title("API key expiration").Options(options...).Value(&choice).Run(); err != nil {
			return nil, err
		}
		for _, p := range expirationPresets {
			if p.flag == choice {
				days = p.days
				break
			}
		}
	}

	if days <= 0 {
		return nil, nil
	}
	ts := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour).Format("2006-01-02T15:04:05")
	return &ts, nil
}

// RefreshSession re-mints the stored access token so newly created orgs and projects appear in its claims.
func RefreshSession() error {
	creds, err := credentials.Load()
	if err != nil || creds.RefreshToken == "" {
		return fmt.Errorf("no stored refresh token")
	}
	c, _, err := NewClient()
	if err != nil {
		return err
	}
	resp, err := c.RefreshAccessTokenAuthRefreshPostWithResponse(Ctx(), client.RefreshTokenRequest{RefreshToken: creds.RefreshToken})
	if err != nil {
		return err
	}
	if err := CheckErr(resp.HTTPResponse, resp.Body); err != nil {
		return err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return err
	}
	newToken, _ := data["access_token"].(string)
	if newToken == "" {
		return fmt.Errorf("refresh returned no access token")
	}
	creds.AccessToken = newToken
	if newRefresh, ok := data["refresh_token"].(string); ok && newRefresh != "" {
		creds.RefreshToken = newRefresh
	}
	return credentials.Save(creds)
}

// PromptLanguages resolves enabled languages from flags, or interactive selectors, falling back to DefaultLanguage.
func PromptLanguages(defaultLang string, secondary []string) ([]string, error) {
	if defaultLang != "" || len(secondary) > 0 {
		if defaultLang == "" {
			defaultLang = DefaultLanguage
		}
		return EnabledLanguages(defaultLang, secondary), nil
	}
	if !prompt.IsTTY() {
		return EnabledLanguages(DefaultLanguage, nil), nil
	}

	defaultChoice := DefaultLanguage
	defaultOptions := make([]huh.Option[string], 0, len(supportedLanguages))
	for _, l := range supportedLanguages {
		defaultOptions = append(defaultOptions, huh.NewOption(l.name, l.code))
	}
	if err := huh.NewSelect[string]().Title("Default language").Options(defaultOptions...).Value(&defaultChoice).Run(); err != nil {
		return nil, err
	}

	var secondaryChoices []string
	secondaryOptions := make([]huh.Option[string], 0, len(supportedLanguages))
	for _, l := range supportedLanguages {
		if l.code == defaultChoice {
			continue
		}
		secondaryOptions = append(secondaryOptions, huh.NewOption(l.name, l.code))
	}
	if err := huh.NewMultiSelect[string]().Title("Additional languages (optional)").Options(secondaryOptions...).Value(&secondaryChoices).Run(); err != nil {
		return nil, err
	}

	return EnabledLanguages(defaultChoice, secondaryChoices), nil
}

// EnabledLanguages returns the default language first, followed by distinct secondary languages.
func EnabledLanguages(defaultLang string, secondary []string) []string {
	langs := []string{defaultLang}
	seen := map[string]bool{defaultLang: true}
	for _, l := range secondary {
		l = strings.TrimSpace(l)
		if l != "" && !seen[l] {
			langs = append(langs, l)
			seen[l] = true
		}
	}
	return langs
}

// ResolveOrgID returns the organization to act on: the explicit id if given, the
// sole organization if there is exactly one, otherwise a prompt (TTY) or an error.
func ResolveOrgID(c *client.ClientWithResponses, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	resp, err := c.ListMyOrgsOrgsGetWithResponse(Ctx())
	if err != nil {
		return "", err
	}
	if err := CheckErr(resp.HTTPResponse, resp.Body); err != nil {
		return "", err
	}
	if resp.JSON200 == nil || len(*resp.JSON200) == 0 {
		return "", fmt.Errorf("no organization found for your account")
	}
	orgs := *resp.JSON200
	if len(orgs) == 1 {
		return orgs[0].Id, nil
	}
	if !prompt.IsTTY() {
		return "", fmt.Errorf("multiple organizations found; specify --org-id")
	}
	var options []huh.Option[string]
	for _, o := range orgs {
		options = append(options, huh.NewOption(o.Name, o.Id))
	}
	return prompt.AskSelect("Select an organization:", options)
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
