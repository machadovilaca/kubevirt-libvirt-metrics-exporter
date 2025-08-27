package libvirt

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"

	"github.com/libvirt/libvirt-go"
)

const maxSocketPathLength = 108

type Client struct {
	socketPath     string
	shortPath      string
	createdSymlink bool

	conn *libvirt.Connect
}

func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
	}
}

func (c *Client) createShortPath() error {
	// If the socket path is within the limit, use it directly
	if len(c.socketPath) <= maxSocketPathLength {
		c.shortPath = c.socketPath
		return nil
	}

	// Check if the original socket exists
	if _, err := os.Stat(c.socketPath); err != nil {
		return fmt.Errorf("original socket does not exist: %w", err)
	}

	// Create a shorter path using MD5 hash of the original path
	hash := fmt.Sprintf("%x", md5.Sum([]byte(c.socketPath)))
	shortPath := filepath.Join("/tmp", fmt.Sprintf("libvirt-%s.sock", hash[:16]))

	// Remove existing symlink if it exists
	_ = os.Remove(shortPath)

	// Create symbolic link
	if err := os.Symlink(c.socketPath, shortPath); err != nil {
		return fmt.Errorf("failed to create symbolic link: %w", err)
	}

	c.shortPath = shortPath
	c.createdSymlink = true
	return nil
}

func (c *Client) Connect() error {
	if err := c.createShortPath(); err != nil {
		return fmt.Errorf("failed to create short path: %w", err)
	}

	uri := fmt.Sprintf("qemu+unix:///session?socket=%s", c.shortPath)

	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		c.cleanupSymlink()
		return fmt.Errorf("failed to connect to Libvirt: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) cleanupSymlink() {
	if c.createdSymlink && c.shortPath != "" {
		_ = os.Remove(c.shortPath)
		c.createdSymlink = false
	}
}

func (c *Client) Close() error {
	var err error
	if c.conn != nil {
		_, err = c.conn.Close()
	}

	c.cleanupSymlink()

	return err
}
