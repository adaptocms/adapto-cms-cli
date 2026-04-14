package errors

import (
	"encoding/json"
	"fmt"
)

// CheckHTTP returns a user-friendly error for non-2xx status codes, or nil.
func CheckHTTP(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	msg := friendlyMessage(statusCode)

	// Try to extract detail from JSON body
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
		return "Unauthorized - check your token (ADAPTO_TOKEN or --token)"
	case 403:
		return "Forbidden - insufficient permissions"
	case 404:
		return "Not found"
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
