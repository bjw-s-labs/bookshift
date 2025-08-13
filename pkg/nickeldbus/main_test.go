package nickeldbus

import (
	"errors"
	"testing"

	"github.com/godbus/dbus/v5"
)

// TestIsInstalled_TrueAndFalse verifies both the installed and not-installed
// paths by toggling the introspection call result.
func TestIsInstalled_TrueAndFalse(t *testing.T) {
	// installed
	origBus, origObj, origIntro := systemBus, objectFor, introspectCall
	t.Cleanup(func() { systemBus, objectFor, introspectCall = origBus, origObj, origIntro })
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	objectFor = func(conn *dbus.Conn) dbus.BusObject { return fakeObj{} }
	introspectCall = func(obj dbus.BusObject) error { return nil }
	if !IsInstalled() {
		t.Fatalf("expected installed true")
	}

	// not installed via introspect error
	introspectCall = func(obj dbus.BusObject) error { return errors.New("x") }
	if IsInstalled() {
		t.Fatalf("expected installed false")
	}
}

// TestGetVersion_Success ensures version retrieval succeeds via the seam.
func TestGetVersion_Success(t *testing.T) {
	origBus, origObj, origVer := systemBus, objectFor, versionCall
	t.Cleanup(func() { systemBus, objectFor, versionCall = origBus, origObj, origVer })
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	objectFor = func(conn *dbus.Conn) dbus.BusObject { return fakeObj{} }
	versionCall = func(obj dbus.BusObject) (string, error) { return "1.2.3", nil }
	v, err := GetVersion()
	if err != nil || v != "1.2.3" {
		t.Fatalf("want 1.2.3, got %q err=%v", v, err)
	}
}

// TestIsInstalled_SystemBusError ensures IsInstalled returns false when
// connecting to the system bus fails.
func TestIsInstalled_SystemBusError(t *testing.T) {
	orig := systemBus
	t.Cleanup(func() { systemBus = orig })
	systemBus = func() (*dbus.Conn, error) { return nil, errors.New("boom") }
	if IsInstalled() {
		t.Fatalf("expected false when system bus errors")
	}
}

// TestGetVersion_SystemBusError ensures GetVersion errors when system bus fails.
func TestGetVersion_SystemBusError(t *testing.T) {
	orig := systemBus
	t.Cleanup(func() { systemBus = orig })
	systemBus = func() (*dbus.Conn, error) { return nil, errors.New("boom") }
	if _, err := GetVersion(); err == nil {
		t.Fatalf("expected error from GetVersion when system bus fails")
	}
}

// TestGetVersion_CallError ensures a failure in the version call propagates an error.
func TestGetVersion_CallError(t *testing.T) {
	origBus, origObj, origVer := systemBus, objectFor, versionCall
	t.Cleanup(func() { systemBus, objectFor, versionCall = origBus, origObj, origVer })
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	objectFor = func(conn *dbus.Conn) dbus.BusObject { return fakeObj{} }
	versionCall = func(obj dbus.BusObject) (string, error) { return "", errors.New("call failed") }
	if _, err := GetVersion(); err == nil {
		t.Fatalf("expected error from version call failure")
	}
}
