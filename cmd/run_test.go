package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

// TestRunCommand_Run_NoSources verifies that Run completes successfully when no sources are configured
// and simply counts existing files without attempting any sync operations.
func TestRunCommand_Run_NoSources(t *testing.T) {
	dir := t.TempDir()
	// create one file then ensure no sources runs cleanly
	if err := os.WriteFile(filepath.Join(dir, "a.epub"), []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := &config.Config{TargetFolder: dir, ValidExtensions: []string{".epub"}}
	var logger = slog.Default()

	var c RunCommand
	if err := c.Run(cfg, logger); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// TestRunCommand_CountFilesErrorStart ensures that an error returned by the initial
// countFiles call is propagated from Run.
func TestRunCommand_CountFilesErrorStart(t *testing.T) {
	old := countFiles
	t.Cleanup(func() { countFiles = old })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) {
		return 0, errors.New("boom")
	}

	var c RunCommand
	cfg := &config.Config{TargetFolder: t.TempDir(), ValidExtensions: []string{".epub"}}
	if err := c.Run(cfg, slog.Default()); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunCommand_CountFilesErrorEnd ensures that an error returned by the final
// countFiles call is propagated from Run.
func TestRunCommand_CountFilesErrorEnd(t *testing.T) {
	calls := 0
	old := countFiles
	t.Cleanup(func() { countFiles = old })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) {
		calls++
		if calls == 1 {
			return 1, nil
		}
		return 0, errors.New("boom")
	}

	// prevent any source runs
	var c RunCommand
	cfg := &config.Config{TargetFolder: t.TempDir(), ValidExtensions: []string{".epub"}}
	if err := c.Run(cfg, slog.Default()); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunCommand_IncreaseButNoKobo verifies that when the file count increases but
// no Kobo device is detected, Run completes without attempting a library update.
func TestRunCommand_IncreaseButNoKobo(t *testing.T) {
	calls := 0
	old := countFiles
	t.Cleanup(func() { countFiles = old })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) {
		calls++
		if calls == 1 {
			return 0, nil
		}
		return 1, nil
	}

	oldIsKobo, oldUpdate := isKoboDevice, updateKoboLibrary
	t.Cleanup(func() { isKoboDevice, updateKoboLibrary = oldIsKobo, oldUpdate })
	isKoboDevice = func() bool { return false }
	updateKoboLibrary = func() error { t.Fatalf("should not be called"); return nil }

	var c RunCommand
	cfg := &config.Config{TargetFolder: t.TempDir(), ValidExtensions: []string{".epub"}}
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test that each source type is dispatched and errors don't abort the run.
func TestRunCommand_Sources_DispatchAndErrors(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TargetFolder:           dir,
		ValidExtensions:        []string{".epub"},
		OverwriteExistingFiles: false,
		Sources: []config.Source{
			{Type: "nfs", Config: &config.NfsNetworkShareConfig{}},
			{Type: "smb", Config: &config.SmbNetworkShareConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
		},
	}

	// Start with one file, end with two to trigger Kobo update path once.
	// We'll control the count via the countFiles seam.
	calls := 0
	oldCount := countFiles
	t.Cleanup(func() { countFiles = oldCount })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) {
		calls++
		if calls == 1 {
			return 1, nil
		}
		return 2, nil
	}

	// Make nfs succeed, smb fail, imap succeed.
	oldNfs, oldSmb, oldImap := doNfs, doSmb, doImap
	t.Cleanup(func() { doNfs, doSmb, doImap = oldNfs, oldSmb, oldImap })
	doNfs = func(_ context.Context, _ *config.NfsNetworkShareConfig, _ string, _ []string, _ bool) error {
		return nil
	}
	doSmb = func(_ context.Context, _ *config.SmbNetworkShareConfig, _ string, _ []string, _ bool) error {
		return errors.New("boom")
	}
	doImap = func(_ context.Context, _ *config.ImapConfig, _ string, _ []string, _ bool) error { return nil }

	// Trigger Kobo update path but stub it to a fast no-op.
	oldIsKobo, oldUpdate := isKoboDevice, updateKoboLibrary
	t.Cleanup(func() { isKoboDevice, updateKoboLibrary = oldIsKobo, oldUpdate })
	isKoboDevice = func() bool { return true }
	updated := false
	updateKoboLibrary = func() error { updated = true; return nil }

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !updated {
		t.Fatalf("expected Kobo update to be called")
	}

	// Ensure the seams were exercised twice for countFiles (start/end)
	if calls != 2 {
		t.Fatalf("expected 2 calls to countFiles, got %d", calls)
	}
}

// Test that when file count doesn't increase, Kobo update is skipped.
func TestRunCommand_NoIncrease_NoKobo(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{TargetFolder: dir, ValidExtensions: []string{".epub"}}

	oldCount := countFiles
	t.Cleanup(func() { countFiles = oldCount })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) {
		// same before and after
		return 1, nil
	}

	oldIsKobo, oldUpdate := isKoboDevice, updateKoboLibrary
	t.Cleanup(func() { isKoboDevice, updateKoboLibrary = oldIsKobo, oldUpdate })
	isKoboDevice = func() bool { t.Fatalf("should not be called"); return false }
	updateKoboLibrary = func() error { t.Fatalf("should not be called"); return nil }

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// Cover path that valid extensions are respected when counting (via args only) and that
// the target path is passed through; this is largely smoke coverage for seams.
func TestRunCommand_CountFilesArgs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "book.epub")
	_ = file // not creating the file as we stub countFiles
	cfg := &config.Config{TargetFolder: dir, ValidExtensions: []string{".epub"}}

	// Verify the inputs our seam receives.
	old := countFiles
	t.Cleanup(func() { countFiles = old })
	called := false
	countFiles = func(folder string, exts []string, recursive bool) (int, error) {
		called = true
		if folder != dir || len(exts) != 1 || exts[0] != ".epub" || !recursive {
			t.Fatalf("unexpected args: %v %v %v", folder, exts, recursive)
		}
		return 0, nil
	}

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Fatalf("countFiles not called")
	}
}

// Test that per-source timeout sets a deadline and the worker observes ctx.Done.
func TestRunCommand_PerSourceTimeout(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TargetFolder:    dir,
		ValidExtensions: []string{".epub"},
		Sources: []config.Source{
			{Type: "nfs", Config: &config.NfsNetworkShareConfig{TimeoutSeconds: 1}},
		},
	}

	oldCount := countFiles
	t.Cleanup(func() { countFiles = oldCount })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) { return 0, nil }

	var deadlineSet atomic.Bool
	oldNfs := doNfs
	t.Cleanup(func() { doNfs = oldNfs })
	doNfs = func(ctx context.Context, _ *config.NfsNetworkShareConfig, _ string, _ []string, _ bool) error {
		if dl, ok := ctx.Deadline(); ok && time.Until(dl) > 0 {
			deadlineSet.Store(true)
		}
		<-ctx.Done()
		return nil
	}

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !deadlineSet.Load() {
		t.Fatalf("expected per-source deadline to be set")
	}
}

// Test that concurrency is bounded by cfg.Concurrency using a counting seam.
func TestRunCommand_ConcurrencyBound(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TargetFolder:    dir,
		ValidExtensions: []string{".epub"},
		Concurrency:     2,
		Sources: []config.Source{
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
		},
	}

	oldCount := countFiles
	t.Cleanup(func() { countFiles = oldCount })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) { return 0, nil }

	var active int32
	var maxObserved int32
	oldIm := doImap
	t.Cleanup(func() { doImap = oldIm })
	doImap = func(ctx context.Context, _ *config.ImapConfig, _ string, _ []string, _ bool) error {
		n := atomic.AddInt32(&active, 1)
		// track maximum
		for {
			m := atomic.LoadInt32(&maxObserved)
			if n > m {
				if atomic.CompareAndSwapInt32(&maxObserved, m, n) {
					break
				}
				continue
			}
			break
		}
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&active, -1)
		return nil
	}

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if maxObserved > 2 {
		t.Fatalf("expected max concurrency <= 2, got %d", maxObserved)
	}
}

// Test that when concurrency is not set, default bound (3) is enforced.
func TestRunCommand_ConcurrencyDefaultBound(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TargetFolder:    dir,
		ValidExtensions: []string{".epub"},
		// Concurrency omitted -> default 3
		Sources: []config.Source{
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
		},
	}

	oldCount := countFiles
	t.Cleanup(func() { countFiles = oldCount })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) { return 0, nil }

	var active int32
	var maxObserved int32
	oldIm := doImap
	t.Cleanup(func() { doImap = oldIm })
	doImap = func(ctx context.Context, _ *config.ImapConfig, _ string, _ []string, _ bool) error {
		n := atomic.AddInt32(&active, 1)
		for {
			m := atomic.LoadInt32(&maxObserved)
			if n > m {
				if atomic.CompareAndSwapInt32(&maxObserved, m, n) {
					break
				}
				continue
			}
			break
		}
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&active, -1)
		return nil
	}

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if maxObserved > 3 {
		t.Fatalf("expected max concurrency <= 3, got %d", maxObserved)
	}
}

// Test that when concurrency exceeds number of sources, it is capped to len(sources).
func TestRunCommand_ConcurrencyCappedToSources(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TargetFolder:    dir,
		ValidExtensions: []string{".epub"},
		Concurrency:     5,
		Sources: []config.Source{
			{Type: "imap", Config: &config.ImapConfig{}},
			{Type: "imap", Config: &config.ImapConfig{}},
		},
	}

	oldCount := countFiles
	t.Cleanup(func() { countFiles = oldCount })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) { return 0, nil }

	var active int32
	var maxObserved int32
	oldIm := doImap
	t.Cleanup(func() { doImap = oldIm })
	doImap = func(ctx context.Context, _ *config.ImapConfig, _ string, _ []string, _ bool) error {
		n := atomic.AddInt32(&active, 1)
		for {
			m := atomic.LoadInt32(&maxObserved)
			if n > m {
				if atomic.CompareAndSwapInt32(&maxObserved, m, n) {
					break
				}
				continue
			}
			break
		}
		time.Sleep(25 * time.Millisecond)
		atomic.AddInt32(&active, -1)
		return nil
	}

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if maxObserved > 2 {
		t.Fatalf("expected max concurrency <= 2, got %d", maxObserved)
	}
}

// Test that invalid config types are handled per source type without panic.
func TestRunCommand_InvalidConfigTypes(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TargetFolder:    dir,
		ValidExtensions: []string{".epub"},
		// Mismatched types for all cases
		Sources: []config.Source{
			{Type: "nfs", Config: &config.ImapConfig{}},
			{Type: "smb", Config: &config.NfsNetworkShareConfig{}},
			{Type: "imap", Config: &config.SmbNetworkShareConfig{}},
		},
	}

	oldCount := countFiles
	t.Cleanup(func() { countFiles = oldCount })
	countFiles = func(folder string, exts []string, recursive bool) (int, error) { return 0, nil }

	var c RunCommand
	if err := c.Run(cfg, slog.Default()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
