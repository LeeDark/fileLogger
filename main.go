package main

// TODO: Urfave CLI and commands
// TODO: command startNode

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "FileLogger"
	app.Version = "0.1"
	app.Usage = "Command-line API for FileLogger"

	app.Flags = GlobalFlags
	app.Commands = Commands

	app.CommandNotFound = CommandNotFound
	app.Before = CommandBefore

	app.Run(os.Args)
}
