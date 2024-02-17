package gamepath

import (
	"runtime"
	"os"

	"fdesc/unknownlauncher/util/logutil"
)

var SeparateInstallation = false
const UserOS 		 = runtime.GOOS
const UserArch string 	 = runtime.GOARCH
const P string 		 = string(os.PathSeparator)

var (
	Gamedir string
	Versionsdir string
	Runtimesdir string
	Librariesdir string
	Assetsdir string
)

func init() {
	Reload()
}

func Reload() {
	if !SeparateInstallation {
		Mcdir()	
		Versions()
		Libraries()
		Assets()
		Runtimes()
		return
	}
	Versions()
	Libraries()
	Assets()
	Runtimes()
}

func Mcdir() string {
	switch UserOS {
	case "windows":
		Gamedir = os.Getenv("APPDATA")+P+".minecraft" 
		err := os.MkdirAll(Gamedir,os.ModePerm)
		if err != nil { logutil.Error("Failed to create game directories",err); os.Exit(1) }
		return Gamedir
	case "darwin":
		Gamedir = os.Getenv("HOME")+P+"Library"+P+"Application Support"+P+"minecraft"
		err := os.MkdirAll(Gamedir,os.ModePerm)
		if err != nil { logutil.Error("Failed to create game directories",err); os.Exit(1) }
		return Gamedir
	case "linux":
		Gamedir = os.Getenv("HOME")+P+".minecraft"
		err := os.MkdirAll(Gamedir,os.ModePerm)
		if err != nil { logutil.Error("Failed to create game directories",err); os.Exit(1) }
		return Gamedir
	}
	logutil.Warn("OS not supported.")
	os.Exit(1)
	return ""
}

func Assets() string {
	Assetsdir = Gamedir+P+"assets"
	err := os.MkdirAll(Assetsdir,os.ModePerm)
	if err != nil { logutil.Error("Failed to create game directories",err); os.Exit(1) }
	return Assetsdir 
}

func Libraries() string {
	Librariesdir = Gamedir+P+"libraries"
	err := os.MkdirAll(Librariesdir,os.ModePerm)
	if err != nil { logutil.Error("Failed to create game directories",err); os.Exit(1) }
	return Librariesdir
}

func Runtimes() string {
	Runtimesdir = Gamedir+P+"runtime"
	err := os.MkdirAll(Runtimesdir,os.ModePerm)
	if err != nil { logutil.Error("Failed to create game directories",err); os.Exit(1) }
	return Runtimesdir
}

func Versions() string {
	Versionsdir = Gamedir+P+"versions"
	err := os.MkdirAll(Versionsdir,os.ModePerm)
	if err != nil { logutil.Error("Failed to create game directories",err); os.Exit(1) }
	return Versionsdir
}
