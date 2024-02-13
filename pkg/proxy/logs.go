package proxy

import (
	"fmt"

	"k8s.io/kubectl/pkg/cmd/logs"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// runLogsFollow start tail of given pod
// currently it does not support using stopCh to stop tailing of logs
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
