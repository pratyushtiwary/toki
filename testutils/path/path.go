package test_path_utils

import (
	"path/filepath"
	"runtime"
)

func GetTestFilePath(relativePath ...string) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	pathParts := append([]string{dir}, "..")
	pathParts = append(pathParts, "resources")
	pathParts = append(pathParts, relativePath...)
	return filepath.Join(pathParts...)
}
