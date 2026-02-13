package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:out
var assets embed.FS

func Handler() fs.FS {
	stripped, err := fs.Sub(assets, "out")
	if err != nil {
		panic(err)
	}
	return stripped
}
