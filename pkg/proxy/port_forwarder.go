package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	portforwardtools "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/portforward"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func (p *proxy) runPortForward(podName string, localPort, remotePort string, stopCh <-chan struct{}) {
	// only display port-forward errors
	pwStreams := genericclioptions.IOStreams{In: p.config.Streams.In, Out: io.Discard, ErrOut: p.config.Streams.ErrOut}
	pwOpts := portforward.PortForwardOptions{
		PortForwarder: &defaultPortForwarder{
			stopCh:    stopCh,
			IOStreams: pwStreams,
		},
		Address: []string{"localhost"},
	}

	ccmd := portforward.NewCmdPortForward(p.config.Factory, pwStreams)
	podReference := fmt.Sprintf("pod/%s", podName)

	// do port-forward
	cmdutil.CheckErr(pwOpts.Complete(p.config.Factory, ccmd, []string{podReference, fmt.Sprintf("%s:%s", localPort, remotePort)}))
	cmdutil.CheckErr(pwOpts.Validate())
	cmdutil.CheckErr(pwOpts.RunPortForward())
}

type defaultPortForwarder struct {
	stopCh  <-chan struct{}
	readyCh chan struct{}
	genericiooptions.IOStreams
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts portforward.PortForwardOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforwardtools.NewOnAddresses(dialer, opts.Address, opts.Ports, f.stopCh, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}
