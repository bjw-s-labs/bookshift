package nfs

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

type NfsFolder struct {
	Folder string

	nfsClient NfsAPI
}

func NewNfsFolder(folder string, conn NfsAPI) *NfsFolder {
	return &NfsFolder{
		Folder:    folder,
		nfsClient: conn,
	}
}

func (s *NfsFolder) FetchFiles(folder string, validExtensions []string, recurse bool) ([]NfsFile, error) {
	allFiles, err := s.fetchAllFiles(folder, "", validExtensions, recurse)
	if err != nil {
		return nil, err
	}

	return allFiles, nil
}

func (s *NfsFolder) fetchAllFiles(rootFolder string, folder string, validExtensions []string, recurse bool) ([]NfsFile, error) {
	var allFiles []NfsFile

	if folder == "" {
		folder = rootFolder
	}

	files, err := s.nfsClient.GetFileList(folder)
	if err != nil {
		return nil, err
	}

	// Build a lower-cased set of valid extensions for case-insensitive match
	lowerExts := make([]string, 0, len(validExtensions))
	for _, e := range validExtensions {
		lowerExts = append(lowerExts, strings.ToLower(e))
	}

	for _, file := range files {
		fullPath := path.Join(folder, file.Name)
		if recurse && file.IsDir {
			tmpFiles, err := s.fetchAllFiles(rootFolder, fullPath, validExtensions, recurse)
			if err != nil {
				return nil, fmt.Errorf("failed to list files in directory %s: %w", fullPath, err)
			}
			allFiles = append(allFiles, tmpFiles...)
		} else if !file.IsDir {
			extension := strings.ToLower(path.Ext(fullPath))
			if len(lowerExts) > 0 && !slices.Contains(lowerExts, extension) {
				continue
			}

			parentFolder := path.Dir(fullPath)
			_, subFolder, _ := strings.Cut(parentFolder, rootFolder)

			allFiles = append(allFiles, *NewNfsFile(rootFolder, subFolder, &file, s))
		}
	}

	return allFiles, nil
}
