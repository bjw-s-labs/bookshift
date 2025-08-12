package nickeldbus

import (
	"context"
	"testing"

	"github.com/godbus/dbus/v5"
)

// captureObj records method calls
type captureObj struct{ calls []string }

func (c *captureObj) record(m string) { c.calls = append(c.calls, m) }

// implement dbus.BusObject
func (c *captureObj) Call(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	c.record(method)
	return &dbus.Call{}
}
func (c *captureObj) CallWithContext(_ context.Context, _ string, _ dbus.Flags, _ ...interface{}) *dbus.Call {
	return &dbus.Call{}
}
func (c *captureObj) Go(_ string, _ dbus.Flags, _ chan *dbus.Call, _ ...interface{}) *dbus.Call {
	return &dbus.Call{}
}
func (c *captureObj) GoWithContext(_ context.Context, _ string, _ dbus.Flags, _ chan *dbus.Call, _ ...interface{}) *dbus.Call {
	return &dbus.Call{}
}
func (c *captureObj) AddMatchSignal(_ string, _ string, _ ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}
func (c *captureObj) RemoveMatchSignal(_ string, _ string, _ ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}
func (c *captureObj) GetProperty(string) (dbus.Variant, error) { return dbus.Variant{}, nil }
func (c *captureObj) StoreProperty(string, interface{}) error  { return nil }
func (c *captureObj) SetProperty(string, interface{}) error    { return nil }
func (c *captureObj) Destination() string                      { return "dest" }
func (c *captureObj) Path() dbus.ObjectPath                    { return dbus.ObjectPath("/x") }

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
