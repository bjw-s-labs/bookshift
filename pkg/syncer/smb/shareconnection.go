package smb

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/jfjallid/go-smb/smb"
)

type SmbShareConnection struct {
	Share         string
	SmbConnection *SmbConnection
}

func NewSmbShareConnection(share string, conn *SmbConnection) *SmbShareConnection {
	return &SmbShareConnection{
		Share:         share,
		SmbConnection: conn,
	}
}

func (s *SmbShareConnection) Connect() error {
	err := s.SmbConnection.Connection.TreeConnect(s.Share)
	if err != nil {
		if err == smb.StatusMap[smb.StatusBadNetworkName] {
			return fmt.Errorf("share %s not found", s.Share)
		}
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (s *SmbShareConnection) Disconnect() error {
	err := s.SmbConnection.Connection.TreeDisconnect(s.Share)
	if err != nil {
		return err
	}
	return nil
}

func (s *SmbShareConnection) FetchFiles(folder string, recurse bool) ([]SmbFile, error) {
	allFiles, err := s.fetchAllFiles(folder, "", recurse)
	if err != nil {
		return nil, err
	}

	return allFiles, nil
}

func (s *SmbShareConnection) fetchAllFiles(rootFolder string, subfolder string, recurse bool) ([]SmbFile, error) {
	var allFiles []SmbFile

	if subfolder == "" {
		subfolder = rootFolder
	}

	files, err := s.SmbConnection.Connection.ListDirectory(s.Share, subfolder, "*")

	if err != nil {
		if err == smb.StatusMap[smb.StatusAccessDenied] {
			return nil, err
		}
		return nil, err
	}

	for _, file := range files {
		cleanPath := strings.ReplaceAll(file.FullPath, "\\", string(os.PathSeparator))
		if file.IsDir && !file.IsJunction {
			tmpFiles, err := s.fetchAllFiles(rootFolder, cleanPath, recurse)
			if err != nil {
				slog.Warn("Failed to list files in directory", "directory", cleanPath, "error", err)
				continue
			}
			allFiles = append(allFiles, tmpFiles...)
		} else if !file.IsDir && !file.IsJunction {
			parentFolder := path.Dir(cleanPath)
			_, subFolder, _ := strings.Cut(parentFolder, rootFolder)

			allFiles = append(allFiles, *NewSmbFile(rootFolder, subFolder, &file, s))
		}
	}

	return allFiles, nil
}
