package main

import (
	"fmt"
	"net"
	"os"

	"github.com/urfave/cli"
)

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug output",
	},
}

var Commands = []cli.Command{
	{
		Name:    "start",
		Aliases: []string{"s"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "address",
				Value: "127.0.0.1",
				Usage: "address of peer-node for P2P network",
			},
			cli.IntFlag{
				Name:  "port",
				Value: 5000,
				Usage: "port of peer-node for P2P network",
			},
		},
		Usage:  "Start node",
		Action: CmdStartNode,
	},
}

// CommandNotFound implements action when subcommand not found
func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

// CommandBefore implements action before run command
func CommandBefore(c *cli.Context) error {
	if c.GlobalBool("debug") {
		Debug.Enabled = true
	}
	return nil
}

func CmdStartNode(c *cli.Context) (err error) {
	// FIXME: Docker networks workaround
	name, err := os.Hostname()
	if err != nil {
		Debug.Printf("Error: %v\n", err)
		return err
	}
	addrs, err := net.LookupHost(name)
	if err != nil {
		Debug.Printf("Error: %v\n", err)
		return err
	}
	address := c.String("address")
	if address == "docker" {
		address = addrs[0]
	}
	port := c.Int("port")
	Info.Printf("Start Node %s:%d", address, port)
	node := NewNode(address, port)
	node.Run()
	return nil
}
