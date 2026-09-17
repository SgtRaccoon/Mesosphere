package server

import (
	"embed"
	"io/fs"
)

// UIAssets is the compiled web UI. Files live under dist/ (copied from web/dist).
//
//go:embed all:dist
var UIAssets embed.FS

// DistFS returns the embedded dist tree rooted at "dist".
func DistFS() (fs.FS, error) {
	return fs.Sub(UIAssets, "dist")
}
