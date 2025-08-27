package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/monitoring/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

type Server struct {
	server *http.Server
}

func NewServer(port int) *Server {
	mux := http.NewServeMux()

	s := &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	s.setupRoutes(mux)

	return s
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readyHandler)
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		klog.Errorf("failed to write health check response: %v", err)
	}
}

func (s *Server) readyHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Ready")); err != nil {
		klog.Errorf("failed to write ready check response: %v", err)
	}
}

func (s *Server) Start(ctx context.Context) {
	klog.Infof("starting metrics server on %s...", s.server.Addr)
	metrics.SetLibvirtMetricsExporterStatus(metrics.StatusRunning)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			klog.Errorf("failed to start metrics server: %v", err)
		}
	}()

	// Wait for signal
	<-ctx.Done()

	klog.Info("shutting down metrics server...")
	_ = s.server.Shutdown(context.Background())
}
