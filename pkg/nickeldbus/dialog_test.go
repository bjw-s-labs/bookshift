package nickeldbus

import (
	"errors"
	"testing"

	"github.com/godbus/dbus/v5"
)

// TestDialogCreate_Update_OK verifies DialogCreate, DialogUpdate, and DialogAddOKButton
// invoke the expected DBus methods in order via the seams.
func TestDialogCreate_Update_OK(t *testing.T) {
	origObj, origCall, origBus := objectFor, callWithArgs, systemBus
	t.Cleanup(func() { objectFor, callWithArgs, systemBus = origObj, origCall, origBus })

	co := &captureObj{}
	objectFor = func(*dbus.Conn) dbus.BusObject { return co }
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	callWithArgs = func(obj dbus.BusObject, method string, args ...interface{}) { obj.Call(method, 0, args...) }

	DialogCreate("hello")
	DialogUpdate("bye")
	DialogAddOKButton()

	want := []string{
		ndbInterface + ".dlgConfirmCreate",
		ndbInterface + ".dlgConfirmSetTitle",
		ndbInterface + ".dlgConfirmSetBody",
		ndbInterface + ".dlgConfirmSetModal",
		ndbInterface + ".dlgConfirmShowClose",
		ndbInterface + ".dlgConfirmShow",
		ndbInterface + ".dlgConfirmSetBody",
		ndbInterface + ".dlgConfirmSetAccept",
	}
	if len(co.calls) != len(want) {
		t.Fatalf("want %d calls, got %d: %#v", len(want), len(co.calls), co.calls)
	}
	for i, m := range want {
		if co.calls[i] != m {
			t.Fatalf("call %d: want %q, got %q", i, m, co.calls[i])
		}
	}
}

// TestDialog_ErrorOnGetObject ensures dialog functions return gracefully
// when the DBus connection cannot be established and do not invoke calls.
func TestDialog_ErrorOnGetObject(t *testing.T) {
	origObj, origCall, origBus := objectFor, callWithArgs, systemBus
	t.Cleanup(func() { objectFor, callWithArgs, systemBus = origObj, origCall, origBus })

	// Simulate system bus error
	systemBus = func() (*dbus.Conn, error) { return nil, errors.New("boom") }

	// Capture any attempted calls (should be none)
	var calls []string
	callWithArgs = func(obj dbus.BusObject, method string, args ...interface{}) { calls = append(calls, method) }

	// Run dialog functions
	DialogCreate("hello")
	DialogUpdate("bye")
	DialogAddOKButton()

	if len(calls) != 0 {
		t.Fatalf("expected no DBus calls when get object fails, got: %#v", calls)
	}
}
