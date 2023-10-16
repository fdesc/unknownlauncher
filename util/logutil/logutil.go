package logutil

import (
	"runtime"
	"bufio"
	"os"
)

var reset string = "\u001b[0m"

const (
	Black = "\u001b[30m"
	Red = "\u001b[31m"
	Green = "\u001b[32m"
	Yellow = "\u001b[33m"
	Blue = "\u001b[34m"
	Magenta = "\u001b[35m"
	Cyan = "\u001b[36m"
	White = "\u001b[37m"
	BrightBlack = "\u001b[30;1m"
	BrightRed = "\u001b[31;1m"
	BrightGreen = "\u001b[32;1m"
	BrightYellow = "\u001b[33;1m"
	BrightBlue = "\u001b[34;1m"
	BrightMagenta = "\u001b[35;1m"
	BrightCyan = "\u001b[36;1m"
	BrightWhite = "\u001b[37;1m"
)

type Loglevel struct {
	Header string
	Critical bool
	Color string
}

func stdoutPrinter(msg string) {
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	writer.WriteString(msg+"\n")
}

func Info(msg string) {
	Custom(&Loglevel{Header:"INFO",Color: Blue},msg)(msg)
}

func Warn(msg string) {
	Custom(&Loglevel{Header:"WARN",Color: Yellow},msg)(msg)
}

func Error(msg string) {
	Custom(&Loglevel{Header:"ERROR",Color: Red},msg)(msg)
	os.Exit(1)
}

func Critical(msg string) {
	Custom(&Loglevel{Header:"CRITICAL",Color: BrightRed,Critical:true},msg)(msg)
}

func Custom(level *Loglevel,msg string) func(msg string) {
	if runtime.GOOS == "windows" {
		level.Color = ""
		reset = ""
	}
	if level.Color != "" {
		if level.Critical {
			return func(string) {
				stdoutPrinter(level.Color+level.Header+reset+": "+msg)
				panic(msg)
			}
		} else {
			return func(string) {
				stdoutPrinter(level.Color+level.Header+reset+": "+msg)
			}
		}
	} else if level.Color == "" {
		if level.Critical {
			return func(string) {
				stdoutPrinter(level.Header+": "+msg)
				panic(msg)
			}
		} else {
			return func(string) {
				stdoutPrinter(level.Header+": "+msg)
			}
		}
	}
	return nil
}
