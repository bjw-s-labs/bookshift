package nfs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"time"

	"github.com/kha7iq/go-nfs-client/nfs4"
)

// NfsAPI is the minimal contract used by folder and file logic.
// It enables injecting a fake in tests.
type NfsAPI interface {
	Connect(timeout time.Duration) error
	Disconnect() error
	GetFileList(path string) ([]nfs4.FileInfo, error)
	ReadFileAll(path string, w io.Writer) (int, error)
	DeleteFile(path string) error
	Host() string
	Port() int
}

type NfsClient struct {
	host string
	port int

	client nfsLowLevel
}

// nfsLowLevel captures the minimal calls used from the underlying nfs4 client.
type nfsLowLevel interface {
	Close()
	GetFileList(path string) ([]nfs4.FileInfo, error)
	ReadFileAll(path string, w io.Writer) (uint64, error)
	DeleteFile(path string) error
}

// dial hook for low-level client (overridable in tests)
var nfsDialLow = func(ctx context.Context, server string, auth nfs4.AuthParams) (nfsLowLevel, error) {
	return nfs4.NewNfsClient(ctx, server, auth)
}

func NewNfsClient(host string, port int) *NfsClient {
	return &NfsClient{
		host: host,
		port: port,
	}
}

// Establishes a connection to the NFS server with a specified timeout.
func (c *NfsClient) Connect(timeout time.Duration) error {
	slog.Debug("Initiating NFS connection", "host", c.host)

	// Create a context with a timeout to prevent the connection from hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Construct the server address in the format "host:port".
	server := fmt.Sprintf("%s:%d", c.host, c.port)
	// Get the hostname of the local machine.
	hostname, _ := os.Hostname()

	// Create a new NFS client instance, passing in the context, server address, and authentication parameters.
	client, err := nfsDialLow(ctx, server, nfs4.AuthParams{
		MachineName: hostname,
	})
	if err != nil {
		return err
	}

	// Store the newly created client instance in the NfsClient struct.
	c.client = client

	return nil
}

// Disconnects from the NFS server by closing the client connection.
func (c *NfsClient) Disconnect() error {
	slog.Debug("Disconnecting NFS connection", "host", c.host)

	// Check if the client instance is not nil before attempting to close it.
	if c.client != nil {
		// Close the client connection to release system resources.
		c.client.Close()
	}
	return nil
}

func (c *NfsClient) GetFileList(path string) ([]nfs4.FileInfo, error) {
	// Check if the client instance is not nil before attempting to use it.
	if c.client != nil {
		return c.client.GetFileList(path)
	}
	return nil, fmt.Errorf("client is not connected")
}

// ReadFileAll streams a remote file to the provided writer.
func (c *NfsClient) ReadFileAll(path string, w io.Writer) (int, error) {
	if c.client == nil {
		return 0, fmt.Errorf("client is not connected")
	}
	n, err := c.client.ReadFileAll(path, w)
	// Cap to the platform's max int to avoid overflow when converting from uint64.
	if n > uint64(math.MaxInt) {
		return math.MaxInt, err
	}
	return int(n), err
}

// DeleteFile removes a remote file.
func (c *NfsClient) DeleteFile(path string) error {
	if c.client == nil {
		return fmt.Errorf("client is not connected")
	}
	return c.client.DeleteFile(path)
}

func (c *NfsClient) Host() string {
	return c.host
}

func (c *NfsClient) Port() int {
	return c.port
}
