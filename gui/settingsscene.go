package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/container"

	"egreg10us/faultylauncher/gui/resources"
)

func settingsScene(currentCanvas fyne.Canvas) {
	heading := widget.NewRichTextFromMarkdown("# Launcher settings")
	appearanceHeading := widget.NewLabel("Appearance")
	appearanceHeading.TextStyle = fyne.TextStyle{Bold:true}
	appearanceThemeLabel := widget.NewLabel("Theme style")
	appearanceTheme := widget.NewRadioGroup([]string{"Dark","Light"},func(option string){
		switch option {
		case "Dark":
			MainApp.Settings().SetTheme(&resources.DefaultDarkTheme{})
		case "Light":
			MainApp.Settings().SetTheme(&resources.DefaultLightTheme{})
		}
	})
	appearanceTheme.Horizontal = true
	appearanceTheme.Selected = lSettings.LauncherTheme
	launcherHeading := widget.NewLabel("Launcher")
	launcherHeading.TextStyle = fyne.TextStyle{Bold:true}
	launchruleLabel := widget.NewLabel("When game launches")
	ruleDescriptors := []string{"Hide the launcher","Exit the launcher","Do nothing"}
	launchruleSelection := widget.NewSelect(ruleDescriptors,func(string){})
	fileValidationCheck := widget.NewCheck("Disable file validation(not recommended)",func(bool){})
	fileValidationCheck.Checked = lSettings.DisableValidation
	switch lSettings.LaunchRule[0] {
	case 'H':
		launchruleSelection.Selected = ruleDescriptors[0]
	case 'E':
		launchruleSelection.Selected = ruleDescriptors[1]
	case 'D':
		launchruleSelection.Selected = ruleDescriptors[2]
	}
	saveButton := widget.NewButton("Save",func(){
		lSettings.LauncherTheme = appearanceTheme.Selected
		switch launchruleSelection.Selected[0] {
		case 'H':
			lSettings.LaunchRule = "Hide"
		case 'E':
			lSettings.LaunchRule = "Exit"
		case 'D':
			lSettings.LaunchRule = "DoNothing"
		}
		lSettings.DisableValidation = fileValidationCheck.Checked
		lSettings.SaveToFile()
		ReloadSettings()
		mainScene(currentCanvas)
	})
	closeButton := widget.NewButton("Cancel",func(){
		ReloadSettings()
		mainScene(currentCanvas)
	})
	currentCanvas.SetContent(
		container.NewVBox(
			heading,
			appearanceHeading,
			appearanceThemeLabel,
			container.NewHBox(appearanceTheme),
			launcherHeading,
			container.New(
				layout.NewFormLayout(),
				launchruleLabel,
				container.New(&MLayout{},launchruleSelection),
			),
			fileValidationCheck,
			layout.NewSpacer(),
			container.NewPadded(container.NewHBox(layout.NewSpacer(), closeButton, saveButton)),
		),
	)
}
