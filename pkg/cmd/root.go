package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spongeprojects/kubebigbrother/pkg/crumbs"
	"github.com/spongeprojects/kubebigbrother/pkg/fileorcreate"
	"github.com/spongeprojects/kubebigbrother/pkg/genericoptions"
	"github.com/spongeprojects/magicconch"
	"k8s.io/klog/v2"
)

var Version = "unknown"

func NewKbbCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "kbb",
	}

	cmd.AddCommand(
		NewControllerCommand(),
		NewHistoryCommand(),
		NewServeCommand(),
		NewWatchCommand(),
	)

	f := cmd.PersistentFlags()
	genericoptions.AddGlobalFlags(f)
	magicconch.Must(viper.BindPFlags(f))

	cobra.OnInitialize(initConfig)

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("KBB") // e.g. KBB_ENV -> env
	viper.AutomaticEnv()

	globalOptions := genericoptions.GetGlobalOptions()

	if globalOptions.Env == crumbs.EnvDebug {
		err := fileorcreate.Ensure(globalOptions.Config, crumbs.ConfigFileTemplate)
		if err != nil {
			klog.Error(errors.Wrap(err, "apply config template error"))
		}
	}

	if globalOptions.Config != "" {
		viper.SetConfigFile(globalOptions.Config)

		if err := viper.ReadInConfig(); err != nil {
			klog.Warning(errors.Wrapf(err, "read in config error, file: %s", viper.ConfigFileUsed()))
		} else {
			klog.Infof("using config file: %s", viper.ConfigFileUsed())
		}
	} else {
		klog.Info("config file not specified, not reading from file")
	}
}
