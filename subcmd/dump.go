package subcmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mersenne-sister/smaf825/smaf/chunk"
	"github.com/mersenne-sister/smaf825/smaf/voice"
	"github.com/urfave/cli"
)

var Dump = cli.Command{
	Name:      "dump",
	Aliases:   []string{"d"},
	Usage:     "Dumps SMAF format files",
	ArgsUsage: " ",
	Subcommands: []cli.Command{
		{
			Name:      "mmf",
			Aliases:   []string{"m"},
			Usage:     "Dumps MMF files",
			ArgsUsage: "<filename>",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json, j",
					Usage: `Dumps in JSON format`,
				},
			},
			Action: func(ctx *cli.Context) error {
				if ctx.NArg() < 1 {
					cli.ShowCommandHelp(ctx, "mmf")
					os.Exit(1)
				}
				mmf, err := chunk.NewFileChunk(ctx.Args()[0])
				if err != nil {
					return cli.NewExitError(err, 1)
				}
				if ctx.Bool("json") {
					j, err := json.MarshalIndent(mmf, "", "  ")
					if err != nil {
						return cli.NewExitError(err, 1)
					}
					fmt.Println(string(j))
				} else {
					fmt.Println(mmf.String())
				}
				return nil
			},
		},
		{
			Name:      "voice",
			Aliases:   []string{"v"},
			Usage:     "Dumps SMAF voice files",
			ArgsUsage: "<filename>",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json, j",
					Usage: `Dumps in JSON format`,
				},
			},
			Action: func(ctx *cli.Context) error {
				if ctx.NArg() < 1 {
					cli.ShowCommandHelp(ctx, "voice")
					os.Exit(1)
				}
				file := ctx.Args()[0]
				var err error
				var lib voice.VoiceLib
				switch file[len(file)-4:] {
				case ".vma":
					lib, err = voice.NewVMAVoiceLib(file)
				case ".vm5":
					lib, err = voice.NewVM5VoiceLib(file)
				default:
					return cli.NewExitError(fmt.Errorf("Unknown file extension"), 1)
				}
				if err != nil {
					return cli.NewExitError(err, 1)
				}
				if ctx.Bool("json") {
					j, err := json.MarshalIndent(lib, "", "  ")
					if err != nil {
						return cli.NewExitError(err, 1)
					}
					fmt.Println(string(j))
				} else {
					fmt.Println(lib.String())
				}
				return nil
			},
		},
	},
}
