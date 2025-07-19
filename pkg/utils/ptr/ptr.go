package ptr

import "strings"

// UpdateStringIfProvided checks if a source string pointer is not nil and holds a non-empty value.
// If it does, it updates the destination with the value from the source.
func UpdateStringIfProvided(dest *string, src *string) {
	if src != nil && strings.TrimSpace(*src) != "" {
		*dest = *src
	}
}