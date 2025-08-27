package libvirt

import "fmt"

func (c *Client) ListDomains() ([]string, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to Libvirt")
	}

	domains, err := c.conn.ListAllDomains(0)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	var domainNames []string
	for _, domain := range domains {
		name, err := domain.GetName()
		if err != nil {
			return nil, fmt.Errorf("failed to get domain name: %w", err)
		}
		domainNames = append(domainNames, name)
		domain.Free()
	}

	return domainNames, nil
}
