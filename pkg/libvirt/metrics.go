package libvirt

import (
	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/monitoring/metrics"
	"k8s.io/klog/v2"
)

func (c *Client) CollectMetrics(namespace string, name string) {
	domains, err := c.ListDomains()
	if err != nil {
		klog.Errorf("failed to list domains for VMI %s/%s: %v", namespace, name, err)
	}

	if len(domains) == 0 {
		klog.Warningf("no domains found for VMI %s/%s", namespace, name)
		return
	}

	for _, domainName := range domains {
		metrics.SetVMIDomainStatus(namespace, name, domainName, metrics.StatusRunning)
	}
}
