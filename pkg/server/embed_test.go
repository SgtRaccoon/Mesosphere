package server

import (
	"io/fs"
	"testing"
)

func TestUIAssetsEmbedded(t *testing.T) {
	data, err := UIAssets.ReadFile("dist/index.html")
	if err != nil {
		t.Fatalf("embed index.html: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty index.html")
	}
	sub, err := DistFS()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(sub, "index.html"); err != nil {
		t.Fatal(err)
	}
}
