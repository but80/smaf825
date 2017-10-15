package subcmd

import (
	"os"

	"github.com/but80/smaf825/sequencer"
	"github.com/but80/smaf825/smaf/chunk"
	"github.com/urfave/cli"
)

var Play = cli.Command{
	Name:      "play",
	Aliases:   []string{"p"},
	Usage:     "Plays SMAF format files (.mmf|.spf)",
	ArgsUsage: "<device> <filename>",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "state, s",
			Usage: `Show state`,
		},
		cli.IntFlag{
			Name:  "volume, v",
			Usage: `Master volume (0..63)`,
			Value: 48,
		},
		cli.IntFlag{
			Name:  "gain, g",
			Usage: `Analog gain (0..3)`,
			Value: 1,
		},
		cli.IntFlag{
			Name:  "seqvol, V",
			Usage: `SeqVol (0..31)`,
			Value: 16,
		},
		cli.IntFlag{
			Name:  "loop, l",
			Usage: `Loop count (0: infinite)`,
			Value: 1,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 2 || ctx.Int("loop") < 0 ||
			ctx.Int("volume") < 0 || 63 < ctx.Int("volume") ||
			ctx.Int("gain") < 0 || 3 < ctx.Int("gain") ||
			ctx.Int("seqvol") < 0 || 31 < ctx.Int("seqvol") {
			cli.ShowCommandHelp(ctx, "play")
			os.Exit(1)
		}
		args := ctx.Args()
		mmf, err := chunk.NewFileChunk(args[1])
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		//fmt.Println(mmf.String())
		q := sequencer.Sequencer{
			DeviceName: args[0],
			ShowState:  ctx.Bool("state"),
		}
		err = q.Play(mmf, ctx.Int("loop"), ctx.Int("volume"), ctx.Int("gain"), ctx.Int("seqvol"))
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		return nil
	},
}
