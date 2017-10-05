package main

import (
	"github.com/urfave/cli"

	"os"

	"github.com/mersenne-sister/smaf825/subcmd"
)

var version = "0.0.1"

func main() {
	app := cli.NewApp()
	//app.EnableBashCompletion = true
	app.Name = "smaf825"
	app.Version = version
	app.Usage = "Plays SMAF format files on YMF825 board"
	app.Authors = []cli.Author{
		{
			Name:  "but80",
			Email: "mersenne.sister@gmail.com",
		},
	}
	app.HelpName = "smaf825"

	app.Commands = []cli.Command{
		subcmd.Dump,
		subcmd.Play,
	}

	app.Action = func(ctx *cli.Context) error {
		cli.ShowAppHelp(ctx)
		return nil
	}

	app.Run(os.Args)
}
