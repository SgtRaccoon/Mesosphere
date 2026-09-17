package browser

import (
	"runtime"
	"testing"
)

func TestOpenBrowserInvokesLauncher(t *testing.T) {
	var name string
	var args []string
	orig := Run
	Run = func(n string, a ...string) error {
		name = n
		args = a
		return nil
	}
	t.Cleanup(func() { Run = orig })

	if err := OpenBrowser("http://localhost:8080"); err != nil {
		t.Fatal(err)
	}
	switch runtime.GOOS {
	case "windows":
		if name != "cmd" || len(args) < 3 || args[len(args)-1] != "http://localhost:8080" {
			t.Fatalf("windows launcher = %s %v", name, args)
		}
	case "darwin":
		if name != "open" {
			t.Fatalf("got %s", name)
		}
	default:
		if name != "xdg-open" {
			t.Fatalf("got %s", name)
		}
	}
}

func TestOpenBrowserEmptyURL(t *testing.T) {
	if err := OpenBrowser(""); err == nil {
		t.Fatal("expected error")
	}
}
