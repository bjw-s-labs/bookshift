package nickeldbus

import (
	"context"

	"github.com/godbus/dbus/v5"
)

// captureObj records method calls to validate DBus invocations in tests.
type captureObj struct{ calls []string }

func (c *captureObj) record(m string) { c.calls = append(c.calls, m) }

// implement dbus.BusObject for captureObj
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

// fakeObj is a simple no-op BusObject used in tests that don't care about method capture.
type fakeObj struct{}

func (fakeObj) Call(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return &dbus.Call{}
}
func (fakeObj) CallWithContext(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return &dbus.Call{}
}
func (fakeObj) Go(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return &dbus.Call{}
}
func (fakeObj) GoWithContext(ctx context.Context, method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return &dbus.Call{}
}
func (fakeObj) AddMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}
func (fakeObj) RemoveMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}
func (fakeObj) GetProperty(p string) (dbus.Variant, error)      { return dbus.Variant{}, nil }
func (fakeObj) StoreProperty(p string, value interface{}) error { return nil }
func (fakeObj) SetProperty(p string, v interface{}) error       { return nil }
func (fakeObj) Destination() string                             { return "dest" }
func (fakeObj) Path() dbus.ObjectPath                           { return dbus.ObjectPath("/x") }
