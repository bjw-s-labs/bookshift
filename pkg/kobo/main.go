package kobo

import (
	"fmt"
	"os"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/nickeldbus"
)

// usbPlugSleep controls how long to wait between simulated USB plug add/remove.
// Overridden in tests to speed up execution.
var usbPlugSleep = 10 * time.Second

// seams for tests
var (
	ndbIsInstalled   = nickeldbus.IsInstalled
	ndbLibraryRescan = nickeldbus.LibraryRescan
)

// nickelUSBplugAddRemove simulates plugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAction(action string) {
	const nickelHWstatusPipe = "/tmp/nickel-hardware-status"

	nickelPipe, err := os.OpenFile(nickelHWstatusPipe, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		// If the pipe does not exist just return silently (e.g. during tests)
		return
	}
	defer nickelPipe.Close()
	_, _ = nickelPipe.WriteString("usb plug " + action)
}

// nickelUSBplugAddRemove simulates plugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAddRemove() {
	nickelUSBplugAction("add")
	time.Sleep(usbPlugSleep)
	nickelUSBplugAction("remove")
}

func UpdateLibrary() error {
	if ndbIsInstalled() {
		if err := ndbLibraryRescan(30000, true); err != nil {
			return fmt.Errorf("could not update library: %w", err)
		}
	} else {
		nickelUSBplugAddRemove()
	}
	return nil
}

func IsKoboDevice() bool {
	_, err := os.Stat("/usr/local/Kobo")
	return err == nil
}
