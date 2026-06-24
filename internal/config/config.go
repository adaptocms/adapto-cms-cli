package config

import (
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
	"github.com/spf13/viper"
)

// Config holds the resolved CLI configuration.
type Config struct {
	APIURL   string
	Token    string
	TenantID string
	JSON     bool
	Verbose  bool

	// TokenFromCreds is true only when Token came from the credentials file (not --token/env), gating auto-refresh.
	TokenFromCreds bool
}

// Load returns the current configuration from viper (env vars + flags).
// Falls back to the credentials file for token and tenant_id.
func Load() Config {
	cfg := Config{
		APIURL:   viper.GetString("api_url"),
		Token:    viper.GetString("token"),
		TenantID: viper.GetString("tenant_id"),
		JSON:     viper.GetBool("json"),
		Verbose:  viper.GetBool("verbose"),
	}
	if cfg.Token == "" || cfg.TenantID == "" {
		if creds, err := credentials.Load(); err == nil {
			if cfg.Token == "" && creds.AccessToken != "" {
				cfg.Token = creds.AccessToken
				cfg.TokenFromCreds = true
			}
			if cfg.TenantID == "" {
				cfg.TenantID = creds.TenantID
			}
		}
	}
	return cfg
}
