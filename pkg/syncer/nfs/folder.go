package nfs

import (
	"log/slog"
	"path"
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

func (s *NfsFolder) FetchFiles(folder string, recurse bool) ([]NfsFile, error) {
	allFiles, err := s.fetchAllFiles(folder, "", recurse)
	if err != nil {
		return nil, err
	}

	return allFiles, nil
}

func (s *NfsFolder) fetchAllFiles(rootFolder string, folder string, recurse bool) ([]NfsFile, error) {
	var allFiles []NfsFile

	if folder == "" {
		folder = rootFolder
	}

	files, err := s.NfsClient.Client.GetFileList(folder)
	if err != nil {
		return nil, err
	}

	slog.Debug("result", "files", files)

	for _, file := range files {
		fullPath := path.Join(folder, file.Name)
		if file.IsDir {
			tmpFiles, err := s.fetchAllFiles(rootFolder, fullPath, recurse)
			if err != nil {
				slog.Warn("Failed to list files in directory", "directory", fullPath, "error", err)
				continue
			}
			allFiles = append(allFiles, tmpFiles...)
		} else if !file.IsDir {
			parentFolder := path.Dir(fullPath)
			_, subFolder, _ := strings.Cut(parentFolder, rootFolder)

			allFiles = append(allFiles, *NewNfsFile(rootFolder, subFolder, &file, s))
		}
	}

	return allFiles, nil
}
