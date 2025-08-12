// Package nickeldbus implements all NickelDbus interactions of KoboMail
package nickeldbus

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

// LibraryRescan sends a request to completely the library
func LibraryRescan(timeoutSeconds int, fullScan bool) error {
	rescanSignal := make(chan *dbus.Signal, 10)
	ndbConn, _ := getSystemDbusConnection()

	// Subscribe to the pfmDoneProcessing signal
	if err := addMatch(ndbConn); err != nil {
		return fmt.Errorf("library rescan: error while adding match signal: %w", err)
	}
	signalTo(ndbConn, rescanSignal)

	// Trigger the rescan
	var scanType = "pfmRescanBooks"
	if fullScan {
		scanType = "pfmRescanBooksFull"
	}

	ndbObj, _ := getNdbObject(ndbConn)
	callNoop(ndbObj, ndbInterface+"."+scanType)

	// Wait for the pfmDoneProcessing signal or timeout
	select {
	case rs := <-rescanSignal:
		valid, err := isDoneProcessingSignal(rs)
		if err != nil {
			return fmt.Errorf("library rescan error: %w", err)
		} else if !valid {
			return fmt.Errorf("library rescan error: expected 'pfmDoneProcessing', got '%s'", rs.Name)
		}
	case <-time.After(time.Duration(timeoutSeconds) * time.Second):
		return fmt.Errorf("library rescan: timeout waiting for rescan to complete")
	}
	return nil
}

func isDoneProcessingSignal(rs *dbus.Signal) (bool, error) {
	if rs.Name != ndbInterface+".pfmDoneProcessing" {
		return false, fmt.Errorf("isDoneProcessingSignal: not valid 'pfmDoneProcessing' signal")
	}
	return true, nil
}

// Injectable wrappers
var (
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
