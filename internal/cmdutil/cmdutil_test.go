package cmdutil_test

import (
	"reflect"
	"testing"

	"github.com/adaptocms/adapto-cms-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func TestStringPtr(t *testing.T) {
	if cmdutil.StringPtr("") != nil {
		t.Fatal("empty string should map to nil")
	}
	if got := cmdutil.StringPtr("x"); got == nil || *got != "x" {
		t.Fatalf("got %v, want pointer to x", got)
	}
}

func TestIntPtr(t *testing.T) {
	if cmdutil.IntPtr(0) != nil {
		t.Fatal("zero should map to nil")
	}
	if got := cmdutil.IntPtr(7); got == nil || *got != 7 {
		t.Fatalf("got %v, want pointer to 7", got)
	}
}

func TestStringSlicePtr(t *testing.T) {
	if cmdutil.StringSlicePtr("") != nil {
		t.Fatal("empty string should map to nil")
	}
	got := cmdutil.StringSlicePtr("a, b ,c")
	want := []string{"a", "b", "c"}
	if got == nil || !reflect.DeepEqual(*got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func customFieldsCmd(value string) *cobra.Command {
	cmd := &cobra.Command{}
	cmdutil.AddCustomFieldsFlag(cmd)
	if value != "" {
		_ = cmd.Flags().Set("custom-fields-json", value)
	}
	return cmd
}

func TestParseCustomFieldsEmpty(t *testing.T) {
	cf, err := cmdutil.ParseCustomFields(customFieldsCmd(""))
	if err != nil || cf != nil {
		t.Fatalf("got %v, %v; want nil, nil", cf, err)
	}
}

func TestParseCustomFieldsValid(t *testing.T) {
	cf, err := cmdutil.ParseCustomFields(customFieldsCmd(`{"seo_title":{"type":"text","value":"Welcome"}}`))
	if err != nil {
		t.Fatalf("ParseCustomFields: %v", err)
	}
	field, ok := (*cf)["seo_title"]
	if !ok || field.Type != "text" {
		t.Fatalf("got %+v, want seo_title text field", cf)
	}
}

func TestParseCustomFieldsRejectsUnknownKeys(t *testing.T) {
	if _, err := cmdutil.ParseCustomFields(customFieldsCmd(`{"f":{"type":"text","bogus":true}}`)); err == nil {
		t.Fatal("expected error for unknown field key")
	}
}

func TestParseCustomFieldsInvalidJSON(t *testing.T) {
	if _, err := cmdutil.ParseCustomFields(customFieldsCmd(`{not json`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
