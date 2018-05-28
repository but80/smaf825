package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/but80/smaf825/sequencer"
	"github.com/but80/smaf825/serial"
	"github.com/golang/protobuf/proto"
	"github.com/urfave/cli"
	"gopkg.in/but80/go-smaf.v1/chunk"
	"gopkg.in/but80/go-smaf.v1/log"
	"gopkg.in/but80/go-smaf.v1/voice"
)

var version string

func init() {
	if version == "" {
		version = "unknown"
	}
}

var playCmd = cli.Command{
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

var dumpCmd = cli.Command{
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
			Name:  "protobuf, p",
			Usage: `Dumps in protobuf`,
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
		} else if ctx.Bool("protobuf") {
			switch d := data.(type) {
			case *voice.VM5VoiceLib:
				b, err := proto.Marshal(d.ToPB())
				if err != nil {
					return cli.NewExitError(err, 1)
				}
				fmt.Print(string(b))
			default:
				return cli.NewExitError(fmt.Errorf("Protobuf conversion for %s is not supported", ext), 1)
			}
		} else {
			fmt.Println(data.String())
		}
		return nil
	},
}

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
		dumpCmd,
		playCmd,
	}

	app.Action = func(ctx *cli.Context) error {
		cli.ShowAppHelp(ctx)
		return nil
	}

	app.Run(os.Args)
}
