package cli

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// Swappable for testing.
var (
	lookPath           = exec.LookPath
	runDelegate        = defaultRunDelegate
	runDelegateOutput  = defaultRunDelegateOutput
)

func defaultRunDelegate(binary string, args []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultRunDelegateOutput(binary string, args []string, env []string, stdin io.Reader) ([]byte, error) {
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = stdin
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// extractProvider returns the provider name from a channel string.
// "telegram-general" → "telegram"
func extractProvider(channel string) string {
	if i := strings.Index(channel, "-"); i >= 0 {
		return channel[:i]
	}
	return channel
}

// resolveProviderBinary finds the provider binary in PATH.
func resolveProviderBinary(provider string) (string, error) {
	return lookPath("comms-" + provider)
}

// delegate runs a provider binary with the given args, inheriting stdio.
func delegate(binary string, args []string) error {
	return runDelegate(binary, args)
}

// delegateWithOutput runs a provider binary, captures stdout, and returns it.
func delegateWithOutput(binary string, args []string, env []string, stdin io.Reader) ([]byte, error) {
	return runDelegateOutput(binary, args, env, stdin)
}
