package gui

import (
	"path/filepath"
	"image/color"
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"

	"egreg10us/faultylauncher/auth"
	"egreg10us/faultylauncher/launcher"
	"egreg10us/faultylauncher/gui/resources"
	"egreg10us/faultylauncher/launcher/profilemanager"
	"egreg10us/faultylauncher/util/logutil"
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
	MainWindow.SetMaster()
	MainWindow.SetFixedSize(true)
}

func ReloadSettings() {
	var err error
	lSettings,err = launcher.ReadLauncherSettings()
	if err != nil {
		logutil.Error("Failed to read launcher settings",err)
		ErrorScene(err)
		return
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

func ErrorScene(err error) {
	MainWindow.Hide()
	errorWindow := MainApp.NewWindow("Fatal error")
	errorWindow.SetOnClosed(func(){
		MainApp.Quit()
		os.Exit(1)
	})
	errorWindow.Resize(fyne.Size{Height: 300, Width: 600})
	icon := canvas.NewImageFromResource(theme.BrokenImageIcon())
	icon.FillMode = canvas.ImageFillOriginal
	rect := canvas.NewRectangle(color.RGBA{R: 25,G:25, B:25,A: 200})
	headingLabel := widget.NewLabel("The application locked itself due to a critical error")
	headingLabel.TextStyle = fyne.TextStyle{Bold:true}
	errContent := widget.NewTextGridFromString("Log output: "+<-logutil.LogChannel+"Complete error: "+err.Error())
	closeButton := widget.NewButton("Exit the application",func(){errorWindow.Close()})
	logButton := widget.NewButton("Open launcher log file",func(){launcher.InvokeDefault(filepath.Join(logutil.CurrentLogPath,"launcher_"+logutil.CurrentLogTime+".log"))})
	errorWindow.SetContent(
		container.NewBorder(
			container.NewVBox(
				container.NewCenter(icon),
				container.NewCenter(headingLabel),
			),
			container.NewHBox(layout.NewSpacer(), closeButton,logButton, layout.NewSpacer()),
			nil,
			nil,
			container.NewStack(rect,container.NewScroll(errContent)),
		),
	)
	errorWindow.ShowAndRun()
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
