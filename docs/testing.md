# Test seams overview

This project includes dedicated “test seams” to make unit testing fast, deterministic, and isolated from real networks or devices. These seams are small interfaces and dial/connect hooks you can override in tests to inject fakes.

## Why seams?

The syncers (IMAP/SMB/NFS) and DBus integrations talk to external systems. Seams let tests run without those systems by swapping real connections for in-memory fakes. Use `t.Cleanup` to restore the original hooks after each test.

## SMB seams

- Low-level connection interface: `smbLowLevel` (TreeConnect, ListDirectory, RetrieveFile, etc.).
- Dial hook: `smbDial` in `pkg/syncer/smb/connection.go`.
- Public interface for higher layers: `SmbConnAPI`.
- Syncer hooks (in `pkg/syncer/smb/syncer.go`):
  - `smbConnect`, `smbDisconnect`
  - `newSmbShare`, `smbShareConnect`, `smbShareDisconnect`

Test pattern:

1. Replace `smbDial` with a fake that returns a `smbLowLevel` implementation.
2. Exercise `SmbConnection.Connect()` and pass-through methods.
3. For syncer.Run tests, override the syncer hooks to inject fakes and return controlled data/errors.

Note: When constructing a `SmbConnection` in tests, provide a non-nil password (e.g., `pw := sensitive.String("pw"); Password: &pw`) to avoid nil dereferences.

Tiny example:

```go
// smb_connection_test.go
type fakeSMB struct{}
func (f *fakeSMB) Close()                                   {}
func (f *fakeSMB) TreeConnect(string) error                  { return nil }
func (f *fakeSMB) TreeDisconnect(string) error               { return nil }
func (f *fakeSMB) ListDirectory(string, string, string) ([]smb.SharedFile, error) { return nil, nil }
func (f *fakeSMB) RetrieveFile(string, string, uint64, func([]byte) (int, error)) error { return nil }
func (f *fakeSMB) DeleteFile(string, string) error           { return nil }

func TestConnectWithFake(t *testing.T) {
  orig := smbDial
  t.Cleanup(func(){ smbDial = orig })
  smbDial = func(opts smb.Options) (smbLowLevel, error) { return &fakeSMB{}, nil }

  pw := sensitive.String("pw")
  c := &SmbConnection{Host: "h", Port: 445, Username: "u", Password: &pw}
  if err := c.Connect(); err != nil { t.Fatalf("connect: %v", err) }
  if err := c.TreeConnect("share"); err != nil { t.Fatalf("tree: %v", err) }
}
```

## NFS seams

- Low-level client interface: `nfsLowLevel` (Close, GetFileList, ReadFileAll, DeleteFile).
- Dial hook: `nfsDialLow` in `pkg/syncer/nfs/client.go`.
- Public interface for higher layers: `NfsAPI`.
- Syncer hooks (in `pkg/syncer/nfs/syncer.go`):
  - `newNfsClient`, `nfsConnect`
  - `nfsNewFolder`, `nfsFetchFiles`, `nfsDownload`

Test pattern:

1. Override `nfsDialLow` to return a fake `nfsLowLevel`.
2. Call `NfsClient.Connect()` and use the client in file/folder logic with fakes.
3. For syncer tests, replace the hooks above to simulate connect/list/download flows and error paths.

Tiny example:

```go
// nfs_client_test.go
type fakeNFS struct{}
func (f *fakeNFS) Close() {}
func (f *fakeNFS) GetFileList(string) ([]nfs4.FileInfo, error) { return nil, nil }
func (f *fakeNFS) ReadFileAll(string, io.Writer) (uint64, error) { return 0, nil }
func (f *fakeNFS) DeleteFile(string) error { return nil }

func TestNfsConnectWithFake(t *testing.T) {
  orig := nfsDialLow
  t.Cleanup(func(){ nfsDialLow = orig })
  nfsDialLow = func(ctx context.Context, server string, auth nfs4.AuthParams) (nfsLowLevel, error) {
    return &fakeNFS{}, nil
  }
  c := NewNfsClient("server", 2049)
  if err := c.Connect(0); err != nil { t.Fatalf("connect: %v", err) }
}
```

## IMAP seams

- Low-level connection interface: `imapConn` with minimal methods (`Login`, `Select`, `Logout`, `Close`) that return small `Wait` interfaces:
  - `waitErr { Wait() error }`
  - `waitSelect { Wait() (*imap.SelectData, error) }`
- Dial hook: `imapDial` in `pkg/syncer/imap/backend.go` returns `(imapConn, *imapclient.Client, error)`.
  - When the “real” client is non-nil, `ImapClient.Connect` wires `Backend` to a production `realImapOps` that calls the actual `imapclient.Client`.
- Syncer hooks (in `pkg/syncer/imap/syncer.go`):
  - `newImapClient`, `imapConnect`, `imapDisconnect`, `imapCollect`, `imapDownload`

Test pattern:

1. Override `imapDial` to return a fake `imapConn` (and optionally a non-nil real client to verify backend wiring).
2. Test `Connect/Disconnect` branches (dial/login/select errors, logout/close errors).
3. For message/attachment flows, provide a fake `ImapOps` backend implementing:
   - `UIDSearch`, `FetchOneByUID`, `StoreAddFlags`, `Expunge`
     and feed controlled message metadata/body content to cover overwrite/skip/rename/base64/deletion paths.

Tip: When creating an `ImapClient` in tests, set a password (`pw := sensitive.String("pw"); Password: &pw`) so the `Login` call can stringify it.

Tiny example:

```go
// imap_client_test.go
type fakeWait struct{ err error }
func (f fakeWait) Wait() error { return f.err }
type fakeSel struct{ err error }
func (f fakeSel) Wait() (*imap.SelectData, error) { return nil, f.err }

type fakeConn struct{}
func (fakeConn) Login(string, string) waitErr { return fakeWait{} }
func (fakeConn) Select(string, *imap.SelectOptions) waitSelect { return fakeSel{} }
func (fakeConn) Logout() waitErr { return fakeWait{} }
func (fakeConn) Close() error { return nil }

func TestImapConnectWithFake(t *testing.T) {
  orig := imapDial
  t.Cleanup(func(){ imapDial = orig })
  imapDial = func(addr string) (imapConn, *imapclient.Client, error) { return fakeConn{}, nil, nil }
  pw := sensitive.String("pw")
  ic := &ImapClient{Host: "mail", Port: 993, Username: "u", Password: &pw}
  if err := ic.Connect("INBOX"); err != nil { t.Fatalf("connect: %v", err) }
}
```

## CMD seams

Top-level CLI and run orchestration have small seams to avoid filesystem sleeps, device checks, and actual syncer execution during tests.

- In `cmd/run.go`:
  - `countFiles` wraps `util.CountFilesInFolder`.
  - `doNfs`, `doSmb`, `doImap` wrap the corresponding syncer `.Run(...)` calls.
  - `isKoboDevice`, `updateKoboLibrary` wrap Kobo detection and library update.
- In `cmd/root.go`:
  - `exit` wraps `os.Exit` and is passed into Kong via `kong.Exit(exit)`; used by `Execute()` and `VersionFlag.BeforeApply`.

Test pattern:

```go
// run_more_test.go
oldCount := countFiles
t.Cleanup(func(){ countFiles = oldCount })
countFiles = func(folder string, exts []string, rec bool) (int, error) {
  // first call: before, second call: after
  if folder == target && rec { return 1, nil }
  return 2, nil
}

oldIsKobo, oldUpd := isKoboDevice, updateKoboLibrary
t.Cleanup(func(){ isKoboDevice, updateKoboLibrary = oldIsKobo, oldUpd })
isKoboDevice = func() bool { return true }
updateKoboLibrary = func() error { return nil }

var c RunCommand
_ = c.Run(cfg, slog.Default())
```

Exit seam usage:

```go
// root_more_test.go
oldExit := exit
t.Cleanup(func(){ exit = oldExit })
var code int
exit = func(c int){ code = c }

// call VersionFlag hook directly to avoid process termination
v := VersionFlag("1.2.3")
_ = v.BeforeApply(nil, kong.Vars{"version": string(v)})
if code != 0 { t.Fatalf("expected 0") }
```

## Kobo seams

The Kobo integration simulates USB plug add/remove when NickelDbus is not present, and triggers a DBus rescan when it is.

- Overridable variables in `pkg/kobo/main.go`:
  - `usbPlugSleep` (default 10s); set to 0 in tests to skip waiting.
  - `ndbIsInstalled` wraps `nickeldbus.IsInstalled`.
  - `ndbLibraryRescan` wraps `nickeldbus.LibraryRescan`.

Tiny examples:

```go
// main_test.go
usbPlugSleep = 0

old := ndbIsInstalled
t.Cleanup(func(){ ndbIsInstalled = old })
ndbIsInstalled = func() bool { return true }

called := false
oldR := ndbLibraryRescan
t.Cleanup(func(){ ndbLibraryRescan = oldR })
ndbLibraryRescan = func(timeout int, full bool) error { called = true; return nil }

if err := UpdateLibrary(); err != nil { t.Fatal(err) }
if !called { t.Fatalf("expected rescan") }
```

## Good practices

- Always restore hooks in tests:
  ```go
  orig := imapDial
  t.Cleanup(func(){ imapDial = orig })
  imapDial = func(addr string) (imapConn, *imapclient.Client, error) { /* fake */ }
  ```
- Keep fakes narrow: implement just the interface methods the code under test needs.
- Prefer testing observable behavior (files written, errors returned, calls recorded) over internal state.
- Add small tests for error branches; they boost coverage and catch regressions in edge handling.

If you add new external integrations, follow this pattern: define a tiny interface for the calls you use, add a top-level dial/connect hook, and thread the interface through higher layers. Document the seam here and add a minimal fake in tests.
