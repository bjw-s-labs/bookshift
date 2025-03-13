package nfs

import (
	"fmt"
	"log/slog"
	"os"
	"path"

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

func (f *NfsFile) CleanFileName() string {
	return f.nfsFile.Name
}

func (f *NfsFile) Download(dstFolder string, dstFileName string, overwriteExistingFile bool, keepFolderStructure bool, deleteSourcFile bool) error {
	// Create folder structure if required
	if keepFolderStructure {
		dstFolder = path.Join(dstFolder, f.subFolder)
	}

	safeFileName := util.SafeFileName(dstFileName)
	dstPath := path.Join(dstFolder, safeFileName)

	// Create folder structure if required
	if _, err := os.Stat(dstFolder); os.IsNotExist(err) {
		slog.Info("Creating local folder", "folder", dstFolder)
		if err := os.MkdirAll(dstFolder, os.ModeDir|0755); err != nil {
			return err
		}
	}

	slog.Info("Downloading file from NFS share", "host", f.nfsFolder.NfsClient.Host, "file", f.remotePath, "destination", dstPath)

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
	tmpFile, err := os.CreateTemp("", "bookshift-")
	if err != nil {
		os.Remove(tmpFile.Name())
		return err
	}
	defer tmpFile.Close()

	slog.Debug("Downloading to temporary file", "file", tmpFile.Name())
	writer := util.NewFileWriter(tmpFile, int64(f.nfsFile.Size), true)
	_, err = f.nfsFolder.NfsClient.Client.ReadFileAll(f.remotePath, writer)
	if err != nil {
		return err
	}

	if err := os.Rename(tmpFile.Name(), dstPath); err != nil {
		os.Remove(tmpFile.Name())
		return err
	}

	// Delete the source file if requested
	if deleteSourcFile {
		if err := f.Delete(); err != nil {
			return err
		}
	}

	slog.Info("Succesfully downloaded file", "filename", safeFileName)
	return nil
}

func (f *NfsFile) Delete() error {
	if err := f.nfsFolder.NfsClient.Client.DeleteFile(f.remotePath); err != nil {
		return fmt.Errorf("failed to delete the file %s: (%w)", f.remotePath, err)
	}
	slog.Info("Deleted file from NFS share", "host", f.nfsFolder.NfsClient.Host, "file", f.remotePath)
	return nil
}
