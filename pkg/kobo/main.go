package kobo

import (
	"fmt"
	"os"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/nickeldbus"
)

// nickelUSBplugAddRemove simulates pugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAction(action string) {
	const nickelHWstatusPipe = "/tmp/nickel-hardware-status"

	nickelPipe, _ := os.OpenFile(nickelHWstatusPipe, os.O_RDWR, os.ModeNamedPipe)
	nickelPipe.WriteString("usb plug " + action)
	nickelPipe.Close()
}

// nickelUSBplugAddRemove simulates pugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAddRemove() {
	nickelUSBplugAction("add")
	time.Sleep(10 * time.Second)
	nickelUSBplugAction("remove")
}

func UpdateLibrary() error {
	if nickeldbus.IsInstalled() {
		if err := nickeldbus.LibraryRescan(30000, true); err != nil {
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
