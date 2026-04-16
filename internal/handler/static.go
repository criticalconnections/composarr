package handler

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var embeddedStatic embed.FS

func init() {
	// Try to load embedded frontend files
	sub, err := fs.Sub(embeddedStatic, "static")
	if err == nil {
		// Check if the directory has real build artifacts (not just the .gitkeep placeholder)
		entries, err := fs.ReadDir(sub, ".")
		if err == nil {
			for _, e := range entries {
				if e.Name() != ".gitkeep" {
					StaticFS = sub
					return
				}
			}
		}
	}
}
