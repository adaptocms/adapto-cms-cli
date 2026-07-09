package errors_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	apierrors "github.com/adaptocms/adapto-cms-cli/internal/errors"
)

func respWith(status int, method, rawURL string) *http.Response {
	u, _ := url.Parse(rawURL)
	return &http.Response{
		StatusCode: status,
		Request:    &http.Request{Method: method, URL: u},
	}
}

func TestCheckResponseSuccess(t *testing.T) {
	if err := apierrors.CheckResponse(nil, nil); err != nil {
		t.Fatalf("nil response: got %v", err)
	}
	if err := apierrors.CheckResponse(respWith(200, "GET", "https://api.example.com/x"), nil); err != nil {
		t.Fatalf("200: got %v", err)
	}
	if err := apierrors.CheckResponse(respWith(204, "DELETE", "https://api.example.com/x"), nil); err != nil {
		t.Fatalf("204: got %v", err)
	}
}

func TestCheckResponseIncludesMethodAndURL(t *testing.T) {
	err := apierrors.CheckResponse(
		respWith(404, "POST", "https://public-api.adaptocms.com/v1/auth/login"),
		[]byte(`{"detail":"Not Found"}`),
	)
	if err == nil {
		t.Fatal("expected error")
	}
	want := "Not found (POST https://public-api.adaptocms.com/v1/auth/login): Not Found"
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestCheckResponseWithoutRequestInfo(t *testing.T) {
	err := apierrors.CheckResponse(&http.Response{StatusCode: 404}, []byte(`{"detail":"gone"}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "Not found: gone"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCheckResponseValidationDetails(t *testing.T) {
	body := []byte(`{"detail":[{"loc":["body","email"],"msg":"value is not a valid email address"}]}`)
	err := apierrors.CheckResponse(respWith(422, "POST", "https://api.adaptocms.com/auth/register"), body)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "Validation error (POST https://api.adaptocms.com/auth/register)") {
		t.Fatalf("unexpected prefix: %q", msg)
	}
	if !strings.Contains(msg, "value is not a valid email address") {
		t.Fatalf("missing detail line: %q", msg)
	}
}

func TestCheckResponseStatusMessages(t *testing.T) {
	cases := map[int]string{
		400: "Bad request",
		401: "Unauthorized",
		403: "Forbidden",
		409: "Conflict",
		429: "Rate limited",
		500: "Internal server error",
		418: "HTTP 418 error",
	}
	for status, want := range cases {
		err := apierrors.CheckResponse(&http.Response{StatusCode: status}, nil)
		if err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Errorf("status %d: got %v, want prefix %q", status, err, want)
		}
	}
}
