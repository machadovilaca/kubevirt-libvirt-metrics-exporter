package kubevirt

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/machadovilaca/kubevirt-libvirt-metrics-exporter/pkg/libvirt"
)

const (
	socketPath = "/var/lib/kubelet/pods/%s/volumes/kubernetes.io~empty-dir/libvirt-runtime/virtqemud-sock"

	maxRetries   = 20
	initialDelay = 1 * time.Second
	maxDelay     = 2 * time.Minute
)

func (e *MetricsExporter) startSyncService(ctx context.Context) {
	opts := []informers.SharedInformerOption{
		informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
			lo.LabelSelector = "kubevirt.io=virt-launcher"
			if e.nodeName != "" {
				lo.FieldSelector = fields.OneTermEqualSelector("spec.nodeName", e.nodeName).String()
			}
		}),
	}

	factory := informers.NewSharedInformerFactoryWithOptions(e.clientset, time.Duration(0), opts...)
	podInformer := factory.Core().V1().Pods().Informer()

	if _, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    e.onPodAdd,
		DeleteFunc: e.onPodDelete,
	}); err != nil {
		klog.Error("failed to add event handler to pod informer: ", err)
		return
	}

	// Start informers (non-blocking)
	factory.Start(ctx.Done())

	// Wait for caches to sync
	if ok := cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced); !ok {
		klog.Error("failed to wait for caches to sync")
		return
	}

	klog.Infoln("sync service started, listening for virt-launcher pod events...")
	<-ctx.Done()
}

func (e *MetricsExporter) onPodAdd(obj any) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.Errorf("onPodAdd received unexpected object type: %T", obj)
		return
	}

	// Start connection attempt in a goroutine with retry logic
	go e.connectWithRetry(pod)
}

func (e *MetricsExporter) connectWithRetry(pod *corev1.Pod) {
	sockPath := fmt.Sprintf(socketPath, pod.UID)

	klog.Infof("attempting to connect to Libvirt socket for pod %s/%s at %s", pod.Namespace, pod.Name, sockPath)

	for attempt := 0; attempt < maxRetries; attempt++ {
		socket := libvirt.NewClient(sockPath)
		if err := socket.Connect(); err != nil {
			delay := time.Duration(1<<uint(attempt)) * initialDelay
			if delay > maxDelay {
				delay = maxDelay
			}

			klog.V(4).Infof("attempt %d/%d: failed to connect to Libvirt socket for pod %s/%s: %v. Retrying in %v",
				attempt+1, maxRetries, pod.Namespace, pod.Name, err, delay)

			if attempt == maxRetries-1 {
				klog.Errorf("failed to connect to Libvirt socket for pod %s/%s after %d attempts: %v",
					pod.Namespace, pod.Name, maxRetries, err)
				return
			}

			time.Sleep(delay)
			continue
		}

		// Connection successful
		klog.Infof("connected to Libvirt socket for pod %s/%s (attempt %d)", pod.Namespace, pod.Name, attempt+1)
		e.mu.Lock()
		e.vmis[types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}] = virtualMachineInstance{
			namespace:     pod.Namespace,
			name:          pod.Labels["vm.kubevirt.io/name"],
			libvirtClient: socket,
		}
		e.mu.Unlock()
		return
	}
}

func (e *MetricsExporter) onPodDelete(obj any) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.Errorf("onPodDelete received unexpected object type: %T", obj)
		return
	}

	klog.Infof("pod %s/%s deleted, closing Libvirt connection", pod.Namespace, pod.Name)

	e.mu.Lock()
	if vmi, exists := e.vmis[types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}]; exists {
		if err := vmi.libvirtClient.Close(); err != nil {
			klog.Errorf("failed to close Libvirt connection for pod %s/%s: %v", pod.Namespace, pod.Name, err)
		} else {
			klog.Infof("closed Libvirt connection for pod %s/%s", pod.Namespace, pod.Name)
		}
		delete(e.vmis, types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
	}

	e.mu.Unlock()
}
