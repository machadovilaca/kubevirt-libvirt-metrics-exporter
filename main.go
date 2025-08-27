package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/kubevirt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/exporter"
	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/monitoring/metrics"
)

const (
	metricsPort = 4443
)

func main() {
	klog.Infoln("<- kubevirt-libvirt-metrics-exporter ->")

	err := metrics.SetupMetrics()
	if err != nil {
		klog.Fatalf("error setting up metrics: %s", err.Error())
	}

	clientset, err := newKubernetesClientset()
	if err != nil {
		klog.Fatalf("error creating kubernetes clientset: %s", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go kubevirt.NewMetricsExporter(clientset).Start(ctx)
	go exporter.NewServer(metricsPort).Start(ctx)

	<-ctx.Done()
}

func newKubernetesClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
