package main

import (
	"fmt"
	"io"
	"os"
)

const maxStdinBytes = 102400 // 100KB

// readStdin reads from stdin if it's a pipe (not a terminal).
// Returns the content, whether it was truncated, and any error.
func readStdin() (string, bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", false, err
	}

	// Check if stdin is a pipe (not a terminal)
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", false, nil
	}

	// Read up to maxStdinBytes + 1 to detect truncation
	buf := make([]byte, maxStdinBytes+1)
	n, err := io.ReadFull(os.Stdin, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", false, fmt.Errorf("reading stdin: %v", err)
	}

	truncated := n > maxStdinBytes
	if truncated {
		n = maxStdinBytes
	}

	return string(buf[:n]), truncated, nil
}
