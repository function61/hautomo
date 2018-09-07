package happylightsserver

import (
	"github.com/function61/home-automation-hub/util/systemdinstaller"
	"github.com/spf13/cobra"
)

func Entrypoint() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Happylights server daemon",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	serverCmd.AddCommand(&cobra.Command{
		Use:   "write-systemd-unit-file",
		Short: "Install unit file to start this on startup",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := systemdinstaller.InstallSystemdServiceFile("happylights", []string{"happylights", "server"}, "Happylights RGB lightstrip daemon"); err != nil {
				panic(err)
			}
		},
	})

	happylightsCmd := &cobra.Command{
		Use:   "happylights",
		Short: "Happylights RGB lightstrip support",
		Args:  cobra.NoArgs,
	}
	happylightsCmd.AddCommand(serverCmd)
	return happylightsCmd
}
