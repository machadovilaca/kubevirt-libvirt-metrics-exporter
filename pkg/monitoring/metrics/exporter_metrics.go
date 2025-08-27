package metrics

import "github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"

type Status int

const (
	StatusRunning = 1
)

var (
	exporterMetrics = []operatormetrics.Metric{
		libvirtMetricsExporterStatus,
	}

	libvirtMetricsExporterStatus = operatormetrics.NewGauge(
		operatormetrics.MetricOpts{
			Name: "kubevirt_libvirt_metrics_exporter_status",
			Help: "Status of the KubeVirt Libvirt Metrics Exporter (1 = up, 0 = down)",
		},
	)
)

func SetLibvirtMetricsExporterStatus(status Status) {
	libvirtMetricsExporterStatus.Set(float64(status))
}
