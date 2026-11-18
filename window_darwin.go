//go:build darwin && cgo

package main

/*
#cgo CFLAGS: -x objective-c -fno-objc-arc -Wno-deprecated-declarations
#cgo LDFLAGS: -framework Cocoa -framework WebKit
#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#include <stdlib.h>

void ferryQuitRequested(void);

static NSString *ferryURL;
static NSWindow *ferryWin;
static id ferryDel;

static void ferryShow(void);

@interface FerryDelegate : NSObject <NSApplicationDelegate, NSWindowDelegate>
@end

@implementation FerryDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)n { ferryShow(); }
// Clicking the Dock icon, or opening Ferry again, brings the window back.
- (BOOL)applicationShouldHandleReopen:(NSApplication *)app hasVisibleWindows:(BOOL)visible {
	ferryShow();
	return NO;
}
// Cmd+Q and Quit go through Ferry, which stops sharing and then ends the loop.
- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)app {
	ferryQuitRequested();
	return NSTerminateCancel;
}
- (void)windowWillClose:(NSNotification *)n { ferryWin = nil; }
@end

static void ferryShow(void) {
	if (ferryWin != nil) {
		[ferryWin makeKeyAndOrderFront:nil];
		[NSApp activateIgnoringOtherApps:YES];
		return;
	}
	NSRect r = NSMakeRect(0, 0, 920, 760);
	ferryWin = [[NSWindow alloc] initWithContentRect:r
		styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable
		backing:NSBackingStoreBuffered defer:NO];
	[ferryWin setTitle:@"Ferry"];
	[ferryWin setMinSize:NSMakeSize(480, 400)];
	[ferryWin setDelegate:ferryDel];
	WKWebViewConfiguration *cfg = [[[WKWebViewConfiguration alloc] init] autorelease];
	WKWebView *web = [[[WKWebView alloc] initWithFrame:r configuration:cfg] autorelease];
	[ferryWin setContentView:web];
	[web loadRequest:[NSURLRequest requestWithURL:[NSURL URLWithString:ferryURL]]];
	[ferryWin center];
	[ferryWin makeKeyAndOrderFront:nil];
	[NSApp activateIgnoringOtherApps:YES];
}

static NSMenuItem *ferryItem(NSString *title, SEL action, NSString *key) {
	return [[[NSMenuItem alloc] initWithTitle:title action:action keyEquivalent:key] autorelease];
}

// The menu bar: Quit, plus Edit so copy and paste shortcuts work in the page.
static void ferryMenu(void) {
	NSMenu *bar = [[[NSMenu alloc] init] autorelease];
	NSMenu *app = [[[NSMenu alloc] initWithTitle:@"Ferry"] autorelease];
	[app addItem:ferryItem(@"Hide Ferry", @selector(hide:), @"h")];
	[app addItem:[NSMenuItem separatorItem]];
	[app addItem:ferryItem(@"Quit Ferry", @selector(terminate:), @"q")];
	NSMenu *edit = [[[NSMenu alloc] initWithTitle:@"Edit"] autorelease];
	[edit addItem:ferryItem(@"Undo", @selector(undo:), @"z")];
	[edit addItem:ferryItem(@"Redo", @selector(redo:), @"Z")];
	[edit addItem:[NSMenuItem separatorItem]];
	[edit addItem:ferryItem(@"Cut", @selector(cut:), @"x")];
	[edit addItem:ferryItem(@"Copy", @selector(copy:), @"c")];
	[edit addItem:ferryItem(@"Paste", @selector(paste:), @"v")];
	[edit addItem:ferryItem(@"Select All", @selector(selectAll:), @"a")];
	NSMenu *win = [[[NSMenu alloc] initWithTitle:@"Window"] autorelease];
	[win addItem:ferryItem(@"Minimize", @selector(performMiniaturize:), @"m")];
	[win addItem:ferryItem(@"Close", @selector(performClose:), @"w")];
	for (NSMenu *m in @[app, edit, win]) {
		NSMenuItem *top = [[[NSMenuItem alloc] init] autorelease];
		[top setSubmenu:m];
		[bar addItem:top];
	}
	[NSApp setMainMenu:bar];
	[NSApp setWindowsMenu:win];
}

static void ferryRun(const char *url) {
	@autoreleasepool {
		[NSApplication sharedApplication];
		[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
		ferryURL = [[NSString alloc] initWithUTF8String:url];
		ferryDel = [[FerryDelegate alloc] init];
		[NSApp setDelegate:ferryDel];
		ferryMenu();
		[NSApp run];
	}
}

static void ferryShowLater(void) {
	dispatch_async(dispatch_get_main_queue(), ^{ ferryShow(); });
}

// ferryStopLater ends [NSApp run] so Go can clean up and exit normally.
static void ferryStopLater(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		if (ferryWin != nil) [ferryWin close];
		[NSApp stop:nil];
		NSEvent *e = [NSEvent otherEventWithType:NSEventTypeApplicationDefined location:NSZeroPoint
			modifierFlags:0 timestamp:0 windowNumber:0 context:nil subtype:0 data1:0 data2:0];
		[NSApp postEvent:e atStart:YES];
	});
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Cocoa only works from the main thread, which main() starts on.
func init() { runtime.LockOSThread() }

var quitHook func()

// nativeWindow shows the app in its own window with WKWebView, so the Dock
// shows Ferry and not a browser. Closing the window keeps Ferry running;
// clicking the Dock icon or a value on show opens it again. It returns once
// quit is closed. onQuit is called for Cmd+Q and the Quit menu.
func nativeWindow(url string, show, quit <-chan struct{}, onQuit func()) bool {
	quitHook = onQuit
	go func() {
		for {
			select {
			case <-show:
				C.ferryShowLater()
			case <-quit:
				C.ferryStopLater()
				return
			}
		}
	}()
	cu := C.CString(url)
	defer C.free(unsafe.Pointer(cu))
	C.ferryRun(cu)
	return true
}
