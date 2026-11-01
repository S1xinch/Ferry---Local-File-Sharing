//go:build !windows

package main

import "syscall"

// allowBroadcast lets the finder send to broadcast addresses. If the system
// refuses, the finder still asks each address one by one.
func allowBroadcast(network, address string, c syscall.RawConn) error {
	c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	})
	return nil
}
