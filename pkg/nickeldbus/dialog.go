// Dialog APIs to show notifications via NickelDbus.
package nickeldbus

import "log"

// DialogCreate creates a dialog to show a notification to the user
func DialogCreate(initialMsg string) {
	ndbObj, err := getNdbObject(nil)
	if err != nil {
		log.Printf("nickeldbus: DialogCreate: getNdbObject error: %v", err)
		return
	}
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmCreate")
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetTitle", "KoboMail")
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetBody", initialMsg)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetModal", false)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmShowClose", false)
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmShow")
}

// DialogUpdate updates a dialog with a new body
func DialogUpdate(body string) {
	ndbObj, err := getNdbObject(nil)
	if err != nil {
		log.Printf("nickeldbus: DialogUpdate: getNdbObject error: %v", err)
		return
	}
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetBody", body)
}

// DialogAddOKButton updates a dialog with a confirmation button
func DialogAddOKButton() {
	ndbObj, err := getNdbObject(nil)
	if err != nil {
		log.Printf("nickeldbus: DialogAddOKButton: getNdbObject error: %v", err)
		return
	}
	callWithArgs(ndbObj, ndbInterface+".dlgConfirmSetAccept", "OK")
}
