package main

import (
	"fmt"
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
	Debug.Println("Start Node with port:", c.Int("port"))
	return nil
}
