package config

import (
	"github.com/eggnita/adapto_cms_cli/internal/credentials"
	"github.com/spf13/viper"
)

// Config holds the resolved CLI configuration.
type Config struct {
	APIURL   string
	Token    string
	TenantID string
	JSON     bool
	Verbose  bool
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
			if cfg.Token == "" {
				cfg.Token = creds.AccessToken
			}
			if cfg.TenantID == "" {
				cfg.TenantID = creds.TenantID
			}
		}
	}
	return cfg
}
