package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpOutput(t *testing.T) {
	c := NewRootCommand()
	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs([]string{"--help"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "mesosphere") || !strings.Contains(out, "serve") {
		t.Fatalf("help missing expected text:\n%s", out)
	}
}

func TestZeroArgsInvokesServe(t *testing.T) {
	called := false
	orig := Serve
	Serve = func(addr string, _ ...bool) error {
		called = true
		return nil
	}
	t.Cleanup(func() { Serve = orig })

	c := NewRootCommand()
	c.SetArgs([]string{})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected Serve on zero args")
	}
}

func TestServeCommandInvokesServe(t *testing.T) {
	var got string
	orig := Serve
	Serve = func(addr string, _ ...bool) error {
		got = addr
		return nil
	}
	t.Cleanup(func() { Serve = orig })

	c := NewRootCommand()
	c.SetArgs([]string{"serve", "--addr", "127.0.0.1:0"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if got != "127.0.0.1:0" {
		t.Fatalf("addr = %q", got)
	}
}
