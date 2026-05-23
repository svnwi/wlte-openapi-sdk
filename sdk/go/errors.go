package wlteopenapi

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	Status     int
	Code       string
	Message    string
	Data       json.RawMessage
	RetryAfter string
}

func (e *APIError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("wlte api error: status=%d message=%s", e.Status, e.Message)
	}
	return fmt.Sprintf("wlte api error: status=%d code=%s message=%s", e.Status, e.Code, e.Message)
}

func isAuthExpired(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Status == 401 && apiErr.Code == "AUTH_EXPIRED"
}
