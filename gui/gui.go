package gui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"egreg10us/faultylauncher/auth"
	"egreg10us/faultylauncher/launcher"
	"egreg10us/faultylauncher/gui/resources"
	"egreg10us/faultylauncher/launcher/profilemanager"
)

var previewAccountSkin 		image.Image 
var previewAccountName 		string
var previewAccountUUID 		string
var previewProfileName 		string
var previewProfileVersion	string
var previewProfileVersionType	string
var lProfiles  		    = 	&profilemanager.ProfilesRoot{}
var lAccounts  		    = 	&auth.AccountsRoot{}
var lSettings  		    =	&launcher.LauncherSettings{}
var MainApp    		    = 	app.New()
var MainWindow 		    = 	MainApp.NewWindow("")

func init() {
	MainWindow.Resize(fyne.Size{Height: 480, Width: 680})
	MainWindow.SetFixedSize(true)
	ReloadSettings()
}

func ReloadSettings() {
	var err error
	lSettings,err = launcher.ReadLauncherSettings()
	if err != nil {
		return
		// exit gui and show a crash window
	}
	if lSettings.LauncherTheme == "Light" {
		MainApp.Settings().SetTheme(&resources.DefaultLightTheme{})
	} else {
		MainApp.Settings().SetTheme(&resources.DefaultDarkTheme{})
	}
}

func SetProfilesRoot(pRoot *profilemanager.ProfilesRoot) {
	lProfiles = pRoot
}

func SetAccountsRoot(aRoot *auth.AccountsRoot) {
	lAccounts = aRoot
}

func setCurrentAccountProperties(skinUrl string) {
	previewAccountName = lAccounts.Accounts[lAccounts.LastUsed].Name
	previewAccountSkin = auth.CropSkinImage(skinUrl)
	previewAccountUUID = lAccounts.Accounts[lAccounts.LastUsed].AccountUUID
}

func setCurrentProfileProperties() {
	if lProfiles.Profiles[lProfiles.LastUsedProfile].Name != "" {
		previewProfileName = lProfiles.Profiles[lProfiles.LastUsedProfile].Name
		previewProfileVersion = lProfiles.Profiles[lProfiles.LastUsedProfile].LastGameVersion
		previewProfileVersionType = lProfiles.Profiles[lProfiles.LastUsedProfile].LastGameType 
	} else {
		previewProfileName = lProfiles.Profiles[lProfiles.LastUsedProfile].Type
		previewProfileVersion = lProfiles.Profiles[lProfiles.LastUsedProfile].LastGameVersion
		previewProfileVersionType = lProfiles.Profiles[lProfiles.LastUsedProfile].LastGameType 
	}
}
