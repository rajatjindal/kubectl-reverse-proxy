package proxy

import (
	"context"
	"fmt"
	"net"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (p *proxy) startController(ctx context.Context) {
	options := metav1.ListOptions{
		LabelSelector: p.config.LabelSelector,
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(_ metav1.ListOptions) (runtime.Object, error) {
				return p.config.K8sClient.DiscoveryV1().EndpointSlices(p.config.Namespace).List(ctx, options)
			},
			WatchFunc: func(_ metav1.ListOptions) (watch.Interface, error) {
				return p.config.K8sClient.DiscoveryV1().EndpointSlices(p.config.Namespace).Watch(ctx, options)
			},
		},
		&discoveryv1.EndpointSlice{},
		0,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			p.updateBackends(informer.GetStore().List())
		},
		UpdateFunc: func(old, newobj interface{}) {
			p.updateBackends(informer.GetIndexer().List())
		},
		DeleteFunc: func(obj interface{}) {
			p.updateBackends(informer.GetIndexer().List())
		},
	})

	informer.Run(p.config.StopCh)
}

func (p *proxy) updateBackends(epsList []interface{}) {
	p.Lock()
	defer p.Unlock()

	activePodMap := map[string]struct{}{}

	for _, epsI := range epsList {
		eps, ok := epsI.(*discoveryv1.EndpointSlice)
		if !ok {
			fmt.Printf("err: found an item of type %T, expected *discoveryv1.EndpointSlice\n", epsI)
			continue
		}

		// ensure all active pods are configured as backends
		for _, ep := range eps.Endpoints {
			if ep.TargetRef == nil {
				continue
			}

			if ep.Conditions.Serving != nil && !*ep.Conditions.Serving {
				continue
			}

			podName := ep.TargetRef.Name
			activePodMap[podName] = struct{}{}

			// this pod is already configured as backend
			if _, exists := p.portMap[podName]; exists {
				continue
			}

			localPort := getRandomPort()
			unwatchPodCh := make(chan struct{})

			remotePort := ""
			if len(eps.Ports) > 0 {
				if eps.Ports[0].Name != nil && *eps.Ports[0].Name != "" {
					remotePort = *eps.Ports[0].Name
				} else {
					remotePort = fmt.Sprintf("%d", *eps.Ports[0].Port)
				}
			}

			go p.runPortForward(podName, localPort, remotePort, unwatchPodCh)
			go p.runLogsFollow(podName, unwatchPodCh)

			p.portMap[podName] = localPort
			p.unwatchPodChMap[podName] = unwatchPodCh
		}

		// remove pods no longer part of endpoint slice
		for podName := range p.portMap {
			_, exists := activePodMap[podName]

			// this pod is currently active
			if exists {
				continue
			}

			// this will stop the port-forward command goroutine
			close(p.unwatchPodChMap[podName])

			// delete from known portMap
			delete(p.portMap, podName)
		}

		// reload caddy config
		p.reverseproxy.Reload(p.portMap)
	}
}

func getRandomPort() string {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)
}
