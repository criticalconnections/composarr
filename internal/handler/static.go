package handler

import (
	"embed"
	"io/fs"
)

//go:embed static
var embeddedStatic embed.FS

func init() {
	// Try to load embedded frontend files
	sub, err := fs.Sub(embeddedStatic, "static")
	if err == nil {
		// Check if the directory has any content
		entries, err := fs.ReadDir(sub, ".")
		if err == nil && len(entries) > 0 {
			StaticFS = sub
		}
	}
}
