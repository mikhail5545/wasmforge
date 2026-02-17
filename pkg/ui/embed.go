package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:out
var assets embed.FS

func Handler() fs.FS {
	// Debug
	entries, _ := assets.ReadDir("out")
	if len(entries) == 0 {
		println("CRITICAL ERROR: No assets found in embedded filesystem. This likely means the UI assets were not built or included correctly.")
	} else {
		println("Embedded UI assets found:")
		for _, entry := range entries {
			println(" - " + entry.Name())
		}
	}

	stripped, err := fs.Sub(assets, "out")
	if err != nil {
		panic(err)
	}
	return stripped
}
