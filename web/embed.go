// Package web embeds the built OBS overlay, so whipbot runs as a single
// binary from any directory. Build the overlay before the Go code.
package web

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Overlay returns the built overlay files.
func Overlay() (fs.FS, error) {
	overlay, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded dist: %w", err)
	}

	return overlay, nil
}
