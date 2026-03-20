//go:build !windows

package daemon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/store"
)

func TestIntegrationSubprocessPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	root := t.TempDir()

	// Create a shell script that speaks the JSONL subprocess protocol.
	scriptPath := filepath.Join(t.TempDir(), "comms-testprov")
	script := `#!/bin/sh
echo '{"type":"ready","provider":"testprov","version":"1"}'
read -r cmd
echo '{"type":"message","offset":100,"id":1,"chat_id":555,"channel":"general","from":"alice","date":"2026-01-01T00:00:00Z","body":"integration test message"}'
sleep 0.2
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	// Save and restore all swappable vars.
	origLookPath := lookPathFunc
	origSpawn := spawnFunc
	origSleep := sleepFunc
	origMaxFail := respawnMaxFailures
	origBackoffCap := respawnBackoffCap
	t.Cleanup(func() {
		lookPathFunc = origLookPath
		spawnFunc = origSpawn
		sleepFunc = origSleep
		respawnMaxFailures = origMaxFail
		respawnBackoffCap = origBackoffCap
	})

	lookPathFunc = func(name string) (string, error) {
		return scriptPath, nil
	}
	spawnFunc = Spawn // use real Spawn — this is the integration test
	sleepFunc = func(_ context.Context, _ time.Duration) error { return nil }
	respawnMaxFailures = 2
	respawnBackoffCap = time.Millisecond

	cfg := config.Config{
		General:   config.GeneralConfig{Format: "markdown"},
		Providers: map[string]map[string]any{"testprov": {"token": "fake"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run blocks until the subprocess exits and respawn gives up.
	_ = Run(ctx, cfg, root, []string{"testprov"})

	// Verify message file was written.
	channelDir := filepath.Join(root, "testprov-general")
	entries, err := os.ReadDir(channelDir)
	if err != nil {
		t.Fatalf("channel dir not created: %v", err)
	}
	foundMsg := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			foundMsg = true
			break
		}
	}
	if !foundMsg {
		t.Error("expected .md message file in channel directory")
	}

	// Verify offset was persisted.
	offset, err := store.ReadOffset(root, "testprov")
	if err != nil {
		t.Fatalf("reading offset: %v", err)
	}
	if offset != 100 {
		t.Errorf("offset = %d, want 100", offset)
	}

	// PID file should be cleaned up after Run returns.
	if _, err := os.Stat(filepath.Join(root, "daemon.pid")); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Run exits")
	}
}
