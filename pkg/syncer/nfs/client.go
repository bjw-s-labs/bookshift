package nfs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kha7iq/go-nfs-client/nfs4"
)

type NfsClient struct {
	host string
	port int

	client *nfs4.NfsClient
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
	client, err := nfs4.NewNfsClient(ctx, server, nfs4.AuthParams{
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

func (c *NfsClient) Host() string {
	return c.host
}

func (c *NfsClient) Port() int {
	return c.port
}
