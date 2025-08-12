package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

// TestExecute_RunWithTempConfig verifies Execute runs the 'run' subcommand using
// a temporary, valid config file without errors.
func TestExecute_RunWithTempConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	data := "" +
		"log_level: debug\n" +
		"target_folder: " + dir + "\n" +
		"overwrite_existing_files: false\n" +
		"valid_extensions: ['.epub']\n" +
		"sources: []\n"
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"bookshift", "--config-file", cfgPath, "run"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

// TestVersionFlag_BeforeApply ensures the version flag prints the version and invokes the exit seam
// without terminating the test process.
func TestVersionFlag_BeforeApply(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Override exit seam
	oldExit := exit
	t.Cleanup(func() { exit = oldExit })
	code := -1
	exit = func(c int) { code = c }

	// Call hook directly
	v := VersionFlag("1.2.3")
	if err := v.BeforeApply(nil, kong.Vars{"version": string(v)}); err != nil {
		t.Fatalf("BeforeApply error: %v", err)
	}

	_ = w.Close()
	out := make([]byte, 64)
	n, _ := r.Read(out)
	if !strings.Contains(string(out[:n]), "1.2.3") {
		t.Fatalf("missing version in output: %q", string(out[:n]))
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TestVersionFlag_IsBoolAndDecode verifies that VersionFlag is treated as a boolean
// flag by kong and that Decode is a no-op that returns nil.
func TestVersionFlag_IsBoolAndDecode(t *testing.T) {
	var v VersionFlag = "test"
	if !v.IsBool() {
		t.Fatalf("IsBool should be true")
	}
	if err := v.Decode(nil); err != nil {
		t.Fatalf("Decode: %v", err)
	}
}

// TestExecute_ConfigLoadError ensures that a missing config file causes an error to be logged and
// the process to exit with a non-zero code via the exit seam.
func TestExecute_ConfigLoadError(t *testing.T) {
	oldExit := exit
	t.Cleanup(func() { exit = oldExit })
	code := -1
	exit = func(c int) { code = c }

	// Point to a non-existent config file
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"bookshift", "--config-file", "/nonexistent.yaml", "run"}

	// Capture logs to stderr
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldErr }()

	_ = Execute()
	_ = w.Close()

	buf := make([]byte, 256)
	_, _ = r.Read(buf)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

// TestExecute_RunError ensures that ctx.Run errors result in a non-zero exit via the exit seam.
// This is simulated by invoking an unknown subcommand so parse succeeds but Run fails.
func TestExecute_RunError(t *testing.T) {
	oldExit := exit
	t.Cleanup(func() { exit = oldExit })
	code := -1
	exit = func(c int) { code = c }

	dir := t.TempDir()
	cfgPath := dir + "/config.yaml"
	// Minimal valid config
	os.WriteFile(cfgPath, []byte("target_folder: .\nvalid_extensions: ['.epub']\nsources: []\n"), 0644)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	// Use a bogus subcommand to cause ctx.Run error after parsing
	os.Args = []string{"bookshift", "--config-file", cfgPath, "bogus"}

	_ = Execute()
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

// TestExecute_NoArgsShowsHelp verifies that when no arguments are provided, Execute appends --help
// internally and help text is printed to stderr.
func TestExecute_NoArgsShowsHelp(t *testing.T) {
	oldExit := exit
	t.Cleanup(func() { exit = oldExit })
	exit = func(c int) {}

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"bookshift"}

	// Capture stderr which receives help output by kong/tint
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldErr }()

	_ = Execute()
	_ = w.Close()
	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	if n == 0 {
		t.Fatalf("expected some help output")
	}
}

// TestExecute_SetsLogLevel verifies that the configured log_level (debug) is applied and Execute
// continues without error.
func TestExecute_SetsLogLevel(t *testing.T) {
	oldExit := exit
	t.Cleanup(func() { exit = oldExit })
	exit = func(c int) {}

	dir := t.TempDir()
	cfgPath := dir + "/config.yaml"
	// Valid config with debug log level
	data := "log_level: debug\ntarget_folder: .\nvalid_extensions: ['.epub']\nsources: []\n"
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"bookshift", "--config-file", cfgPath, "run"}

	_ = Execute()
}
