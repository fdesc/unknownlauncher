package main

import (
	"path/filepath"
	"time"

	"fdesc/unknownlauncher/gui"
	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/launcher"
	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/launcher/versionmanager"
	"fdesc/unknownlauncher/util/gamepath"
	"fdesc/unknownlauncher/util/logutil"
)

const appName 	 = "unknownLauncher"
const appVersion = "Alpha 0.1"

// TODO: use zstd compression for launcher logs
// TODO: use maps instead of slices in function generateArguments at launcher package

func main() {
	gamepath.Reload()
	logutil.CurrentLogPath = filepath.Join(gamepath.Gamedir,"logs","launcher")
	logutil.Info("Starting application... the time is "+logutil.CurrentLogTime+" | "+time.Now().Format("15.04.05"))
	versionmanager.GetVersionList()
	launcher.GetLauncherContent()
	profilesData,err := profilemanager.ReadProfilesRoot()
	if err != nil { 
		logutil.Error("Failed to read profiles root",err)
		gui.ErrorScene(err)
	}
	authData,err := auth.ReadAccountsRoot()
	if err != nil { 
		logutil.Error("Failed to read accounts root",err)
		gui.ErrorScene(err)
	}
	gui.ReloadSettings()
	gui.SetAccountsRoot(&authData)
	gui.SetProfilesRoot(&profilesData)
	mainCanvas := gui.MainWindow.Canvas()
	gui.NewAccountScene(mainCanvas)
	gui.MainWindow.SetTitle(appName+": "+appVersion)
	gui.MainWindow.ShowAndRun()
	close(logutil.LogChannel)
}
