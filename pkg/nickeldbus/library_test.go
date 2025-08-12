package nickeldbus

import (
	"fmt"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// TestIsDoneProcessingSignal verifies that only the pfmDoneProcessing signal is accepted.
func TestIsDoneProcessingSignal(t *testing.T) {
	s := &dbus.Signal{Name: ndbInterface + ".pfmDoneProcessing"}
	ok, err := isDoneProcessingSignal(s)
	if err != nil || !ok {
		t.Fatalf("expected ok, got err=%v", err)
	}
	s2 := &dbus.Signal{Name: ndbInterface + ".other"}
	ok, err = isDoneProcessingSignal(s2)
	if err == nil || ok {
		t.Fatalf("expected error for other signal")
	}
}

type fakeConn struct{ sigCh chan<- *dbus.Signal }

func (f *fakeConn) AddMatchSignal(options ...dbus.MatchOption) error { return nil }
func (f *fakeConn) Signal(ch chan<- *dbus.Signal)                    { f.sigCh = ch }

// TestLibraryRescan_Success simulates a happy path where the done signal arrives.
func TestLibraryRescan_Success(t *testing.T) {
	// inject wrappers
	origAdd, origSig, origBus, origObj, origCall := addMatch, signalTo, systemBus, objectFor, callNoop
	t.Cleanup(func() {
		addMatch, signalTo, systemBus, objectFor, callNoop = origAdd, origSig, origBus, origObj, origCall
	})

	fc := &fakeConn{}
	addMatch = func(conn *dbus.Conn) error { return fc.AddMatchSignal() }
	signalTo = func(conn *dbus.Conn, ch chan<- *dbus.Signal) { fc.Signal(ch) }
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	objectFor = func(conn *dbus.Conn) dbus.BusObject { return fakeObj{} }
	callNoop = func(obj dbus.BusObject, method string) {
		// simulate async completion
		go func() {
			time.Sleep(10 * time.Millisecond)
			fc.sigCh <- &dbus.Signal{Name: ndbInterface + ".pfmDoneProcessing"}
		}()
	}

	if err := LibraryRescan(1, false); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

// TestLibraryRescan_Timeout verifies the call times out when no signal is received.
func TestLibraryRescan_Timeout(t *testing.T) {
	origAdd, origSig, origBus, origObj, origCall := addMatch, signalTo, systemBus, objectFor, callNoop
	t.Cleanup(func() {
		addMatch, signalTo, systemBus, objectFor, callNoop = origAdd, origSig, origBus, origObj, origCall
	})
	fc := &fakeConn{}
	addMatch = func(conn *dbus.Conn) error { return fc.AddMatchSignal() }
	signalTo = func(conn *dbus.Conn, ch chan<- *dbus.Signal) { fc.Signal(ch) }
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	objectFor = func(conn *dbus.Conn) dbus.BusObject { return fakeObj{} }
	callNoop = func(obj dbus.BusObject, method string) {}

	if err := LibraryRescan(0, false); err == nil {
		t.Fatalf("expected timeout error")
	}
}

// TestLibraryRescan_WrongSignal ensures a different signal results in an error.
func TestLibraryRescan_WrongSignal(t *testing.T) {
	origAdd, origSig, origBus, origObj, origCall := addMatch, signalTo, systemBus, objectFor, callNoop
	t.Cleanup(func() {
		addMatch, signalTo, systemBus, objectFor, callNoop = origAdd, origSig, origBus, origObj, origCall
	})
	fc := &fakeConn{}
	addMatch = func(conn *dbus.Conn) error { return fc.AddMatchSignal() }
	signalTo = func(conn *dbus.Conn, ch chan<- *dbus.Signal) { fc.Signal(ch) }
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	objectFor = func(conn *dbus.Conn) dbus.BusObject { return fakeObj{} }
	callNoop = func(obj dbus.BusObject, method string) {
		go func() { fc.sigCh <- &dbus.Signal{Name: ndbInterface + ".other"} }()
	}

	if err := LibraryRescan(1, true); err == nil {
		t.Fatalf("expected wrong-signal error")
	}
}

// TestLibraryRescan_AddMatchError ensures AddMatch failure bubbles up.
func TestLibraryRescan_AddMatchError(t *testing.T) {
	origAdd, origSig, origBus, origObj, origCall := addMatch, signalTo, systemBus, objectFor, callNoop
	t.Cleanup(func() {
		addMatch, signalTo, systemBus, objectFor, callNoop = origAdd, origSig, origBus, origObj, origCall
	})
	addMatch = func(conn *dbus.Conn) error { return fmt.Errorf("nope") }
	signalTo = func(conn *dbus.Conn, ch chan<- *dbus.Signal) {}
	systemBus = func() (*dbus.Conn, error) { return &dbus.Conn{}, nil }
	objectFor = func(conn *dbus.Conn) dbus.BusObject { return fakeObj{} }
	callNoop = func(obj dbus.BusObject, method string) {}
	if err := LibraryRescan(1, false); err == nil {
		t.Fatalf("expected add-match error")
	}
}
