// Package utils provides utility functions for the EMS application.
package utils //nolint:revive // utils is a common and acceptable package name

import "time"

// GetUTCString formats a time.Time to the ENTSO-E API format YYYYMMDDHHmm.
func GetUTCString(t time.Time) string {
	return t.UTC().Format("200601021504")
}
