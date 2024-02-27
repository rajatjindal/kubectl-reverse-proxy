package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rajatjindal/kubectl-reverse-proxy/pkg/dashboard"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Version is set during build time
var Version = "unknown"

var configFlags = genericclioptions.NewConfigFlags(true)
var labelSelector = ""

// rootCmd represents the base command when called without any subcommands
var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "reverse-proxy [service name]",
		Short:   "Starts a reverse proxy to all pods behind a service",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			var name string
			if len(args) > 0 {
				name = args[0]
			}

			namespace := getNamespace(configFlags)
			localPort, err := cmd.Flags().GetString("local-port")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			adminPort, err := cmd.Flags().GetString("admin-port")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			metrics, err := cmd.Flags().GetBool("metrics")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			k8sclient, err := getKubernetesClientset()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			factory, streams := NewCommandFactory()
			stopCh := make(chan struct{})
			opts := ReverseProxyOptions{
				Name:          name,
				Namespace:     namespace,
				LabelSelector: labelSelector,
				LocalPort:     localPort,
				AdminPort:     adminPort,
				Factory:       factory,
				IOStreams:     streams,
				K8sClient:     k8sclient,
				StopCh:        stopCh,
				EnableMetrics: metrics,
			}

			err = StartReverseProxy(cmd.Context(), opts)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			dash := dashboard.New(9092, fmt.Sprintf("http://localhost:%s/metrics", adminPort))
			dash.Start()

			sigterm := make(chan os.Signal, 1)
			signal.Notify(sigterm, syscall.SIGTERM)
			signal.Notify(sigterm, syscall.SIGINT)
			<-sigterm

			close(stopCh)
			fmt.Println("stopping proxy. Press Ctrl + c again to kill immediately")

			for {
				select {
				case <-sigterm:
					return
				case <-time.NewTicker(2 * time.Second).C:
					return
				}
			}
		},
	}

	configFlags.AddFlags(rootCmd.Flags())
	rootCmd.Flags().StringP("local-port", "p", "9090", "Local port to listen on")
	rootCmd.Flags().StringP("admin-port", "a", "2019", "admin port for reverse proxy")
	rootCmd.Flags().BoolP("metrics", "m", false, "enable middleware metrics for reverse proxy")
	cmdutil.AddLabelSelectorFlagVar(rootCmd, &labelSelector)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// getNamespace takes a set of kubectl flag values and returns the namespace we should be operating in
func getNamespace(flags *genericclioptions.ConfigFlags) string {
	namespace, _, err := flags.ToRawKubeConfigLoader().Namespace()
	if err != nil || len(namespace) == 0 {
		namespace = "default"
	}

	return namespace
}
