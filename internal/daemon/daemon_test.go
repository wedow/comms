package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wedow/comms/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		General: config.GeneralConfig{Format: "markdown"},
	}
}

func TestRunWritesPIDAndCleansUp(t *testing.T) {
	root := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so Run returns

	err := Run(ctx, testConfig(), root, []string{"telegram"})
	if err == nil || err != context.Canceled {
		t.Fatalf("Run returned %v, want context.Canceled", err)
	}

	// PID file should be removed after Run returns
	if _, err := os.Stat(filepath.Join(root, "daemon.pid")); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Run exits")
	}
}
