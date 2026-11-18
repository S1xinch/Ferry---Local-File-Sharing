package main

import (
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

var (
	user32          = windows.NewLazySystemDLL("user32.dll")
	procLoadImage   = user32.NewProc("LoadImageW")
	procSendMessage = user32.NewProc("SendMessageW")
	procShowWindow  = user32.NewProc("ShowWindow")
	procSetFg       = user32.NewProc("SetForegroundWindow")
)

// nativeWindow shows the app in its own window with WebView2, which comes
// with Windows 10 and 11, so the taskbar shows Ferry and not Edge. It returns
// false straight away if WebView2 is missing, so the caller can open a browser
// window instead. Closing the window keeps Ferry running; a value on show
// opens it again. It returns once quit is closed.
func nativeWindow(url string, show, quit <-chan struct{}, _ func()) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	for first := true; ; first = false {
		w := webview2.NewWithOptions(webview2.WebViewOptions{
			AutoFocus: true,
			DataPath:  filepath.Join(configDir(), "webview"),
			WindowOptions: webview2.WindowOptions{
				Title: "Ferry", Width: 920, Height: 760, Center: true,
			},
		})
		if w == nil {
			if first {
				return false
			}
			openWindow(url) // WebView2 stopped working: fall back to a browser window
		} else {
			hwnd := uintptr(w.Window())
			setAppIcon(hwnd)
			w.SetSize(480, 400, webview2.HintMin)
			w.Navigate(url)
			closed := make(chan struct{})
			go func() {
				for {
					select {
					case <-quit:
						w.Destroy()
						return
					case <-show:
						w.Dispatch(func() { bringToFront(hwnd) })
					case <-closed:
						return
					}
				}
			}()
			w.Run()
			close(closed)
		}
		select {
		case <-quit:
			return true
		case <-show:
		}
	}
}

// setAppIcon gives the window Ferry's icon from the exe's resources, which
// go-winres adds as "APP". Without it the taskbar shows a blank program icon.
func setAppIcon(hwnd uintptr) {
	const imageIcon, lrDefaultSize, lrShared, wmSetIcon, iconSmall, iconBig = 1, 0x40, 0x8000, 0x80, 0, 1
	var hinst windows.Handle
	windows.GetModuleHandleEx(0, nil, &hinst)
	name, _ := windows.UTF16PtrFromString("APP")
	load := func(size, flags uintptr) uintptr {
		h, _, _ := procLoadImage.Call(uintptr(hinst), uintptr(unsafe.Pointer(name)), imageIcon, size, size, flags)
		return h
	}
	if big := load(0, lrDefaultSize|lrShared); big != 0 {
		procSendMessage.Call(hwnd, wmSetIcon, iconBig, big)
	}
	if small := load(16, lrShared); small != 0 {
		procSendMessage.Call(hwnd, wmSetIcon, iconSmall, small)
	}
}

func bringToFront(hwnd uintptr) {
	const swRestore = 9
	procShowWindow.Call(hwnd, swRestore)
	procSetFg.Call(hwnd)
}
