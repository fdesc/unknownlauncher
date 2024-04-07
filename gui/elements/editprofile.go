package elements

import (
	"strconv"
	"errors"
	"os"

	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/launcher/versionmanager"
	"fdesc/unknownlauncher/gui/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ProfileEdit struct {
	ProfileUUID             string
	Profile                 *profilemanager.ProfileProperties
	NameEntry               *widget.Entry
	VersionTSelect          *widget.Select
	VersionSelect           *widget.Select
	GameDirEntry            *widget.Entry
	SeparateInstallation    *widget.Check
	ResolutionHEntry        *widget.Entry
	ResolutionWEntry        *widget.Entry
	FullscreenCheck         *widget.Check
	JavaExecEntry           *widget.Entry
	JavaArgsEntry           *widget.Entry
	BtnOk                   *widget.Button
	BtnCancel               *widget.Button
	SaveProfileFunc         func()
	BaseCnt                 *fyne.Container
}

func NewProfileEdit() *ProfileEdit {
	pe := &ProfileEdit{}
	pe.Profile = &profilemanager.ProfileProperties{}
	icon := canvas.NewImageFromResource(resources.ProfileIcon)
	icon.FillMode = canvas.ImageFillOriginal
	heading := widget.NewRichTextFromMarkdown("# Edit profile")
	nameLabel := widget.NewLabel("Name")
	nameLabel.TextStyle = fyne.TextStyle{Bold:true}
	versionLabel := widget.NewLabel("Version")
	separatorLine := canvas.NewLine(theme.OverlayBackgroundColor())
	separatorLine.StrokeWidth = 1
	gameDirLabel := widget.NewLabel("Game directory")
	gameDirButton := widget.NewButton("Browse",func(){})
	resolutionLabel := widget.NewLabel("Resolution")
	javaDirLabel := widget.NewLabel("Java executable ")
	javaDirButton := widget.NewButton("Browse",func(){})
	javaArgsLabel := widget.NewLabel("JVM arguments")
	pe.NameEntry = widget.NewEntry()
	var versionTypeSlice []string
	for k := range versionmanager.VersionList {
		versionTypeSlice = append(versionTypeSlice, k)
	}
	pe.VersionTSelect = widget.NewSelect(versionTypeSlice,func(s string){
		pe.VersionSelect.Options = versionmanager.VersionList[s]
		pe.VersionSelect.Selected = versionmanager.VersionList[s][0]
		pe.VersionSelect.Refresh()
	})
	pe.VersionTSelect.Selected = versionTypeSlice[0]
	pe.VersionSelect = widget.NewSelect(versionmanager.VersionList[pe.VersionTSelect.Selected],func(string){})
	if pe.Profile.LastGameType != "" {
		pe.VersionTSelect.Selected = pe.Profile.LastGameType
	}
	pe.GameDirEntry = widget.NewEntry()
	pe.GameDirEntry.Validator = fyne.StringValidator(func(path string) error {
		if path == "" { return nil }
		f,err := os.Open(path)
		if err != nil { return err }
		fStat,err := f.Stat()
		if err != nil { return err }
		if fStat.IsDir() {
			return nil
		} else {
			return errors.New("Not a directory")
		}
	})
	pe.SeparateInstallation = widget.NewCheck("Separate installation",func(bool){})
	pe.ResolutionHEntry = widget.NewEntry()
	pe.ResolutionHEntry.SetPlaceHolder("Height")
	pe.ResolutionWEntry = widget.NewEntry()
	pe.ResolutionWEntry.SetPlaceHolder("Width")
	pe.FullscreenCheck = widget.NewCheck("Fullscreen",func(checked bool){
		if checked {
			pe.ResolutionWEntry.Disable()
			pe.ResolutionHEntry.Disable()
		} else {
			pe.ResolutionWEntry.Text = "854"
			pe.ResolutionHEntry.Text = "480"
		}
	})
	pe.JavaExecEntry = widget.NewEntry()
	pe.JavaExecEntry.Validator = fyne.StringValidator(func(path string) error {
		if path == "" { return nil }
		f,err := os.Open(path)
		if err != nil { return err }
		fStat,err := f.Stat()
		if err != nil { return err }
		if fStat.IsDir() {
			return errors.New("Not a file")
		} else {
			return nil
		}
	})
	pe.JavaArgsEntry = widget.NewEntry()
	pe.BtnCancel = widget.NewButton("Cancel",func(){})
	pe.BtnOk = widget.NewButton("Save",func() {
		// log
		if pe.GameDirEntry.Validate() != nil {
			return
		}
		if pe.JavaExecEntry.Validate() != nil {
			return
		}
		pe.Profile.Name = pe.NameEntry.Text
		pe.Profile.LastGameVersion = pe.VersionSelect.Selected
		pe.Profile.LastGameType = pe.VersionTSelect.Selected
		if len(pe.JavaArgsEntry.Text) >= 5 || pe.JavaArgsEntry.Text == "" {
			pe.Profile.JVMArgs = pe.JavaArgsEntry.Text
		}
		if pe.FullscreenCheck.Checked {
			pe.Profile.Resolution = &profilemanager.ProfileResolution{Fullscreen:pe.FullscreenCheck.Checked}
		} else if pe.ResolutionHEntry.Text != "" && pe.ResolutionWEntry.Text != "" && !pe.FullscreenCheck.Checked {
			h,err := strconv.Atoi(pe.ResolutionHEntry.Text)
			if err != nil { /* log */ }
			w,err := strconv.Atoi(pe.ResolutionWEntry.Text)
			if err != nil { /* log */ }
			pe.Profile.Resolution = &profilemanager.ProfileResolution{Width:w,Height:h,Fullscreen:false}
		} else {
			pe.Profile.Resolution = nil
		}
		pe.GameDirEntry.Text = pe.Profile.GameDirectory
		pe.JavaExecEntry.Text = pe.Profile.JavaDirectory
		pe.Profile.SeparateInstallation = pe.SeparateInstallation.Checked
		pe.SaveProfileFunc()
	})
	pe.BaseCnt = container.NewVBox(
		container.NewPadded(container.NewBorder(nil,nil,icon,nil,heading)),
		nameLabel,
		container.New(&HalfLayout{}, pe.NameEntry),
		versionLabel,
		container.NewPadded(
			container.New(
				layout.NewGridLayoutWithRows(1),
				container.NewHBox(pe.VersionTSelect,pe.VersionSelect),
			),
		),
		separatorLine,
		layout.NewSpacer(),
		container.NewBorder(nil,nil,gameDirLabel,container.NewHBox(gameDirButton,pe.SeparateInstallation),pe.GameDirEntry),
		layout.NewSpacer(),
		container.NewBorder(
			nil,
			nil,
			resolutionLabel,
			pe.FullscreenCheck,
			container.New(
				layout.NewGridLayoutWithRows(1),
				pe.ResolutionWEntry,
				pe.ResolutionHEntry,
			),
		),
		layout.NewSpacer(),
		container.NewBorder(nil,nil,javaDirLabel,javaDirButton,pe.JavaExecEntry),
		layout.NewSpacer(),
		container.NewBorder(nil,nil,javaArgsLabel,nil,pe.JavaArgsEntry),
		layout.NewSpacer(),
		container.NewCenter(container.NewHBox(pe.BtnCancel,pe.BtnOk)),
	)
	return pe
}

func (pe *ProfileEdit) Update(profile *profilemanager.ProfileProperties,uuid string) {
	pe.Profile = profile
	pe.ProfileUUID = uuid
	pe.NameEntry.SetText("")
	if pe.Profile.Name != "" {
		pe.NameEntry.SetText(pe.Profile.Name)
	}
	if pe.Profile.LastGameVersion != "" {
		pe.VersionSelect.Selected = pe.Profile.LastGameVersion
		pe.VersionTSelect.Selected = pe.Profile.LastGameType
	} else {
		pe.VersionSelect.Selected = versionmanager.LatestRelease
		pe.VersionTSelect.Selected = "release"
	}
	pe.JavaArgsEntry.SetText(pe.Profile.JVMArgs)
	if pe.Profile.Resolution != nil {
		pe.FullscreenCheck.Checked = pe.Profile.Resolution.Fullscreen
		if pe.Profile.Resolution.Width != 0 && pe.Profile.Resolution.Height != 0 {
			pe.ResolutionHEntry.SetText(strconv.Itoa(pe.Profile.Resolution.Height))
			pe.ResolutionWEntry.SetText(strconv.Itoa(pe.Profile.Resolution.Width))
		}
	}
	pe.GameDirEntry.SetText(pe.Profile.GameDirectory)
	pe.JavaExecEntry.SetText(pe.Profile.JavaDirectory)
	pe.JavaArgsEntry.SetText(pe.Profile.JVMArgs)
	pe.SeparateInstallation.Checked = pe.Profile.SeparateInstallation
}
