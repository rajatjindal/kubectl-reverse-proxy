package proxy

import (
	"context"
	"sync"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type reverseproxy interface {
	Start()
	Reload(map[string]string)
	Stop()
}

type proxy struct {
	reverseproxy reverseproxy
	config       *Config

	portMap         map[string]string
	unwatchPodChMap map[string]chan struct{}
	sync.RWMutex
}

type Config struct {
	LabelSelector string
	Namespace     string
	K8sClient     kubernetes.Interface
	ListenPort    string
	Factory       cmdutil.Factory
	Streams       genericclioptions.IOStreams
	StopCh        <-chan struct{}
}

func Start(ctx context.Context, config *Config) {
	p := &proxy{
		portMap:         map[string]string{},
		unwatchPodChMap: map[string]chan struct{}{},
		config:          config,
		reverseproxy:    NewCaddyReverseProxy(config.ListenPort, config.StopCh),
	}

	go p.reverseproxy.Start()
	go p.startController(ctx)
}
