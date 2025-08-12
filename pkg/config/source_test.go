package config

import "testing"

// TestSourceUnmarshal_Smb ensures SMB source config selects the right concrete type.
func TestSourceUnmarshal_Smb(t *testing.T) {
	y := []byte("type: smb\nconfig:\n  host: h\n  share: s\n  folder: f\n")
	var s Source
	if err := s.UnmarshalYAML(y); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if s.Type != "smb" {
		t.Fatalf("got %q", s.Type)
	}
	if _, ok := s.Config.(*SmbNetworkShareConfig); !ok {
		t.Fatalf("wrong type: %T", s.Config)
	}
}

// TestSourceUnmarshal_Invalid verifies an unknown type triggers an error.
func TestSourceUnmarshal_Invalid(t *testing.T) {
	y := []byte("type: nope\nconfig: {}\n")
	var s Source
	if err := s.UnmarshalYAML(y); err == nil {
		t.Fatalf("expected error")
	}
}

// TestSourceUnmarshal_Nfs ensures NFS source config selects the correct type.
func TestSourceUnmarshal_Nfs(t *testing.T) {
	y := []byte("type: nfs\nconfig:\n  host: h\n  folder: f\n")
	var s Source
	if err := s.UnmarshalYAML(y); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if _, ok := s.Config.(*NfsNetworkShareConfig); !ok {
		t.Fatalf("wrong type: %T", s.Config)
	}
}

// TestSourceUnmarshal_Imap ensures IMAP source config selects the correct type.
func TestSourceUnmarshal_Imap(t *testing.T) {
	y := []byte("type: imap\nconfig:\n  host: h\n  mailbox: INBOX\n  filter_field: subject\n  filter_value: x\n")
	var s Source
	if err := s.UnmarshalYAML(y); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if _, ok := s.Config.(*ImapConfig); !ok {
		t.Fatalf("wrong type: %T", s.Config)
	}
}

// TestSourceUnmarshal_MissingType checks for failure when the type is omitted.
func TestSourceUnmarshal_MissingType(t *testing.T) {
	y := []byte("config: {}\n")
	var s Source
	if err := s.UnmarshalYAML(y); err == nil {
		t.Fatalf("expected error for missing type")
	}
}
