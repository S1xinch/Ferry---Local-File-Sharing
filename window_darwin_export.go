//go:build darwin && cgo

package main

import "C"

// ferryQuitRequested is called from the Mac menu (Cmd+Q). It is in its own
// file because cgo does not allow exports next to C definitions.
//
//export ferryQuitRequested
func ferryQuitRequested() {
	if quitHook != nil {
		go quitHook()
	}
}
