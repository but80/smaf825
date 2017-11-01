package log

import (
	"fmt"
	"os"

	"strings"

	"github.com/fatih/color"
)

type LogLevel int

const (
	LogLevel_None = iota
	LogLevel_Warn
	LogLevel_Info
	LogLevel_Debug
)

var Level LogLevel = LogLevel_Info

var cyan = color.New(color.FgCyan)
var yellow = color.New(color.FgYellow)

func Warnf(f string, args ...interface{}) {
	if LogLevel_Warn <= Level {
		yellow.Fprintf(os.Stderr, "[WARNING] "+f+"\n", args...)
	}
}

func Infof(f string, args ...interface{}) {
	if LogLevel_Info <= Level {
		fmt.Fprintf(os.Stderr, f+"\n", args...)
	}
}

var indent = 0

func Debugf(f string, args ...interface{}) {
	if LogLevel_Debug <= Level {
		cyan.Fprintf(os.Stderr, strings.Repeat("  ", indent)+f+"\n", args...)
	}
}

func Enter() {
	indent++
}

func Leave() {
	indent--
}
