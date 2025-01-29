package nfs

import (
	"context"
	"fmt"
	"os"

	"github.com/kha7iq/go-nfs-client/nfs4"
)

type NfsClient struct {
	Host string
	Port int

	Client *nfs4.NfsClient
}

func (c *NfsClient) Connect() error {
	ctx := context.Background()
	server := fmt.Sprintf("%s:%d", c.Host, c.Port)
	hostname, _ := os.Hostname()

	client, err := nfs4.NewNfsClient(ctx, server, nfs4.AuthParams{
		MachineName: hostname,
	})
	if err != nil {
		return err
	}

	c.Client = client
	return nil
}

func (c *NfsClient) Disconnect() error {
	c.Client.Close()
	return nil
}
