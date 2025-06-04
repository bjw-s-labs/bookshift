package smb

import "fmt"

var (
	ErrSmbDisconnected = fmt.Errorf("not connected to the SMB server")
)
