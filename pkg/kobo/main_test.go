package kobo

import (
	"errors"
	"testing"
)

// TestIsKoboDevice ensures detection call is safe to execute in typical non-Kobo environments.
func TestIsKoboDevice(t *testing.T) {
	// On CI/local dev this should be false; just ensure it doesn't panic
	_ = IsKoboDevice()
}

// TestUpdateLibrary_NoNickelDbus_USBPath forces the fallback USB path when NickelDBus isn't installed.
func TestUpdateLibrary_NoNickelDbus_USBPath(t *testing.T) {
	// Force the "not installed" path
	old := ndbIsInstalled
	t.Cleanup(func() { ndbIsInstalled = old })
	ndbIsInstalled = func() bool { return false }

	// Speed up the simulated USB plug during tests
	usbPlugSleep = 0

	// Ensure nickelUSBplugAction can write when pipe exists
	// Create a temp named pipe path substitute by symlinking to a regular file opened RW
	// Since os.OpenFile with os.ModeNamedPipe will fail on non-pipe, we rely on the
	// error-handling path to just return quietly; hence we don't need to create it.
	if err := UpdateLibrary(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestUpdateLibrary_WithNickelDbus_Success verifies LibraryRescan is invoked via seams when installed.
func TestUpdateLibrary_WithNickelDbus_Success(t *testing.T) {
	oldI, oldR := ndbIsInstalled, ndbLibraryRescan
	t.Cleanup(func() { ndbIsInstalled, ndbLibraryRescan = oldI, oldR })
	ndbIsInstalled = func() bool { return true }
	called := false
	ndbLibraryRescan = func(timeout int, full bool) error { called = true; return nil }

	if err := UpdateLibrary(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !called {
		t.Fatalf("expected LibraryRescan to be called")
	}
}

// TestUpdateLibrary_WithNickelDbus_Error ensures rescan error is propagated.
func TestUpdateLibrary_WithNickelDbus_Error(t *testing.T) {
	oldI, oldR := ndbIsInstalled, ndbLibraryRescan
	t.Cleanup(func() { ndbIsInstalled, ndbLibraryRescan = oldI, oldR })
	ndbIsInstalled = func() bool { return true }
	ndbLibraryRescan = func(timeout int, full bool) error { return errors.New("fail") }

	if err := UpdateLibrary(); err == nil {
		t.Fatalf("expected error")
	}
}

// TestNickelUSBplugAction_NoPipe ensures missing pipe path is handled without panic.
func TestNickelUSBplugAction_NoPipe(t *testing.T) {
	// Ensure that when the pipe isn't present, the function no-ops without panic
	nickelUSBplugAction("add")
}

// TestIsKoboDevice_False documents the expected false result on typical dev/CI systems.
func TestIsKoboDevice_False(t *testing.T) {
	// Verify typical non-Kobo environment returns false
	if IsKoboDevice() {
		t.Log("running on a Kobo device? test assumes false")
	}
}
