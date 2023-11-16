package gamepath

import (
	"runtime"
	"os"

	"egreg10us/faultylauncher/util/logutil"
)

const UserOS string = runtime.GOOS
const UserArch string = runtime.GOARCH
const P string = string(os.PathSeparator)

var (
	Gamedir string
	Versionsdir string
	Runtimesdir string
	Librariesdir string
	Assetsdir string
)

func init() {
	Minecraft()
	Versions()
	Libraries()
	Assets()
	Runtimes()
}

func Makedir(path string) error {
	err := os.MkdirAll(path,os.ModePerm)
	return err
}

func Minecraft() string {
	switch UserOS {
	case "windows":
		Gamedir = os.Getenv("APPDATA")+P+".minecraft" 
		Makedir(Gamedir)
		return Gamedir
	case "darwin":
		Gamedir = os.Getenv("HOME")+P+"Library"+P+"Application Support"+P+"minecraft"
		Makedir(Gamedir)
		return Gamedir
	case "linux":
		Gamedir = os.Getenv("HOME")+P+".minecraft"
		Makedir(Gamedir)
		return Gamedir
	}
	logutil.Critical("OS not supported.")
	return ""
}

func Assets() string {
	Assetsdir = Gamedir+P+"assets"
	Makedir(Assetsdir)
	return Assetsdir 
}

func Libraries() string {
	Librariesdir = Gamedir+P+"libraries"
	Makedir(Librariesdir)
	return Librariesdir
}

func Runtimes() string {
	Runtimesdir = Gamedir+P+"runtime"
	Makedir(Runtimesdir)
	return Runtimesdir
}

func Versions() string {
	Versionsdir = Gamedir+P+"versions"
	Makedir(Versionsdir)
	return Versionsdir
}
