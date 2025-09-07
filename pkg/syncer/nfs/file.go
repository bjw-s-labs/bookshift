package nfs

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/bjw-s-labs/bookshift/pkg/util"
	"github.com/kha7iq/go-nfs-client/nfs4"
)

type NfsFile struct {
	nfsFolder *NfsFolder
	nfsFile   *nfs4.FileInfo

	rootFolder string
	subFolder  string
	remotePath string
}

func NewNfsFile(rootFolder string, subFolder string, file *nfs4.FileInfo, nfsFolder *NfsFolder) *NfsFile {
	return &NfsFile{
		nfsFolder: nfsFolder,
		nfsFile:   file,

		rootFolder: rootFolder,
		subFolder:  subFolder,
		remotePath: path.Join(rootFolder, subFolder, file.Name),
	}
}

func (f *NfsFile) Download(dstFolder string, dstFileName string, overwriteExistingFile bool, keepFolderStructure bool, deleteSourceFile bool) error {
	// Create folder structure if required
	if keepFolderStructure {
		dstFolder = filepath.Join(dstFolder, f.subFolder)
	}

	// If no destination filename is provided, use the remote file name by default
	if dstFileName == "" {
		dstFileName = f.nfsFile.Name
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

	slog.Info("Downloading file from NFS share", "host", f.nfsFolder.nfsClient.Host(), "file", f.remotePath, "destination", dstPath)

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

	writer := util.NewFileWriter(tmpFile, int64(f.nfsFile.Size), true)
	_, err = f.nfsFolder.nfsClient.ReadFileAll(f.remotePath, writer)
	if err != nil {
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
			slog.Info("[dry-run] Would delete file from NFS share", "file", f.remotePath)
		} else {
			if err := f.Delete(); err != nil {
				return err
			}
		}
	}

	slog.Info("Successfully downloaded file", "filename", safeFileName)
	return nil
}

func (f *NfsFile) Delete() error {
	if err := f.nfsFolder.nfsClient.DeleteFile(f.remotePath); err != nil {
		return fmt.Errorf("failed to delete the file %s: (%w)", f.remotePath, err)
	}
	slog.Info("Deleted file from NFS share", "host", f.nfsFolder.nfsClient.Host(), "file", f.remotePath)
	return nil
}
