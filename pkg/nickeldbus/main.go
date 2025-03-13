// Package nickeldbus implements all NickelDbus interactions of KoboMail
package nickeldbus

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	ndbInterface  = "com.github.shermp.nickeldbus"
	ndbObjectPath = "/nickeldbus"
)

func getSystemDbusConnection() (*dbus.Conn, error) {
	var err error
	var ndbConn *dbus.Conn

	ndbConn, err = dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	return ndbConn, nil
}

func getNdbObject(conn *dbus.Conn) (dbus.BusObject, error) {
	var err error
	if conn == nil {
		conn, err = getSystemDbusConnection()
		if err != nil {
			return nil, err
		}
	}
	return conn.Object(ndbInterface, ndbObjectPath), nil
}

// IsInstalled returns if NickelDbus is installed
func IsInstalled() bool {
	var err error
	var ndbObj dbus.BusObject

	ndbObj, err = getNdbObject(nil)
	if err != nil {
		return false
	}

	_, err = introspect.Call(ndbObj)
	installed := err == nil

	return installed
}

// GetVersion returns the current NickelDbus version
func GetVersion() (string, error) {
	var err error
	var ndbVersion string
	var ndbObj dbus.BusObject

	ndbObj, err = getNdbObject(nil)
	if err != nil {
		return "", err
	}

	err = ndbObj.Call(ndbInterface+".ndbVersion", 0).Store(&ndbVersion)
	if err != nil {
		return "", err
	}

	return ndbVersion, nil
}
