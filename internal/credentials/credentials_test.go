package credentials_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
)

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	creds, err := credentials.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if creds.AccessToken != "" || creds.RefreshToken != "" || creds.TenantID != "" {
		t.Fatalf("expected empty credentials, got %+v", creds)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	in := &credentials.Credentials{AccessToken: "a1", RefreshToken: "r1", TenantID: "t1"}
	if err := credentials.Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	out, err := credentials.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *out != *in {
		t.Fatalf("got %+v, want %+v", out, in)
	}

	info, err := os.Stat(credentials.Path())
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("credentials file perm = %o, want 0600", perm)
	}
	if got, want := credentials.Path(), filepath.Join(os.Getenv("HOME"), ".config", "adapto", "credentials.json"); got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}

func TestClear(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := credentials.Clear(); err != nil {
		t.Fatalf("Clear on missing file: %v", err)
	}

	if err := credentials.Save(&credentials.Credentials{AccessToken: "a"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := credentials.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if _, err := os.Stat(credentials.Path()); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, stat err = %v", err)
	}
}

func TestLoadCorruptFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	p := credentials.Path()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := credentials.Load(); err == nil {
		t.Fatal("expected error for corrupt file")
	}
}
