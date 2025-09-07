package util

import (
	"os"
	"path/filepath"
	"strings"
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
			ext := strings.ToLower(filepath.Ext(fullPath))
			// Normalize configured extensions to lowercase once for comparison
			if len(validExtensions) > 0 {
				match := false
				for _, ve := range validExtensions {
					if strings.ToLower(ve) == ext {
						match = true
						break
					}
				}
				if match {
					count++
				}
			} else {
				count++
			}
		}
	}

	return count, nil
}
