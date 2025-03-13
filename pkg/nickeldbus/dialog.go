// Package nickeldbus implements all NickelDbus interactions of KoboMail
package nickeldbus

// DialogCreate creates a dialog to show a notification to the user
func DialogCreate(initialMsg string) {
	ndbObj, _ := getNdbObject(nil)
	ndbObj.Call(ndbInterface+".dlgConfirmCreate", 0)
	ndbObj.Call(ndbInterface+".dlgConfirmSetTitle", 0, "KoboMail")
	ndbObj.Call(ndbInterface+".dlgConfirmSetBody", 0, initialMsg)
	ndbObj.Call(ndbInterface+".dlgConfirmSetModal", 0, false)
	ndbObj.Call(ndbInterface+".dlgConfirmShowClose", 0, false)
	ndbObj.Call(ndbInterface+".dlgConfirmShow", 0)
}

// DialogUpdate updates a dialog with a new body
func DialogUpdate(body string) {
	ndbObj, _ := getNdbObject(nil)
	ndbObj.Call(ndbInterface+".dlgConfirmSetBody", 0, body)
}

// DialogAddOKButton updates a dialog with a confirmation button
func DialogAddOKButton() {
	ndbObj, _ := getNdbObject(nil)
	ndbObj.Call(ndbInterface+".dlgConfirmSetAccept", 0, "OK")
}
