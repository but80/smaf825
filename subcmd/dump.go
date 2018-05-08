package subcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/but80/smaf825/smaf/chunk"
	"github.com/but80/smaf825/smaf/log"
	"github.com/but80/smaf825/smaf/voice"
	"github.com/urfave/cli"
)

var Dump = cli.Command{
	Name:      "dump",
	Aliases:   []string{"d"},
	Usage:     "Dumps SMAF format files (.mmf|.spf|.vma|.vm3|.vm5)",
	ArgsUsage: "<filename>",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "json, j",
			Usage: `Dumps in JSON format`,
		},
		cli.BoolFlag{
			Name:  "voice, v",
			Usage: `Dumps voice data only`,
		},
		cli.BoolFlag{
			Name:  "exclusive, x",
			Usage: `Dumps exclusives only`,
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: `Show debug messages`,
		},
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: `Suppress information messages`,
		},
		cli.BoolFlag{
			Name:  "silent, Q",
			Usage: `Do not output any messages`,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			cli.ShowCommandHelp(ctx, "dump")
			os.Exit(1)
		}
		if ctx.Bool("debug") {
			log.Level = log.LogLevel_Debug
		} else if ctx.Bool("silent") {
			log.Level = log.LogLevel_None
		} else if ctx.Bool("quiet") {
			log.Level = log.LogLevel_Warn
		}
		file := ctx.Args()[0]
		ext := ""
		i := len(file) - 4
		if 0 <= i {
			ext = strings.ToLower(file[i:])
		}
		var data fmt.Stringer
		var err error
		switch ext {
		case ".mmf", ".spf":
			fc, err := chunk.NewFileChunk(file)
			data = fc
			if err == nil && (ctx.Bool("voice") || ctx.Bool("exclusive")) {
				exclusives := fc.CollectExclusives()
				data = exclusives
				if ctx.Bool("voice") {
					data = exclusives.Voices()
				}
			}
		case ".vma":
			data, err = voice.NewVMAVoiceLib(file)
		case ".vm3":
			data, err = voice.NewVM3VoiceLib(file)
		case ".vm5":
			data, err = voice.NewVM5VoiceLib(file)
		default:
			return cli.NewExitError(fmt.Errorf("Unknown file extension"), 1)
		}
		if err != nil {
			switch data.(type) {
			case nil:
				return cli.NewExitError(err, 1)
			default:
				log.Warnf(err.Error())
			}
		}
		if ctx.Bool("json") {
			j, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return cli.NewExitError(err, 1)
			}
			fmt.Println(string(j))
		} else {
			fmt.Println(data.String())
		}
		return nil
	},
}
