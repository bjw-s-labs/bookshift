package smb

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/bjw-s-labs/bookshift/pkg/util"
	"github.com/jfjallid/go-smb/smb"
)

type SmbFile struct {
	smbShareConn *SmbShareConnection
	smbFile      *smb.SharedFile

	rootFolder string
	subFolder  string
	remotePath string
}

func NewSmbFile(rootFolder string, subFolder string, file *smb.SharedFile, share *SmbShareConnection) *SmbFile {
	return &SmbFile{
		smbShareConn: share,
		smbFile:      file,

		rootFolder: rootFolder,
		subFolder:  subFolder,
		remotePath: path.Join(rootFolder, subFolder, file.Name),
	}
}

func (f *SmbFile) CleanFileName() string {
	return f.smbFile.Name
}

func (f *SmbFile) Download(dstFolder string, dstFileName string, overwriteExistingFile bool, keepFolderStructure bool, deleteSourcFile bool) error {
	// Create folder structure if required
	if keepFolderStructure {
		dstFolder = path.Join(dstFolder, f.subFolder, dstFileName)
	}

	dstPath := path.Join(dstFolder, dstFileName)

	// Create folder structure if required
	if _, err := os.Stat(dstFolder); os.IsNotExist(err) {
		slog.Info("Creating local folder", "folder", dstFolder)
		err := os.MkdirAll(dstFolder, os.ModeDir|0755)
		if err != nil {
			return err
		}
	}

	slog.Info("Downloading file from SMB share", "host", f.smbShareConn.SmbConnection.Host, "share", f.smbShareConn.Share, "file", f.remotePath, "destination", dstPath)

	// Check if the file already exists
	_, err := os.Stat(dstPath)
	if !os.IsNotExist(err) {
		if !overwriteExistingFile {
			slog.Warn("File already exists, skipping download", "file", dstPath)
			return nil
		}

		slog.Info("Overwriting existing file", "file", dstPath)
	}

	// Download the file
	tmpFile, err := os.CreateTemp("", "bsookshift-")
	if err != nil {
		os.Remove(tmpFile.Name())
		return err
	}
	defer tmpFile.Close()

	writer := util.NewFileWriter(tmpFile, int64(f.smbFile.Size), true)
	err = f.smbShareConn.SmbConnection.Connection.RetrieveFile(f.smbShareConn.Share, f.smbFile.FullPath, 0, writer.Write)
	if err != nil {
		return err
	}
	os.Rename(tmpFile.Name(), dstPath)

	// Delete the source file if requested
	if deleteSourcFile {
		err = f.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *SmbFile) Delete() error {
	err := f.smbShareConn.SmbConnection.Connection.DeleteFile(f.smbShareConn.Share, f.smbFile.FullPath)
	if err != nil {
		return fmt.Errorf("failed to delete the file (%s)", f.remotePath)
	}
	slog.Info("Deleted file from SMB share", "host", f.smbShareConn.SmbConnection.Host, "share", f.smbShareConn.Share, "file", f.remotePath)
	return nil
}
