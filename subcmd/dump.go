package subcmd

import (
	"encoding/json"
	"fmt"
	"os"

	"strings"

	"github.com/but80/smaf825/smaf/chunk"
	"github.com/but80/smaf825/smaf/voice"
	"github.com/urfave/cli"
)

var Dump = cli.Command{
	Name:      "dump",
	Aliases:   []string{"d"},
	Usage:     "Dumps SMAF format files (.mmf|.spf|.vma|.vm5)",
	ArgsUsage: "<filename>",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "json, j",
			Usage: `Dumps in JSON format`,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			cli.ShowCommandHelp(ctx, "dump")
			os.Exit(1)
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
			data, err = chunk.NewFileChunk(file)
		case ".vma":
			data, err = voice.NewVMAVoiceLib(file)
		case ".vm5":
			data, err = voice.NewVM5VoiceLib(file)
		default:
			return cli.NewExitError(fmt.Errorf("Unknown file extension"), 1)
		}
		if err != nil {
			return cli.NewExitError(err, 1)
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
