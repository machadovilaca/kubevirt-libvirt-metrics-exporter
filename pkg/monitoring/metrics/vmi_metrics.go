package metrics

import (
	"github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"
)

var (
	vmiMetrics = []operatormetrics.Metric{
		domainStatus,
	}

	domainStatus = operatormetrics.NewGaugeVec(
		operatormetrics.MetricOpts{
			Name: "kubevirt_vmi_domain_status",
			Help: "Status of the VMI domain (1 = up, 0 = down).",
		},
		[]string{"namespace", "name", "domain"},
	)
)

func SetVMIDomainStatus(namespace, name, domain string, status Status) {
	domainStatus.WithLabelValues(namespace, name, domain).Set(float64(status))
}
