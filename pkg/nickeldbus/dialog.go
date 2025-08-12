// Package nickeldbus implements all NickelDbus interactions of KoboMail
package nickeldbus

import "github.com/godbus/dbus/v5"

// DialogCreate creates a dialog to show a notification to the user
func DialogCreate(initialMsg string) {
	ndbObj, _ := getNdbObject(nil)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmCreate")
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetTitle", "KoboMail")
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetBody", initialMsg)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetModal", false)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmShowClose", false)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmShow")
}

// DialogUpdate updates a dialog with a new body
func DialogUpdate(body string) {
	ndbObj, _ := getNdbObject(nil)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetBody", body)
}

// DialogAddOKButton updates a dialog with a confirmation button
func DialogAddOKButton() {
	ndbObj, _ := getNdbObject(nil)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetAccept", "OK")
}

// Injectable call wrapper for tests
var callWithArgs = func(obj dbus.BusObject, method string, args ...interface{}) { obj.Call(method, 0, args...) }
