package main

import (
	"path/filepath"

	"fdesc/unknownlauncher/gui"
	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/launcher"
	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/launcher/versionmanager"
	"fdesc/unknownlauncher/util/gamepath"
	"fdesc/unknownlauncher/util/logutil"
)

const appName    = "unknownLauncher"
const appVersion = "Alpha 0.1"

func main() {
	gamepath.Reload()
	logutil.CurrentLogPath = filepath.Join(gamepath.Gamedir,"logs","launcher")
	logutil.Info("Starting application... the time is "+logutil.CurrentLogDate+" | "+logutil.CurrentLogTime)
	versionmanager.GetVersionList()
	launcher.GetLauncherContent()
	settings,err := launcher.ReadLauncherSettings()
	if err != nil {
		logutil.Error("Failed to read settings(launcher_settings)",err)
      gui.InitialError("Failed to read settings(launcher_settings)",err)
      return
	}
	profilesData,err := profilemanager.ReadProfilesRoot()
	if err != nil {
		logutil.Error("Failed to read profiles root(launcher_profiles.json)",err)
      gui.InitialError("Failed to read profiles root(launcher_profiles.json)",err)
      return
	}
	authData,err := auth.ReadAccountsRoot()
	if err != nil {
		logutil.Error("Failed to read accounts root(accounts.json)",err)
      gui.InitialError("Failed to read accounts root(accounts.json)",err)
      return
	}
   gui.SetSettings(settings)
   gui.SetProfilesRoot(&profilesData)
   gui.SetAccountsRoot(&authData)
   g := gui.NewGui()
   g.SetProperties()
   g.Start(appName+":"+appVersion)
}
