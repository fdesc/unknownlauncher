package logutil

import (
	"path/filepath"
	"runtime"
	"bufio"
	"time"
	"os"
)

var CurrentLogData string
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

type Log struct {
	Data string
	Date time.Time 
}

type Loglevel struct {
	Header string
	Critical bool
	Color string
}

func stdoutPrinter(msg string) {
	CurrentLogData = CurrentLogData + "\n" + msg
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	writer.WriteString(msg+"\n")
}

func Save(path string,logtime time.Time) string {
	var file *os.File
	var err error
	fileName := "launcher_"+logtime.Format("2006-01-02")+".log"
	file,err = os.Create(filepath.Join(path,"logs","launcher",fileName))
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Join(path,"logs","launcher"),os.ModePerm)
			if err != nil {
				Error("Failed to create directories for log path",err)
				return ""
			}
			file,err = os.Create(filepath.Join(path,"logs","launcher",fileName))
		} else {
			Error("Failed to create file for log",err)
			return ""
		}
	}
	defer file.Close()
	_,err = file.WriteString(CurrentLogData)
	if err != nil { Error("Failed to write log",err) }
	return filepath.Join(path,"logs","launcher",fileName)
}

func Info(msg string) {
	Custom(&Loglevel{Header:"INFO",Color: Blue},msg)(msg)
}

func Warn(msg string) {
	Custom(&Loglevel{Header:"WARN",Color: Yellow},msg)(msg)
}

func Error(msg string,err error) {
	Custom(&Loglevel{Header:"ERROR",Color: Red},msg+" "+err.Error())(msg+" "+err.Error())
}

func Critical(msg string,err error) {
	Custom(&Loglevel{Header:"CRITICAL",Color: BrightRed,Critical:true},msg+" "+err.Error())(msg+" "+err.Error())
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
				os.Exit(1)
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
				os.Exit(1)
			}
		} else {
			return func(string) {
				stdoutPrinter(level.Header+": "+msg)
			}
		}
	}
	return nil
}
