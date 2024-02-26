package cmd

import (
	"context"
	"fmt"

	"github.com/rajatjindal/kubectl-reverse-proxy/pkg/proxy"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type ReverseProxyOptions struct {
	K8sClient     kubernetes.Interface
	IOStreams     genericclioptions.IOStreams
	Factory       cmdutil.Factory
	Name          string
	LabelSelector string
	AdminPort     string
	LocalPort     string
	Namespace     string
	StopCh        chan struct{}
}

func StartReverseProxy(ctx context.Context, opts ReverseProxyOptions) error {
	if opts.Name == "" && opts.LabelSelector == "" {
		return fmt.Errorf("one of name or label selector is required")
	}

	if opts.Name != "" && opts.LabelSelector != "" {
		return fmt.Errorf("only one of name or label selector can be provided")
	}

	if opts.Name != "" {
		opts.LabelSelector = fmt.Sprintf("kubernetes.io/service-name=%s", opts.Name)
	}

	config := &proxy.Config{
		K8sClient:     opts.K8sClient,
		LabelSelector: opts.LabelSelector,
		Namespace:     opts.Namespace,
		ListenPort:    opts.LocalPort,
		AdminPort:     opts.AdminPort,
		Factory:       opts.Factory,
		Streams:       opts.IOStreams,
		StopCh:        opts.StopCh,
	}

	fmt.Printf("starting reverse proxy listening on localhost:%s\n", opts.LocalPort)

	//starts in background
	proxy.Start(ctx, config)

	return nil
}
