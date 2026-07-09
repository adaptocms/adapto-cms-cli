package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/test/mockapi"
)

var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "adapto-cli-integration")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	binPath = filepath.Join(dir, "adapto")

	build := exec.Command("go", "build", "-o", binPath, "../..")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building CLI: %v\n%s", err, out)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

type result struct {
	stdout   string
	stderr   string
	exitCode int
}

func run(t *testing.T, home string, extraEnv []string, args ...string) result {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = append([]string{"HOME=" + home, "PATH=" + os.Getenv("PATH")}, extraEnv...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := cmd.Run()
	exitCode := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("running %v: %v", args, err)
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), exitCode: exitCode}
}

func credsPath(home string) string {
	return filepath.Join(home, ".config", "adapto", "credentials.json")
}

func seedCredentials(t *testing.T, home, access, refresh, tenant string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(credsPath(home)), 0700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
		"tenant_id":     tenant,
	})
	if err := os.WriteFile(credsPath(home), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func loadCredentials(t *testing.T, home string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(credsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	var creds map[string]string
	if err := json.Unmarshal(data, &creds); err != nil {
		t.Fatal(err)
	}
	return creds
}

func apiEnv(api *mockapi.Server) []string {
	return []string{"ADAPTO_CLI_API_URL=" + api.URL}
}

func articlesFixture() client.PaginatedResponseArticleResponseModel {
	return client.PaginatedResponseArticleResponseModel{
		Total: 2,
		Page:  1,
		Pages: 1,
		Limit: 10,
		Items: []client.ArticleResponseModel{
			{Id: "a1", Title: "First Article", Status: "published", Language: "en", Slug: "first-article", Author: "Jane"},
			{Id: "a2", Title: "Second Article", Status: "draft", Language: "en", Slug: "second-article", Author: "John"},
		},
	}
}

func TestLoginSavesCredentialsAndAutoSelectsTenant(t *testing.T) {
	home := t.TempDir()
	api := mockapi.New()
	defer api.Close()

	api.Handle("POST", "/auth/login", 200, map[string]string{"access_token": "at1", "refresh_token": "rt1"})
	api.Handle("GET", "/orgs", 200, []client.Organization{{Id: "org1", Name: "Acme"}})
	api.Handle("GET", "/tenants/by-org/org1", 200, []client.Tenant{
		{Id: "t1", Name: "Main", OrganizationId: "org1", EnabledLanguages: []string{"en"}},
	})

	res := run(t, home, apiEnv(api), "auth", "login", "--email", "a@b.c", "--password", "pw")
	if res.exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %q", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "Logged in") {
		t.Fatalf("stdout = %q, want login confirmation", res.stdout)
	}
	if !strings.Contains(res.stdout, "t1") {
		t.Fatalf("stdout = %q, want auto-selected tenant t1", res.stdout)
	}

	var loginBody map[string]string
	logins := api.RequestsTo("POST", "/auth/login")
	if len(logins) != 1 {
		t.Fatalf("login requests = %d, want 1", len(logins))
	}
	if err := json.Unmarshal(logins[0].Body, &loginBody); err != nil {
		t.Fatal(err)
	}
	if loginBody["email"] != "a@b.c" || loginBody["password"] != "pw" {
		t.Fatalf("login body = %v", loginBody)
	}

	orgs := api.RequestsTo("GET", "/orgs")
	if len(orgs) != 1 || orgs[0].Token != "at1" {
		t.Fatalf("orgs requests = %+v, want one with Bearer at1", orgs)
	}

	creds := loadCredentials(t, home)
	want := map[string]string{"access_token": "at1", "refresh_token": "rt1", "tenant_id": "t1"}
	for k, v := range want {
		if creds[k] != v {
			t.Fatalf("credentials[%s] = %q, want %q (full: %v)", k, creds[k], v, creds)
		}
	}
}

func TestLoginAgainstWrongAPIIsSelfDiagnosing(t *testing.T) {
	home := t.TempDir()
	api := mockapi.New()
	defer api.Close()

	res := run(t, home, apiEnv(api), "auth", "login", "--email", "a@b.c", "--password", "pw")
	if res.exitCode == 0 {
		t.Fatal("expected nonzero exit")
	}
	if !strings.Contains(res.stderr, "Not found (POST "+api.URL+"/auth/login): Not Found") {
		t.Fatalf("stderr = %q, want method+URL in error", res.stderr)
	}
}

func TestExpiredTokenIsRefreshedAndRetried(t *testing.T) {
	home := t.TempDir()
	seedCredentials(t, home, "expired", "rt1", "t1")
	api := mockapi.New()
	defer api.Close()

	api.HandleFunc("GET", "/manage/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fresh" {
			mockapi.WriteJSON(w, 401, map[string]string{"detail": "token expired"})
			return
		}
		mockapi.WriteJSON(w, 200, articlesFixture())
	})
	api.Handle("POST", "/auth/refresh", 200, map[string]string{"access_token": "fresh", "refresh_token": "rt2"})

	res := run(t, home, apiEnv(api), "articles", "list")
	if res.exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %q", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "First Article") {
		t.Fatalf("stdout = %q, want article listing after refresh", res.stdout)
	}

	creds := loadCredentials(t, home)
	if creds["access_token"] != "fresh" || creds["refresh_token"] != "rt2" {
		t.Fatalf("credentials not rotated: %v", creds)
	}
}

func TestExpiredRefreshTokenFailsWithSessionExpired(t *testing.T) {
	home := t.TempDir()
	seedCredentials(t, home, "expired", "dead", "t1")
	api := mockapi.New()
	defer api.Close()

	api.Handle("GET", "/manage/articles", 401, map[string]string{"detail": "token expired"})
	api.Handle("POST", "/auth/refresh", 401, map[string]string{"detail": "refresh expired"})

	res := run(t, home, apiEnv(api), "articles", "list")
	if res.exitCode == 0 {
		t.Fatal("expected nonzero exit")
	}
	if !strings.Contains(res.stderr, "session expired") {
		t.Fatalf("stderr = %q, want session expired message", res.stderr)
	}
}

func TestArticlesListTableOutput(t *testing.T) {
	home := t.TempDir()
	seedCredentials(t, home, "at1", "rt1", "t1")
	api := mockapi.New()
	defer api.Close()

	api.Handle("GET", "/manage/articles", 200, articlesFixture())

	res := run(t, home, apiEnv(api), "articles", "list")
	if res.exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %q", res.exitCode, res.stderr)
	}
	for _, want := range []string{"Total: 2", "First Article", "Second Article", "published", "draft"} {
		if !strings.Contains(res.stdout, want) {
			t.Fatalf("stdout = %q, missing %q", res.stdout, want)
		}
	}

	reqs := api.RequestsTo("GET", "/manage/articles")
	if len(reqs) != 1 || reqs[0].Token != "at1" || reqs[0].TenantID != "t1" {
		t.Fatalf("requests = %+v, want one with Bearer at1 and X-Tenant-ID t1", reqs)
	}
}

func TestArticlesListJSONOutput(t *testing.T) {
	home := t.TempDir()
	seedCredentials(t, home, "at1", "rt1", "t1")
	api := mockapi.New()
	defer api.Close()

	api.Handle("GET", "/manage/articles", 200, articlesFixture())

	res := run(t, home, apiEnv(api), "articles", "list", "--json")
	if res.exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %q", res.exitCode, res.stderr)
	}

	var got client.PaginatedResponseArticleResponseModel
	if err := json.Unmarshal([]byte(res.stdout), &got); err != nil {
		t.Fatalf("--json output is not valid JSON: %v\n%s", err, res.stdout)
	}
	if got.Total != 2 || len(got.Items) != 2 || got.Items[0].Title != "First Article" {
		t.Fatalf("decoded = %+v", got)
	}
}

func TestCreateBatchSendsAllItemsInOneRequest(t *testing.T) {
	home := t.TempDir()
	seedCredentials(t, home, "at1", "rt1", "t1")
	api := mockapi.New()
	defer api.Close()

	batchPath := "/manage/custom-collections/col1/items/batch"
	api.Handle("POST", batchPath, 201, []any{})

	itemsJSON := `{"items":[
		{"title":"Jane Doe","slug":"jane-doe","language":"en","data":{"role":"Engineer"}},
		{"title":"John Doe","slug":"john-doe","language":"en","data":{"role":"Designer"}}
	]}`
	res := run(t, home, apiEnv(api), "collections", "items", "create-batch", "col1", "--items-json", itemsJSON)
	if res.exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %q", res.exitCode, res.stderr)
	}

	reqs := api.RequestsTo("POST", batchPath)
	if len(reqs) != 1 {
		t.Fatalf("batch requests = %d, want exactly 1", len(reqs))
	}
	var sent struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(reqs[0].Body, &sent); err != nil {
		t.Fatal(err)
	}
	if len(sent.Items) != 2 || sent.Items[0]["title"] != "Jane Doe" || sent.Items[1]["title"] != "John Doe" {
		t.Fatalf("sent body = %s", reqs[0].Body)
	}
	if reqs[0].TenantID != "t1" {
		t.Fatalf("X-Tenant-ID = %q, want t1", reqs[0].TenantID)
	}
}

func TestNonInteractiveMissingArgFails(t *testing.T) {
	home := t.TempDir()
	api := mockapi.New()
	defer api.Close()

	res := run(t, home, apiEnv(api), "auth", "login")
	if res.exitCode == 0 {
		t.Fatal("expected nonzero exit")
	}
	if !strings.Contains(res.stderr, "required: --email") {
		t.Fatalf("stderr = %q, want required --email error", res.stderr)
	}
	if len(api.Requests()) != 0 {
		t.Fatalf("expected no API requests, got %+v", api.Requests())
	}
}

func TestLegacyEnvVarsWarnAndAreIgnored(t *testing.T) {
	home := t.TempDir()

	res := run(t, home, []string{"ADAPTO_API_URL=https://public-api.adaptocms.com/v1"}, "version")
	if res.exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %q", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "ADAPTO_API_URL is ignored") {
		t.Fatalf("stderr = %q, want legacy var warning", res.stderr)
	}

	res = run(t, home, []string{
		"ADAPTO_API_URL=https://public-api.adaptocms.com/v1",
		"ADAPTO_CLI_API_URL=https://api.adaptocms.com",
	}, "version")
	if strings.Contains(res.stderr, "ADAPTO_API_URL") {
		t.Fatalf("stderr = %q, want no warning when new var is set", res.stderr)
	}
}

func TestLegacyAPIURLDoesNotRedirectRequests(t *testing.T) {
	home := t.TempDir()
	seedCredentials(t, home, "at1", "rt1", "t1")
	api := mockapi.New()
	defer api.Close()

	api.Handle("GET", "/manage/articles", 200, articlesFixture())

	res := run(t, home, []string{
		"ADAPTO_CLI_API_URL=" + api.URL,
		"ADAPTO_API_URL=https://public-api.adaptocms.com/v1",
	}, "articles", "list")
	if res.exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %q", res.exitCode, res.stderr)
	}
	if len(api.RequestsTo("GET", "/manage/articles")) != 1 {
		t.Fatal("expected request to hit ADAPTO_CLI_API_URL server")
	}
}
