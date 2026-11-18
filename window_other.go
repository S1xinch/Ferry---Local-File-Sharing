//go:build !windows && !(darwin && cgo)

package main

// nativeWindow is not available here (Linux, or a Mac build without cgo), so
// the app opens in a Chrome or Edge app window instead.
func nativeWindow(url string, show, quit <-chan struct{}, onQuit func()) bool { return false }
