// Library APIs to control the Nickel library via DBus.
package nickeldbus

import (
	"context"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

// LibraryRescan sends a request to completely the library
func LibraryRescan(timeoutSeconds int, fullScan bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	return LibraryRescanContext(ctx, fullScan)
}

// LibraryRescanContext sends a request to rescan the library, honoring the provided context for cancellation/timeout.
func LibraryRescanContext(ctx context.Context, fullScan bool) error {
	rescanSignal := make(chan *dbus.Signal, 10)
	ndbConn, err := getSystemDbusConnection()
	if err != nil {
		return fmt.Errorf("library rescan: system bus: %w", err)
	}

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
	case <-ctx.Done():
		return fmt.Errorf("library rescan: timeout/canceled: %w", ctx.Err())
	}
	return nil
}

func isDoneProcessingSignal(rs *dbus.Signal) (bool, error) {
	if rs.Name != ndbInterface+".pfmDoneProcessing" {
		return false, fmt.Errorf("isDoneProcessingSignal: not valid 'pfmDoneProcessing' signal")
	}
	return true, nil
}
