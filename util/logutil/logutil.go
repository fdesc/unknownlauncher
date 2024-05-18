package logutil

import (
   "path/filepath"
   "bufio"
   "time"
   "os"
)

var CurrentLogDate = time.Now().Format("2006-01-02")
var CurrentLogTime = time.Now().Format("15.04.05")
var CurrentLogPath string

type Loglevel struct {
   Header string
   IsError bool
}

func stdoutPrinter(msg string,target *os.File) {
   Save(msg)
   writer := bufio.NewWriter(target)
   defer writer.Flush()
   writer.WriteString(msg+"\n")
}

func Save(msg string) string {
   var file *os.File
   var err error
   fileName := "launcher_"+CurrentLogDate+".log"
   _,patherr := os.Stat(filepath.Join(CurrentLogPath,fileName))
   if patherr == nil {
      file,err = os.OpenFile(filepath.Join(CurrentLogPath,fileName),os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
      if err != nil { Error("Failed to open file",err); return "" }
      defer file.Close()
      _,err = file.WriteString(msg+"\n")
      if err != nil { Error("Failed to write log",err); return "" }
      return filepath.Join(CurrentLogPath,fileName)
   } else {
      file,err = os.Create(filepath.Join(CurrentLogPath,fileName))
      if err != nil {
         err := os.MkdirAll(CurrentLogPath,os.ModePerm)
         if err != nil {
            Error("Failed to create directories for log path",err)
            return ""
         }
         file,_ = os.Create(filepath.Join(CurrentLogPath,fileName))
      }
      defer file.Close()
      _,err = file.WriteString(msg+"\n")
      if err != nil { Error("Failed to write log",err); return "" }
      return filepath.Join(CurrentLogPath,fileName)
   }
}

func Info(msg string) {
   Custom(&Loglevel{Header:"INFO"},msg)(msg)
}

func Warn(msg string) {
   Custom(&Loglevel{Header:"WARN"},msg)(msg)
}

func Error(msg string,err error) {
   Custom(&Loglevel{Header:"ERROR"},msg+" "+err.Error())(msg+" "+err.Error())
}

func Custom(level *Loglevel,msg string) func(msg string) {
   if level.IsError {
      return func(string) {
         stdoutPrinter(level.Header+": "+msg,os.Stderr)
      }
   } else {
      return func(string) {
         stdoutPrinter(level.Header+": "+msg,os.Stdout)
      }
   }
}
