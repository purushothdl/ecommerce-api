// pkg/utils/jsonutil/jsonutil.go
package jsonutil

import (
	"encoding/json"
	"log/slog"
)

// MustMarshal is a helper function that marshals a value to JSON.
// It panics if the marshalling fails. This should only be used for types
// that are known to be valid for marshalling, where an error would
// indicate a severe programmer error.
func MustMarshal(v any) []byte {
	bytes, err := json.Marshal(v)
	if err != nil {
		// A panic is appropriate here because if a struct we defined
		// cannot be marshalled, it's a non-recoverable programmer error.
		slog.Error("CRITICAL: Failed to marshal known-good type", "error", err)
		panic("jsonutil: failed to marshal: " + err.Error())
	}
	return bytes
}