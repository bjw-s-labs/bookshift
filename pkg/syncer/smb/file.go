package smb

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

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

func (f *SmbFile) Download(dstFolder string, dstFileName string, overwriteExistingFile bool, keepFolderStructure bool, deleteSourceFile bool) error {
	// Create folder structure if required
	if keepFolderStructure {
		dstFolder = filepath.Join(dstFolder, f.subFolder)
	}

	// Default to the remote filename when none is provided
	if dstFileName == "" {
		dstFileName = f.smbFile.Name
	}
	safeFileName := util.SafeFileName(dstFileName)
	dstPath := filepath.Join(dstFolder, safeFileName)

	// Create folder structure if required
	if _, err := os.Stat(dstFolder); os.IsNotExist(err) {
		if util.DryRun {
			slog.Info("[dry-run] Would create local folder", "folder", dstFolder)
			return nil
		}
		slog.Info("Creating local folder", "folder", dstFolder)
		if err := os.MkdirAll(dstFolder, 0755); err != nil {
			return err
		}
	}

	// Best-effort host info not available via interface; omit host here
	slog.Info("Downloading file from SMB share", "share", f.smbShareConn.Share, "file", f.remotePath, "destination", dstPath)

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
	if util.DryRun {
		slog.Info("[dry-run] Would download file", "source", f.remotePath, "destination", dstPath)
		return nil
	}
	tmpFile, err := os.CreateTemp(dstFolder, "bookshift-")
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	slog.Debug("Downloading to temporary file", "file", tmpFile.Name())
	writer := util.NewFileWriter(tmpFile, int64(f.smbFile.Size), true)
	if err := f.smbShareConn.SmbConnection.RetrieveFile(
		f.smbShareConn.Share,
		f.smbFile.FullPath,
		0,
		writer.Write,
	); err != nil {
		return err
	}

	if err := tmpFile.Sync(); err != nil { // ensure data flushed before rename
		os.Remove(tmpFile.Name())
		return err
	}
	if err := os.Rename(tmpFile.Name(), dstPath); err != nil {
		os.Remove(tmpFile.Name())
		return err
	}

	// Delete the source file if requested
	if deleteSourceFile {
		if util.DryRun {
			slog.Info("[dry-run] Would delete file from SMB share", "file", f.remotePath)
		} else {
			if err := f.Delete(); err != nil {
				return err
			}
		}
	}

	slog.Info("Successfully downloaded file", "filename", safeFileName)
	return nil
}

func (f *SmbFile) Delete() error {
	if err := f.smbShareConn.SmbConnection.DeleteFile(f.smbShareConn.Share, f.smbFile.FullPath); err != nil {
		return fmt.Errorf("failed to delete the file (%s): %w", f.remotePath, err)
	}
	slog.Info("Deleted file from SMB share", "share", f.smbShareConn.Share, "file", f.remotePath)
	return nil
}
