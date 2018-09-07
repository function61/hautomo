package happylightsclientcli

import (
	"github.com/function61/home-automation-hub/libraries/happylights/client"
	"github.com/function61/home-automation-hub/libraries/happylights/types"
	"github.com/spf13/cobra"
	"strconv"
)

func BindEntrypoint(rootEntrypoint *cobra.Command) *cobra.Command {
	rootEntrypoint.AddCommand(&cobra.Command{
		Use:   "send [serverAddr] [btAddr] [red] [green] [blue]",
		Short: "Send command to local or remote happylights daemon",
		Args:  cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			serverAddr := args[0]
			btAddr := args[1]

			red, err := strconv.ParseUint(args[2], 10, 8)
			if err != nil {
				panic(err)
			}
			green, err := strconv.ParseUint(args[3], 10, 8)
			if err != nil {
				panic(err)
			}
			blue, err := strconv.ParseUint(args[4], 10, 8)
			if err != nil {
				panic(err)
			}

			req := types.LightRequestColor(btAddr, uint8(red), uint8(green), uint8(blue))

			if err := client.SendRequest(serverAddr, req); err != nil {
				panic(err)
			}
		},
	})

	return rootEntrypoint
}
