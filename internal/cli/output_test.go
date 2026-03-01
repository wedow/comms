package cli

import (
	"bytes"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	err := PrintJSON(buf, map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}

	got := buf.String()
	want := `{"name":"test"}` + "\n"
	if got != want {
		t.Errorf("PrintJSON = %q, want %q", got, want)
	}
}

func TestPrintJSONNewlineTerminated(t *testing.T) {
	buf := new(bytes.Buffer)
	_ = PrintJSON(buf, map[string]int{"count": 42})

	got := buf.String()
	if got[len(got)-1] != '\n' {
		t.Error("PrintJSON output not newline-terminated")
	}
}
