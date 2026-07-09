package config_test

import (
	"testing"

	"github.com/adaptocms/adapto-cms-cli/internal/config"
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
	"github.com/spf13/viper"
)

func TestLoadFallsBackToCredentialsFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	viper.Reset()
	t.Cleanup(viper.Reset)

	if err := credentials.Save(&credentials.Credentials{AccessToken: "stored-token", TenantID: "stored-tenant"}); err != nil {
		t.Fatal(err)
	}

	cfg := config.Load()
	if cfg.Token != "stored-token" {
		t.Fatalf("Token = %q, want stored-token", cfg.Token)
	}
	if !cfg.TokenFromCreds {
		t.Fatal("TokenFromCreds = false, want true for credentials-file token")
	}
	if cfg.TenantID != "stored-tenant" {
		t.Fatalf("TenantID = %q, want stored-tenant", cfg.TenantID)
	}
}

func TestLoadExplicitTokenWinsAndDisablesRefresh(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	viper.Reset()
	t.Cleanup(viper.Reset)

	if err := credentials.Save(&credentials.Credentials{AccessToken: "stored-token", TenantID: "stored-tenant"}); err != nil {
		t.Fatal(err)
	}
	viper.Set("token", "explicit-token")

	cfg := config.Load()
	if cfg.Token != "explicit-token" {
		t.Fatalf("Token = %q, want explicit-token", cfg.Token)
	}
	if cfg.TokenFromCreds {
		t.Fatal("TokenFromCreds = true, want false for explicit token")
	}
	if cfg.TenantID != "stored-tenant" {
		t.Fatalf("TenantID = %q, want stored-tenant (credentials fallback)", cfg.TenantID)
	}
}

func TestLoadWithoutAnySource(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	viper.Reset()
	t.Cleanup(viper.Reset)

	cfg := config.Load()
	if cfg.Token != "" || cfg.TenantID != "" || cfg.TokenFromCreds {
		t.Fatalf("expected empty config, got %+v", cfg)
	}
}
