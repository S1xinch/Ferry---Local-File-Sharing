package main

import (
	"os/exec"
	"syscall"
)

// hideWindow stops helper programs from flashing a black console window.
func hideWindow(cmd *exec.Cmd) {
	const createNoWindow = 0x08000000
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
