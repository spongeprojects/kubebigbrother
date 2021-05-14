package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spongeprojects/kubebigbrother/pkg/fileorcreate"
	"github.com/spongeprojects/kubebigbrother/pkg/informers"
	"github.com/spongeprojects/kubebigbrother/pkg/log"
	"github.com/spongeprojects/kubebigbrother/pkg/watcher"
	"github.com/spongeprojects/magicconch"
	"os"
	"os/signal"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Run watch to watch specific resource's event",
	Run: func(cmd *cobra.Command, args []string) {
		informersConfigPath := viper.GetString("informers-config")
		err := fileorcreate.Ensure(informersConfigPath, InformersConfigFileTemplate)
		if err != nil {
			log.Error(errors.Wrap(err, "apply informers config template error"))
		}

		informersConfig, err := informers.LoadConfigFromFile(informersConfigPath)
		if err != nil {
			log.Fatal(errors.Wrap(err, "informers.LoadConfigFromFile error"))
		}
		watcher, err := watcher.Setup(watcher.Options{
			KubeConfig:      viper.GetString("kubeconfig"),
			InformersConfig: informersConfig,
		})
		if err != nil {
			log.Fatal(errors.Wrap(err, "setup watcher error"))
		}

		stopCh := make(chan struct{})

		// Ctrl+C
		interrupted := make(chan os.Signal)
		signal.Notify(interrupted, os.Interrupt)

		go func() {
			<-interrupted
			close(stopCh)
			<-interrupted // exit when interrupted again
			os.Exit(1)
		}()

		watcher.Start(stopCh)

		<-stopCh
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	f := watchCmd.PersistentFlags()
	f.String("kubeconfig", defaultKubeconfig, "path to kubeconfig file")
	f.String("informers-config", DefaultInformersConfigFile, "path to informers config")

	magicconch.Must(viper.BindPFlags(f))
}
