package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/systemdinstaller"
	"github.com/function61/hautomo/pkg/ezstack/ezhub"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Home Automation hub from function61.com",
		Version: dynversion.Version,
	}

	rootCmd.AddCommand(serverEntry())
	rootCmd.AddCommand(ezhubEntrypoint())

	osutil.ExitIfError(rootCmd.Execute())
}

func serverEntry() *cobra.Command {
	install := false

	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if install {
				// TODO: make working directory by default where we are, not the binary's directory
				service := systemdinstaller.Service(
					"hautomo",
					"Home Automation",
					systemdinstaller.Args("server"),
					systemdinstaller.Docs("https://github.com/function61/hautomo", "https://function61.com/"))

				osutil.ExitIfError(systemdinstaller.Install(service))

				fmt.Println(systemdinstaller.EnableAndStartCommandHints(service))

				return
			}

			rootLogger := logex.StandardLogger()

			osutil.ExitIfError(runServer(
				osutil.CancelOnInterruptOrTerminate(rootLogger),
				rootLogger))
		},
	}

	server.Flags().BoolVarP(&install, "install", "", install, "Install Systemd unit file to start this on startup")

	server.AddCommand(&cobra.Command{
		Use:   "lint",
		Short: "Verifies the syntax of the configuration file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			_, err := readConfigurationFile()
			osutil.ExitIfError(err)
		},
	})

	return server
}

func ezhubEntrypoint() *cobra.Command {
	packetCapture := ""
	joinEnable := false
	settingsFlash := false
	install := false

	cmd := &cobra.Command{
		Use:   "ezhub",
		Short: "Runs Zigbee hub",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			rootLogger := logex.StandardLogger()

			if install {
				wd, err := os.Getwd()
				osutil.ExitIfError(err)

				service := systemdinstaller.Service(
					filepath.Base(wd),
					"Zigbee hub",
					systemdinstaller.Args("ezhub"),
					systemdinstaller.Docs("https://github.com/function61/hautomo", "https://function61.com/"))

				osutil.ExitIfError(systemdinstaller.Install(service))

				fmt.Println(systemdinstaller.EnableAndStartCommandHints(service))

				return
			}

			osutil.ExitIfError(ezhub.Run(
				osutil.CancelOnInterruptOrTerminate(rootLogger),
				joinEnable,
				packetCapture,
				settingsFlash,
				rootLogger))
		},
	}

	cmd.Flags().BoolVarP(&install, "install", "", install, "Install Systemd unit file to start this on startup")
	cmd.Flags().StringVarP(&packetCapture, "packet-capture", "", packetCapture, "Capture received UNP frames to a file")
	cmd.Flags().BoolVarP(&joinEnable, "join-enable", "", joinEnable, "Enable devices joining this network (for a short time)")
	cmd.Flags().BoolVarP(&settingsFlash, "settings-flash", "", settingsFlash, "Temporary flag to indicate flashing Zigbee radio settings")

	cmd.AddCommand(&cobra.Command{
		Use:   "new-config",
		Short: "Generate new configuration",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			osutil.ExitIfError(ezhub.GenerateConfiguration(os.Stdout))
		},
	})

	return cmd
}
