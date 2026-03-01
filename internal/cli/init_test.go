package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/wedow/comms/internal/config"
)

func TestInitCreatesDirectoryTree(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".comms")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// root dir must exist
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		t.Errorf("root dir %s not created", root)
	}
	// docs subdir must exist
	docsDir := filepath.Join(root, "docs")
	if info, err := os.Stat(docsDir); err != nil || !info.IsDir() {
		t.Errorf("docs dir %s not created", docsDir)
	}
}

func TestInitWritesDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".comms")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfgPath := filepath.Join(root, "config.toml")
	var got config.Config
	if _, err := toml.DecodeFile(cfgPath, &got); err != nil {
		t.Fatalf("failed to decode config.toml: %v", err)
	}

	want := config.Default()
	if got.General.Format != want.General.Format {
		t.Errorf("format = %q, want %q", got.General.Format, want.General.Format)
	}
	if got.Callback.Delay != want.Callback.Delay {
		t.Errorf("delay = %q, want %q", got.Callback.Delay, want.Callback.Delay)
	}
}

func TestInitDeploysTelegramDoc(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".comms")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	docPath := filepath.Join(root, "docs", "telegram-setup.md")
	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("telegram-setup.md not found: %v", err)
	}
	if len(data) == 0 {
		t.Error("telegram-setup.md is empty")
	}
}

func TestInitDoesNotOverwriteExistingConfig(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".comms")

	// Run init once
	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// Modify config.toml
	cfgPath := filepath.Join(root, "config.toml")
	custom := []byte("# custom config\n")
	if err := os.WriteFile(cfgPath, custom, 0o644); err != nil {
		t.Fatalf("failed to write custom config: %v", err)
	}

	// Run init again
	cmd2 := newRootCmd()
	cmd2.SetArgs([]string{"init", "--dir", root})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("second init failed: %v", err)
	}

	// Config must be unchanged
	got, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if string(got) != string(custom) {
		t.Errorf("config was overwritten: got %q, want %q", got, custom)
	}
}

func TestInitOutputsJSON(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".comms")

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"init", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["status"] != "initialized" {
		t.Errorf("status = %q, want %q", result["status"], "initialized")
	}
	if result["path"] == "" {
		t.Error("path is empty")
	}
	// path should be absolute
	if !filepath.IsAbs(result["path"]) {
		t.Errorf("path %q is not absolute", result["path"])
	}
}
