package proxy

import (
	"fmt"

	"k8s.io/kubectl/pkg/cmd/logs"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// runLogsFollow is a method that runs a log stream for a specific pod in a Kubernetes cluster. It continuously streams
// the logs and prefixes each log line with the timestamp. The method takes the pod name as a parameter and a channel to
// signal when to stop streaming the logs.
//
// TODO: add support to use stopCh to stop tailing of logs.
func (p *proxy) runLogsFollow(podName string, _ <-chan struct{}) {
	logOpts := logs.NewLogsOptions(p.config.Streams, false)
	logOpts.Follow = true
	logOpts.Prefix = true
	logOpts.Timestamps = true

	lccmd := logs.NewCmdLogs(p.config.Factory, p.config.Streams)

	cmdutil.CheckErr(logOpts.Complete(p.config.Factory, lccmd, []string{fmt.Sprintf("pod/%s", podName)}))
	cmdutil.CheckErr(logOpts.Validate())
	cmdutil.CheckErr(logOpts.RunLogs())
}
