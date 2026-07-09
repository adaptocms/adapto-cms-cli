package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CheckResponse returns a user-friendly error for non-2xx responses, or nil.
func CheckResponse(resp *http.Response, body []byte) error {
	if resp == nil || (resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil
	}

	msg := friendlyMessage(resp.StatusCode)
	if resp.Request != nil && resp.Request.URL != nil {
		msg += fmt.Sprintf(" (%s %s)", resp.Request.Method, resp.Request.URL)
	}

	var detail struct {
		Detail interface{} `json:"detail"`
	}
	if err := json.Unmarshal(body, &detail); err == nil && detail.Detail != nil {
		switch d := detail.Detail.(type) {
		case string:
			msg += ": " + d
		case []interface{}:
			for _, item := range d {
				if m, ok := item.(map[string]interface{}); ok {
					if loc, ok := m["loc"]; ok {
						msg += fmt.Sprintf("\n  - %v: %v", loc, m["msg"])
					}
				}
			}
		}
	}

	return fmt.Errorf("%s", msg)
}

func friendlyMessage(code int) string {
	switch code {
	case 400:
		return "Bad request"
	case 401:
		return "Unauthorized - check your token (ADAPTO_CLI_TOKEN or --token)"
	case 403:
		return "Forbidden - insufficient permissions"
	case 404:
		return "Missing"
	case 409:
		return "Conflict - resource already exists"
	case 422:
		return "Validation error"
	case 429:
		return "Rate limited - try again later"
	case 500:
		return "Internal server error"
	default:
		return fmt.Sprintf("HTTP %d error", code)
	}
}
