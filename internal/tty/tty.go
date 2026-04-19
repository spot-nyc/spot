// Package tty provides a minimal TTY-detection helper for the Spot CLI.
package tty

import "golang.org/x/term"

// IsTerminal reports whether the given file descriptor refers to a terminal.
// Wraps term.IsTerminal with an int conversion so callers can pass *os.File.Fd()
// directly (which returns uintptr) without conversion noise.
func IsTerminal(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
