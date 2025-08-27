package kubevirt

import (
	"os"
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/libvirt"
)

type MetricsExporter struct {
	clientset *kubernetes.Clientset
	nodeName  string
	sockets   map[types.NamespacedName]libvirt.Client
	mu        sync.Mutex
	done      chan struct{}
}

func NewMetricsExporter(clientset *kubernetes.Clientset) *MetricsExporter {
	return &MetricsExporter{
		clientset: clientset,
		nodeName:  os.Getenv("NODE_NAME"),
		sockets:   make(map[types.NamespacedName]libvirt.Client),
		done:      make(chan struct{}),
	}
}

func (e *MetricsExporter) Run() {
	go e.startSyncService()
	<-e.done
}
