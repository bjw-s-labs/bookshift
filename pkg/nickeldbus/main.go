// Package nickeldbus implements NickelDbus interactions used by bookshift.
package nickeldbus

import (
	"github.com/godbus/dbus/v5"
)

const (
	ndbInterface  = "com.github.shermp.nickeldbus"
	ndbObjectPath = "/nickeldbus"
)

func getSystemDbusConnection() (*dbus.Conn, error) {
	return systemBus()
}

func getNdbObject(conn *dbus.Conn) (dbus.BusObject, error) {
	var err error
	if conn == nil {
		conn, err = getSystemDbusConnection()
		if err != nil {
			return nil, err
		}
	}
	return objectFor(conn), nil
}

// IsInstalled returns if NickelDbus is installed
func IsInstalled() bool {
	var err error
	var ndbObj dbus.BusObject

	ndbObj, err = getNdbObject(nil)
	if err != nil {
		return false
	}

	err = introspectCall(ndbObj)
	installed := err == nil

	return installed
}

// GetVersion returns the current NickelDbus version
func GetVersion() (string, error) {
	var err error
	var ndbObj dbus.BusObject

	ndbObj, err = getNdbObject(nil)
	if err != nil {
		return "", err
	}
	return versionCall(ndbObj)
}
