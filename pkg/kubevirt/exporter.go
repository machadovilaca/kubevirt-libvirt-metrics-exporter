package kubevirt

import (
	"context"
	"os"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/libvirt"
)

type MetricsExporter struct {
	clientset *kubernetes.Clientset
	nodeName  string
	vmis      map[types.NamespacedName]virtualMachineInstance
	mu        sync.Mutex
}

type virtualMachineInstance struct {
	namespace     string
	name          string
	libvirtClient *libvirt.Client
}

func NewMetricsExporter(clientset *kubernetes.Clientset) *MetricsExporter {
	return &MetricsExporter{
		clientset: clientset,
		nodeName:  os.Getenv("NODE_NAME"),
		vmis:      make(map[types.NamespacedName]virtualMachineInstance),
	}
}

func (e *MetricsExporter) Start(ctx context.Context) {
	go e.startSyncService(ctx)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	e.collect()
	for {
		select {
		case <-ticker.C:
			e.collect()
		case <-ctx.Done():
			return
		}
	}
}

func (e *MetricsExporter) collect() {
	klog.Infoln("collecting metrics from all VM instances...")

	e.mu.Lock()
	defer e.mu.Unlock()

	for _, vmi := range e.vmis {
		klog.Infof("collecting VM instance %s in namespace %s", vmi.name, vmi.namespace)

		if vmi.libvirtClient == nil {
			continue
		}

		vmi.libvirtClient.CollectMetrics(vmi.namespace, vmi.name)
	}

	klog.Infoln("metrics collection completed")
}
