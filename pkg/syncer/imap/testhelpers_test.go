package imap

import (
	"os"

	imapv2 "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// ---- Minimal wait wrappers used by fake connections ----

type fakeWaitErr struct{ err error }

func (f fakeWaitErr) Wait() error { return f.err }

type fakeWaitSelect struct{ err error }

func (f fakeWaitSelect) Wait() (*imapv2.SelectData, error) { return nil, f.err }

// ---- Fake low-level connections used by client tests ----

type fakeConn struct {
	loginErr  error
	selErr    error
	logoutErr error
	closed    bool
}

func (f *fakeConn) Login(u, p string) waitErr { return fakeWaitErr{err: f.loginErr} }
func (f *fakeConn) Select(m string, o *imapv2.SelectOptions) waitSelect {
	return fakeWaitSelect{err: f.selErr}
}
func (f *fakeConn) Logout() waitErr { return fakeWaitErr{err: f.logoutErr} }
func (f *fakeConn) Close() error    { f.closed = true; return nil }

// Variant to exercise disconnect error paths

type fakeConn2 struct{ logoutErr, closeErr error }

func (f *fakeConn2) Login(u, p string) waitErr                           { return fakeWaitErr{} }
func (f *fakeConn2) Select(m string, o *imapv2.SelectOptions) waitSelect { return fakeWaitSelect{} }
func (f *fakeConn2) Logout() waitErr                                     { return fakeWaitErr{err: f.logoutErr} }
func (f *fakeConn2) Close() error                                        { return f.closeErr }

// ---- Backend fakes used by multiple tests ----

type fakeBackend struct {
	uids       []imapv2.UID
	fetch      map[imapv2.UID]*imapclient.FetchMessageBuffer
	searchErr  error
	fetchErr   error
	storeErr   error
	expungeErr error
}

func (f *fakeBackend) UIDSearch(criteria *imapv2.SearchCriteria) ([]imapv2.UID, error) {
	return f.uids, f.searchErr
}
func (f *fakeBackend) FetchOneByUID(uid imapv2.UID, options *imapv2.FetchOptions) (*imapclient.FetchMessageBuffer, error) {
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	return f.fetch[uid], nil
}
func (f *fakeBackend) StoreAddFlags(uid imapv2.UID, flags []imapv2.Flag) error { return f.storeErr }
func (f *fakeBackend) Expunge() error                                          { return f.expungeErr }

// recordingBackend simulates the ImapOps backend and records actions.

type recordingBackend struct {
	meta   map[imapv2.UID]*imapclient.FetchMessageBuffer
	bodies map[imapv2.UID]*imapclient.FetchMessageBuffer

	stored   []imapv2.UID
	expunges int
}

func (r *recordingBackend) UIDSearch(criteria *imapv2.SearchCriteria) ([]imapv2.UID, error) {
	// Return all UIDs we have meta for
	uids := make([]imapv2.UID, 0, len(r.meta))
	for uid := range r.meta {
		uids = append(uids, uid)
	}
	return uids, nil
}

func (r *recordingBackend) FetchOneByUID(uid imapv2.UID, options *imapv2.FetchOptions) (*imapclient.FetchMessageBuffer, error) {
	if options != nil && (options.Envelope || options.BodyStructure != nil) {
		return r.meta[uid], nil
	}
	if options != nil && len(options.BodySection) > 0 {
		return r.bodies[uid], nil
	}
	return r.meta[uid], nil
}

func (r *recordingBackend) StoreAddFlags(uid imapv2.UID, flags []imapv2.Flag) error {
	r.stored = append(r.stored, uid)
	return nil
}
func (r *recordingBackend) Expunge() error { r.expunges++; return nil }

// backend that errors on BodySection fetch

type errBodyBackend struct {
	meta *imapclient.FetchMessageBuffer
}

func (e *errBodyBackend) UIDSearch(criteria *imapv2.SearchCriteria) ([]imapv2.UID, error) {
	return []imapv2.UID{1}, nil
}
func (e *errBodyBackend) FetchOneByUID(uid imapv2.UID, options *imapv2.FetchOptions) (*imapclient.FetchMessageBuffer, error) {
	if options != nil && (options.Envelope || options.BodyStructure != nil) {
		return e.meta, nil
	}
	return nil, os.ErrPermission
}
func (e *errBodyBackend) StoreAddFlags(uid imapv2.UID, flags []imapv2.Flag) error { return nil }
func (e *errBodyBackend) Expunge() error                                          { return nil }

// Builders used across tests

func buildMeta(subject, fromName, fromMailbox, fromHost, filename string, size uint32) *imapclient.FetchMessageBuffer {
	sp := &imapv2.BodyStructureSinglePart{}
	sp.Extended = &imapv2.BodyStructureSinglePartExt{
		Disposition: &imapv2.BodyStructureDisposition{Value: "attachment", Params: map[string]string{"filename": filename}},
	}
	sp.Size = size

	env := &imapv2.Envelope{Subject: subject}
	env.From = []imapv2.Address{{Name: fromName, Mailbox: fromMailbox, Host: fromHost}}

	return &imapclient.FetchMessageBuffer{
		Envelope:      env,
		BodyStructure: sp,
	}
}

func buildBody(contentBase64 string) *imapclient.FetchMessageBuffer {
	return &imapclient.FetchMessageBuffer{
		BodySection: []imapclient.FetchBodySectionBuffer{{Section: &imapv2.FetchItemBodySection{Part: []int{1}}, Bytes: []byte(contentBase64)}},
	}
}
