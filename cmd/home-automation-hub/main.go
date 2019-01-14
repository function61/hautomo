package main

import (
	"fmt"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/spf13/cobra"
	"os"
)

// replaced in build process with actual version
var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Home automation hub from function61.com",
		Version: version,
	}
	rootCmd.AddCommand(serverEntry())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func serverEntry() *cobra.Command {
	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServer(); err != nil {
				panic(err)
			}
		},
	}

	server.AddCommand(&cobra.Command{
		Use:   "lint",
		Short: "Verifies the syntax of the configuration file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			_, err := readConfigurationFile()
			if err != nil {
				panic(err)
			}
		},
	})

	server.AddCommand(&cobra.Command{
		Use:   "write-systemd-unit-file",
		Short: "Install unit file to start this on startup",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			systemdHints, err := systemdinstaller.InstallSystemdServiceFile("homeautomation", []string{"server"}, "home automation hub")
			if err != nil {
				panic(err)
			}

			fmt.Println(systemdHints)
		},
	})

	return server
}
