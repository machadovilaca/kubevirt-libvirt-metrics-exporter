package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/kubevirt"
)

func main() {
	klog.Infoln("starting kubevirt-libvirt-metrics-exporter")

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	exporter := kubevirt.NewMetricsExporter(clientset)
	exporter.Run()
}
