// Package browser opens the system default web browser.
package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Run is the process starter; tests may replace it.
var Run = func(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

// OpenBrowser launches the OS default browser at url.
func OpenBrowser(url string) error {
	if url == "" {
		return fmt.Errorf("url is required")
	}
	var name string
	var args []string
	switch runtime.GOOS {
	case "windows":
		name = "cmd"
		args = []string{"/c", "start", "", url}
	case "darwin":
		name = "open"
		args = []string{url}
	default:
		name = "xdg-open"
		args = []string{url}
	}
	if err := Run(name, args...); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
