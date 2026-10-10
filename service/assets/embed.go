package assets

import (
	"embed"
	"strings"
)

//go:embed version conf.example.ini lang/*
var files embed.FS

// Asset retains the upstream helper API without generated source or build tools.
func Asset(name string) ([]byte, error) {
	return files.ReadFile(strings.TrimPrefix(name, "assets/"))
}
