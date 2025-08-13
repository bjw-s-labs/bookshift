// syncer_seams.go: Test seams for the NFS syncer.
//
// This file defines small interfaces and function variables that production code
// uses by default, but tests can override to substitute fakes. Keeping these
// seams in a separate file keeps the main implementation clean while preserving
// simple, package-scoped overrides in tests.
package nfs

import "time"

// test hooks (seams) for dependency injection in tests
var (
	newNfsClient  = func(host string, port int) NfsAPI { return NewNfsClient(host, port) }
	nfsConnect    = func(c NfsAPI, timeout time.Duration) error { return c.Connect(timeout) }
	nfsNewFolder  = func(folder string, conn NfsAPI) *NfsFolder { return NewNfsFolder(folder, conn) }
	nfsFetchFiles = func(f *NfsFolder, folder string, valid []string, recurse bool) ([]NfsFile, error) {
		return f.FetchFiles(folder, valid, recurse)
	}
	nfsDownload = func(nf *NfsFile, dst, name string, overwrite, keep, del bool) error {
		return nf.Download(dst, name, overwrite, keep, del)
	}
)
