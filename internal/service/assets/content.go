// Package assets handles generated static.
// To update views - add changes to template htmls.
package assets

import (
	"embed"
	"path/filepath"
)

const (
	dir = "templates"
)

// content holds our puzzles inputs content.
//
//go:embed templates/*
var content embed.FS

// Load loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Load(name string) ([]byte, error) {
	return content.ReadFile(filepath.Clean(
		filepath.Join(dir, name)))
}

// MustLoad loads and returns the asset for the given name.
// It panics if the asset could not be found or
// could not be loaded.
func MustLoad(name string) []byte {
	res, err := Load(name)
	if err != nil {
		panic(err)
	}

	return res
}
