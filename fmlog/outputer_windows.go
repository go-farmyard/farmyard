package log

import (
	"golang.org/x/sys/windows"
)

func enableVTMode(console windows.Handle) bool {
	mode := uint32(0)
	err := windows.GetConsoleMode(console, &mode)
	if err != nil {
		return false
	}

	// EnableVirtualTerminalProcessing is the console mode to allow ANSI code interpretation on the console.
	// See: https://docs.microsoft.com/en-us/windows/console/setconsolemode
	// For Windows 10 or later
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	err = windows.SetConsoleMode(console, mode)
	return err == nil
}

func init() {
	enableVTMode(windows.Stdout)
	enableVTMode(windows.Stderr)
}
