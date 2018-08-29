package main

import (
	"../client"
	"../types"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 6 {
		panic("usage: " + os.Args[0] + " <serverAddr> <btAddr> <r> <g> <b>")
	}

	serverAddr := os.Args[1]
	btAddr := os.Args[2]

	r, err := strconv.ParseUint(os.Args[3], 10, 8)
	if err != nil {
		panic(err)
	}
	g, err := strconv.ParseUint(os.Args[4], 10, 8)
	if err != nil {
		panic(err)
	}
	b, err := strconv.ParseUint(os.Args[5], 10, 8)
	if err != nil {
		panic(err)
	}

	req := types.LightRequestColor(btAddr, uint8(r), uint8(g), uint8(b))

	if err := client.SendRequest(serverAddr, req); err != nil {
		panic(err)
	}

	log.Printf("job complete")
}
