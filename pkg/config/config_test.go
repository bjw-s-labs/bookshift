package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfigLoad verifies that a valid config file is loaded with the expected
// fields and defaults, exercising strict YAML parsing and validation.
func TestConfigLoad(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	data := "" +
		"log_level: debug\n" +
		"target_folder: /tmp/books\n" +
		"overwrite_existing_files: true\n" +
		"valid_extensions: ['.epub']\n" +
		"sources: []\n"
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var c Config
	if err := c.Load(cfgPath); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.TargetFolder != "/tmp/books" || len(c.ValidExtensions) != 1 {
		t.Fatalf("unexpected config: %+v", c)
	}
}

// TestConfigLoad_FileNotFound ensures Load returns an error for a missing file.
func TestConfigLoad_FileNotFound(t *testing.T) {
	var c Config
	if err := c.Load("/definitely/not/here.yaml"); err == nil {
		t.Fatalf("expected error for missing file")
	}
}

// TestConfigLoad_ValidationMissingFields checks validator rejects configs
// missing required fields like target_folder and valid_extensions.
func TestConfigLoad_ValidationMissingFields(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.yaml")
	// Missing target_folder and valid_extensions
	data := "sources: []\n"
	if err := os.WriteFile(p, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := c.Load(p); err == nil {
		t.Fatalf("expected validation error")
	}
}

// TestConfigLoad_StrictUnknownField ensures unknown top-level YAML keys
// cause strict mode unmarshalling to fail.
func TestConfigLoad_StrictUnknownField(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.yaml")
	// Include an unknown top-level key to trigger yaml.Strict error
	data := "log_level: info\nunknown: true\ntarget_folder: /tmp\nvalid_extensions: ['.epub']\nsources: []\n"
	if err := os.WriteFile(p, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := c.Load(p); err == nil {
		t.Fatalf("expected strict mode error for unknown field")
	}
}
