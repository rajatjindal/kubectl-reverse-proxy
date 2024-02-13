package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rajatjindal/kubectl-reverse-proxy/pkg/proxy"
)

type ReverseProxyOptions struct {
	LabelSelector string
	LocalPort     string
	Namespace     string
}

func StartReverseProxy(ctx context.Context, opts ReverseProxyOptions) error {
	if opts.LabelSelector == "" {
		return fmt.Errorf("labelselector is mandatory")
	}

	var err error
	k8sclient, err := getKubernetesClientset()
	if err != nil {
		return err
	}

	stopCh := make(chan struct{})
	factory, streams := NewCommandFactory()

	config := &proxy.Config{
		K8sClient:     k8sclient,
		LabelSelector: opts.LabelSelector,
		Namespace:     opts.Namespace,
		ListenPort:    fmt.Sprintf(":%s", opts.LocalPort),
		Factory:       factory,
		Streams:       streams,
		StopCh:        stopCh,
	}

	fmt.Printf("starting reverse proxy listening on localhost:%s\n", opts.LocalPort)

	//starts in background
	proxy.Start(ctx, config)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

	close(stopCh)
	fmt.Println("stopping proxy. Press Ctrl + c again to kill immediately")

	for {
		select {
		case <-sigterm:
			return nil
		case <-time.NewTicker(2 * time.Second).C:
			return nil
		}
	}
}
