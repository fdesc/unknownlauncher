package main

import (
	"time"

	"egreg10us/faultylauncher/gui"
	"egreg10us/faultylauncher/auth"
	"egreg10us/faultylauncher/launcher"
	"egreg10us/faultylauncher/launcher/profilemanager"
	"egreg10us/faultylauncher/launcher/versionmanager"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
)

const appName 	 = "unknownLauncher"
const appVersion = "DevBuild 0.1"

func main() {
	gamepath.Reload()
	currentTime := time.Now()
	launcher.GetLauncherContent()
	versionmanager.GetVersionList()
	profilesData,err := profilemanager.ReadProfilesRoot()
	if err != nil { logutil.Error("Failed to read profiles root",err) }
	authData,err := auth.ReadAccountsRoot()
	if err != nil { logutil.Error("Failed to read accounts root",err) }
	gui.ReloadSettings()
	gui.SetAccountsRoot(&authData)
	gui.SetProfilesRoot(&profilesData)
	mainCanvas := gui.MainWindow.Canvas()
	gui.NewAccountScene(mainCanvas)
	gui.MainWindow.SetTitle(appName+": "+appVersion)
	gui.MainWindow.ShowAndRun()
	logutil.Save(gamepath.Gamedir,currentTime)
}
