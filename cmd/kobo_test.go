package cmd

import "testing"

// TestKoboCommands_Run verifies the Kobo uninstall and update subcommands execute
// without side effects by stubbing updateKoboLibrary to avoid delays, and return no errors.
func TestKoboCommands_Run(t *testing.T) {
	if err := (&KoboUninstallCommand{}).Run(nil); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	// Override update seam to avoid 10s sleep via kobo.UpdateLibrary fallback
	old := updateKoboLibrary
	t.Cleanup(func() { updateKoboLibrary = old })
	updateKoboLibrary = func() error { return nil }
	if err := (&KoboUpdateLibraryCommand{}).Run(nil); err != nil {
		t.Fatalf("update library: %v", err)
	}
}
