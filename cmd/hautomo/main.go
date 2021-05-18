package main

import (
	"fmt"
	"os"

	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/systemdinstaller"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Home Automation hub from function61.com",
		Version: dynversion.Version,
	}

	rootCmd.AddCommand(serverEntry())

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
