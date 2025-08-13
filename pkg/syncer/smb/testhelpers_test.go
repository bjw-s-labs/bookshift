package smb

type smbConnBase struct{}

func (s *smbConnBase) Connect() error                    { return nil }
func (s *smbConnBase) Disconnect() error                 { return nil }
func (s *smbConnBase) TreeConnect(share string) error    { return nil }
func (s *smbConnBase) TreeDisconnect(share string) error { return nil }
