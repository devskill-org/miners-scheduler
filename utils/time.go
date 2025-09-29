package utils

import "time"

// GetUTC formats a time.Time to the ENTSO-E API format YYYYMMDDHHmm
func GetUTCString(t time.Time) string {
	return t.UTC().Format("200601021504")
}
