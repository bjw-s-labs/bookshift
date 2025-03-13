package nfs

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

type NfsFolder struct {
	Folder    string
	NfsClient *NfsClient
}

func NewNfsFolder(folder string, conn *NfsClient) *NfsFolder {
	return &NfsFolder{
		Folder:    folder,
		NfsClient: conn,
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

	files, err := s.NfsClient.Client.GetFileList(folder)
	if err != nil {
		return nil, err
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
			extension := path.Ext(fullPath)
			if len(validExtensions) > 0 && !slices.Contains(validExtensions, extension) {
				continue
			}

			parentFolder := path.Dir(fullPath)
			_, subFolder, _ := strings.Cut(parentFolder, rootFolder)

			allFiles = append(allFiles, *NewNfsFile(rootFolder, subFolder, &file, s))
		}
	}

	return allFiles, nil
}
