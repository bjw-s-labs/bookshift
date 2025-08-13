// nickeldbus_seams.go: Test seams for the NickelDbus interactions.
//
// This file defines small function variables that production code uses by
// default, but tests can override to substitute fakes. Keeping these seams in a
// separate file mirrors the pattern used in the syncer packages and keeps the
// main implementation files focused on behavior.
package nickeldbus

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

// Injectable wrappers for testability
var (
	systemBus      = dbus.SystemBus
	objectFor      = func(conn *dbus.Conn) dbus.BusObject { return conn.Object(ndbInterface, ndbObjectPath) }
	introspectCall = func(obj dbus.BusObject) error {
		_, err := introspect.Call(obj)
		return err
	}
	versionCall = func(obj dbus.BusObject) (string, error) {
		var v string
		if err := obj.Call(ndbInterface+".ndbVersion", 0).Store(&v); err != nil {
			return "", err
		}
		return v, nil
	}

	// Dialog seams
	callWithArgs = func(obj dbus.BusObject, method string, args ...interface{}) { obj.Call(method, 0, args...) }

	// Library seams
	addMatch = func(conn *dbus.Conn) error {
		return conn.AddMatchSignal(
			dbus.WithMatchObjectPath(ndbObjectPath),
			dbus.WithMatchInterface(ndbInterface),
			dbus.WithMatchMember("pfmDoneProcessing"),
		)
	}
	signalTo = func(conn *dbus.Conn, ch chan<- *dbus.Signal) { conn.Signal(ch) }
	callNoop = func(obj dbus.BusObject, method string) { obj.Call(method, 0) }
)
