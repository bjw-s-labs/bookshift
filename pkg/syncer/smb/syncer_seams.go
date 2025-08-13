// syncer_seams.go: Test seams for the SMB syncer.
//
// This file defines small function variables that production code uses by default,
// but tests can override to substitute fakes. Keeping these seams in a separate
// file keeps the main implementation clean while preserving simple overrides in tests.
package smb

// test hooks
var (
	smbConnect    = func(c *SmbConnection) error { return c.Connect() }
	smbDisconnect = func(c *SmbConnection) error { return c.Disconnect() }
	newSmbShare   = func(share string, conn SmbConnAPI) *SmbShareConnection {
		return NewSmbShareConnection(share, conn)
	}
	smbShareConnect    = func(s *SmbShareConnection) error { return s.Connect() }
	smbShareDisconnect = func(s *SmbShareConnection) error { return s.Disconnect() }
)
