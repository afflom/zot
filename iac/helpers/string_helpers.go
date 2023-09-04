// helpers/string_helpers.go
// This file contains helper functions for string operations.

package helpers

// IfEmpty returns defaultVal if val is an empty string.
func IfEmpty(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
