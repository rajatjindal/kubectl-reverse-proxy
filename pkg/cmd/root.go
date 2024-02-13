package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var configFlags = genericclioptions.NewConfigFlags(true)
var labelSelector = ""

// rootCmd represents the base command when called without any subcommands
var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "reverse-proxy",
		Short:   "Starts a reverse proxy to all pods behind a service",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			namespace := getNamespace(configFlags)
			localPort, _ := cmd.Flags().GetString("local-port")

			opts := ReverseProxyOptions{
				LabelSelector: labelSelector,
				Namespace:     namespace,
				LocalPort:     localPort,
			}

			err := StartReverseProxy(cmd.Context(), opts)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	configFlags.AddFlags(rootCmd.Flags())
	rootCmd.Flags().StringP("local-port", "p", "9090", "Local port to listen on")
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
