package util

import (
	"os"
	"path/filepath"
)

// resolvePath returns an absolute path, preferring cwd then repo root for relative inputs.
func resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	cwd := filepath.Join(CwdPath, path)
	root := filepath.Join(RootPath, path)
	if exists(cwd) {
		return filepath.Clean(cwd)
	}
	if exists(root) {
		return filepath.Clean(root)
	}
	return filepath.Clean(cwd) // fallback
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
