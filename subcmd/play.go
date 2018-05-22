package subcmd

import (
	"os"

	"github.com/but80/smaf825/sequencer"
	"github.com/but80/smaf825/serial"
	"github.com/urfave/cli"
	"gopkg.in/but80/go-smaf.v1/chunk"
	"gopkg.in/but80/go-smaf.v1/log"
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
		cli.IntFlag{
			Name:  "baudrate, r",
			Usage: `Baud rate ` + serial.BaudRateList(),
			Value: 57600,
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
		if ctx.NArg() < 2 || ctx.Int("loop") < 0 ||
			ctx.Int("volume") < 0 || 63 < ctx.Int("volume") ||
			ctx.Int("gain") < 0 || 3 < ctx.Int("gain") ||
			ctx.Int("seqvol") < 0 || 31 < ctx.Int("seqvol") ||
			!serial.IsValidBaudRate(ctx.Int("baudrate")) {
			cli.ShowCommandHelp(ctx, "play")
			os.Exit(1)
		}
		if ctx.Bool("debug") {
			log.Level = log.LogLevel_Debug
		} else if ctx.Bool("silent") {
			log.Level = log.LogLevel_None
		} else if ctx.Bool("quiet") {
			log.Level = log.LogLevel_Warn
		}
		args := ctx.Args()
		mmf, err := chunk.NewFileChunk(args[1])
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		q := sequencer.Sequencer{
			DeviceName: args[0],
			ShowState:  ctx.Bool("state"),
		}
		opts := &sequencer.SequencerOptions{
			Loop:     ctx.Int("loop"),
			Volume:   ctx.Int("volume"),
			Gain:     ctx.Int("gain"),
			SeqVol:   ctx.Int("seqvol"),
			BaudRate: ctx.Int("baudrate"),
		}
		err = q.Play(mmf, opts)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		return nil
	},
}
