package imap

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// ImapOps abstracts the subset of IMAP operations used by this package; implemented by realImapOps and test fakes.
type ImapOps interface {
	UIDSearch(criteria *imap.SearchCriteria) ([]imap.UID, error)
	FetchOneByUID(uid imap.UID, options *imap.FetchOptions) (*imapclient.FetchMessageBuffer, error)
	StoreAddFlags(uid imap.UID, flags []imap.Flag) error
	Expunge() error
}

// imapConn is a minimal low-level interface used by Connect/Disconnect.
type imapConn interface {
	Login(username, password string) waitErr
	Select(mailbox string, options *imap.SelectOptions) waitSelect
	Logout() waitErr
	Close() error
}

// Minimal wait command interfaces wrapping the go-imap v2 async API.
type waitErr interface{ Wait() error }
type waitSelect interface {
	Wait() (*imap.SelectData, error)
}

// clientWrapper adapts *imapclient.Client to our minimal imapConn.
type clientWrapper struct{ *imapclient.Client }

func (c *clientWrapper) Login(u, p string) waitErr { return c.Client.Login(u, p) }
func (c *clientWrapper) Select(m string, opt *imap.SelectOptions) waitSelect {
	return c.Client.Select(m, opt)
}
func (c *clientWrapper) Logout() waitErr { return c.Client.Logout() }
func (c *clientWrapper) Close() error    { return c.Client.Close() }

// imapDial is a dial seam for tests; returns a low-level wrapper and the real client for backend wiring.
var imapDial = func(addr string) (imapConn, *imapclient.Client, error) {
	c, err := imapclient.DialTLS(addr, nil)
	if err != nil {
		return nil, nil, err
	}
	return &clientWrapper{Client: c}, c, nil
}

// realImapOps is the production implementation over imapclient.Client.
type realImapOps struct{ c *imapclient.Client }

func (r *realImapOps) UIDSearch(criteria *imap.SearchCriteria) ([]imap.UID, error) {
	data, err := r.c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, err
	}
	return data.AllUIDs(), nil
}

func (r *realImapOps) FetchOneByUID(uid imap.UID, options *imap.FetchOptions) (*imapclient.FetchMessageBuffer, error) {
	msgs, err := r.c.Fetch(imap.UIDSetNum(uid), options).Collect()
	if err != nil {
		return nil, err
	}
	if len(msgs) != 1 {
		return nil, fmt.Errorf("len(messages) = %v, want 1", len(msgs))
	}
	return msgs[0], nil
}

func (r *realImapOps) StoreAddFlags(uid imap.UID, flags []imap.Flag) error {
	_, err := r.c.Store(imap.UIDSetNum(uid), &imap.StoreFlags{Op: imap.StoreFlagsAdd, Flags: flags}, nil).Collect()
	return err
}

func (r *realImapOps) Expunge() error {
	_, err := r.c.Expunge().Collect()
	return err
}
