package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// PrintJSON marshals v as JSON and writes one newline-terminated line to w.
func PrintJSON(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// PrintError writes a JSON error object to stderr.
func PrintError(msg string, args ...any) {
	data, _ := json.Marshal(map[string]string{"error": fmt.Sprintf(msg, args...)})
	fmt.Fprintln(os.Stderr, string(data))
}
