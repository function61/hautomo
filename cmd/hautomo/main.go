package main

import (
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/stopper"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Home Automation hub from function61.com",
		Version: dynversion.Version,
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
			rootLogger := logex.StandardLogger()

			workers := stopper.NewManager()

			go func(logl *logex.Leveled) {
				logl.Info.Printf("got %s; stopping", <-ossignal.InterruptOrTerminate())

				workers.StopAllWorkersAndWait()
			}(logex.Levels(logex.Prefix("main", rootLogger)))

			if err := runServer(rootLogger, workers.Stopper()); err != nil {
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
			systemdHints, err := systemdinstaller.InstallSystemdServiceFile("hautomo", []string{"server"}, "Home Automation")
			if err != nil {
				panic(err)
			}

			fmt.Println(systemdHints)
		},
	})

	return server
}
