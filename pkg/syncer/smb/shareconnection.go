package smb

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
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
	slog.Debug("Initiating SMB connection", "share", s.Share)
	if err := s.SmbConnection.connection.TreeConnect(s.Share); err != nil {
		if err == smb.StatusMap[smb.StatusBadNetworkName] {
			return fmt.Errorf("share %s not found", s.Share)
		}
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (s *SmbShareConnection) Disconnect() error {
	slog.Debug("Disconnecting SMB connection", "share", s.Share)
	if err := s.SmbConnection.connection.TreeDisconnect(s.Share); err != nil {
		return err
	}
	return nil
}

func (s *SmbShareConnection) FetchFiles(folder string, validExtensions []string, recurse bool) ([]SmbFile, error) {
	allFiles, err := s.fetchAllFiles(folder, "", validExtensions, recurse)
	if err != nil {
		return nil, err
	}

	return allFiles, nil
}

func (s *SmbShareConnection) fetchAllFiles(rootFolder string, subfolder string, validExtensions []string, recurse bool) ([]SmbFile, error) {
	var allFiles []SmbFile

	if subfolder == "" {
		subfolder = rootFolder
	}

	files, err := s.SmbConnection.connection.ListDirectory(s.Share, subfolder, "*")
	if err != nil {
		if err == smb.StatusMap[smb.StatusAccessDenied] {
			return nil, err
		}
		return nil, err
	}

	for _, file := range files {
		cleanPath := strings.ReplaceAll(file.FullPath, "\\", string(os.PathSeparator))
		if file.IsDir && !file.IsJunction {
			tmpFiles, err := s.fetchAllFiles(rootFolder, cleanPath, validExtensions, recurse)
			if err != nil {
				slog.Warn("Failed to list files in directory", "directory", cleanPath, "error", err)
				continue
			}
			allFiles = append(allFiles, tmpFiles...)
		} else if !file.IsDir && !file.IsJunction {
			extension := path.Ext(cleanPath)
			if len(validExtensions) > 0 && !slices.Contains(validExtensions, extension) {
				continue
			}

			parentFolder := path.Dir(cleanPath)
			_, subFolder, _ := strings.Cut(parentFolder, rootFolder)

			allFiles = append(allFiles, *NewSmbFile(rootFolder, subFolder, &file, s))
		}
	}

	return allFiles, nil
}
