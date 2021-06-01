package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spongeprojects/kubebigbrother/pkg/cmd/controller"
	"github.com/spongeprojects/kubebigbrother/pkg/cmd/genericoptions"
	"github.com/spongeprojects/kubebigbrother/pkg/crumbs"
	"github.com/spongeprojects/kubebigbrother/pkg/informers"
	"github.com/spongeprojects/kubebigbrother/pkg/utils/signals"
	"github.com/spongeprojects/kubebigbrother/staging/fileorcreate"
	"github.com/spongeprojects/magicconch"
	"k8s.io/klog/v2"
)

type controllerOptions struct {
	GlobalOptions     *genericoptions.GlobalOptions
	DatabaseOptions   *genericoptions.DatabaseOptions
	InformersOptions  *genericoptions.InformersOptions
	KubeconfigOptions *genericoptions.KubeconfigOptions
}

func getControllerOptions() *controllerOptions {
	o := &controllerOptions{
		GlobalOptions:     genericoptions.GetGlobalOptions(),
		DatabaseOptions:   genericoptions.GetDatabaseOptions(),
		InformersOptions:  genericoptions.GetInformersOptions(),
		KubeconfigOptions: genericoptions.GetKubeconfigOptions(),
	}
	return o
}

func newControllerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Run controller, watch events and persistent into database (only one instance should be running)",
		Run: func(cmd *cobra.Command, args []string) {
			o := getControllerOptions()

			informersConfigPath := o.InformersOptions.InformersConfig

			if o.GlobalOptions.IsDebugging() {
				err := fileorcreate.Ensure(informersConfigPath, crumbs.InformersConfigFileTemplate)
				if err != nil {
					klog.Error(errors.Wrap(err, "apply informers config template error"))
				}
			}

			informersConfig, err := informers.LoadConfigFromFile(informersConfigPath)
			if err != nil {
				klog.Exit(errors.Wrap(err, "informers.LoadConfigFromFile error"))
			}

			c, err := controller.Setup(controller.Config{
				DBDialect:       o.DatabaseOptions.DBDialect,
				DBArgs:          o.DatabaseOptions.DBArgs,
				KubeConfig:      o.KubeconfigOptions.Kubeconfig,
				InformersConfig: informersConfig,
			})
			if err != nil {
				klog.Exit(errors.Wrap(err, "setup controller error"))
			}

			stopCh := signals.SetupSignalHandler()

			if err := c.Start(stopCh); err != nil {
				klog.Exit(errors.Wrap(err, "start controller error"))
			}
			defer c.Shutdown()

			<-stopCh
		},
	}

	f := cmd.PersistentFlags()
	genericoptions.AddDatabaseFlags(f)
	genericoptions.AddInformersFlags(f)
	genericoptions.AddKubeconfigFlags(f)
	magicconch.Must(viper.BindPFlags(f))

	return cmd
}
