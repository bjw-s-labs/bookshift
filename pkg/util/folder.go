package util

import (
	"os"
	"path"
	"path/filepath"
	"slices"
)

// CountFilesInFolder returns the number of files in a folder
func CountFilesInFolder(folder string, validExtensions []string, recurse bool) (int, error) {
	var count int

	files, err := os.ReadDir(folder)
	if err != nil {
		return 0, err
	}

	for _, file := range files {
		fullPath := filepath.Join(folder, file.Name())

		// If the file is a directory and recursion is enabled, process it
		if file.IsDir() && recurse {
			subCount, err := CountFilesInFolder(fullPath, validExtensions, recurse)
			if err != nil {
				return 0, err
			}
			count += subCount
		} else if !file.IsDir() {
			extension := path.Ext(fullPath)
			if len(validExtensions) > 0 && slices.Contains(validExtensions, extension) {
				count++
			}
		}
	}

	return count, nil
}
