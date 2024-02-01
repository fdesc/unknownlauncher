package elements

import (
	"errors"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func NewProfileNameElem(pName string) (*widget.Label,*widget.Entry) {
	text := widget.NewLabel("Name")
	text.TextStyle = fyne.TextStyle{Bold:true}
	entry := widget.NewEntry()
	if pName != "" {
		entry.Text = pName
	}
	return text,entry
}

func NewVersionElem(typeList []string,versionList map[string][]string,lastVersion string,lastType string) (*widget.Select,*widget.Select,*widget.Label) {
	versionText := widget.NewLabel("Version")
	versionText.TextStyle = fyne.TextStyle{Bold:true}
	vType := widget.NewSelect(
		typeList,
		func(selection string){},
	)
	if lastType != "" {
		vType.Selected = lastType
		vType.Refresh()
	} else {
		vType.Selected = typeList[0]
	}
	vList := widget.NewSelect(
		versionList[lastType],
		func(selection string){},
	)
	if lastVersion != "" {
		vList.Selected = lastVersion
		vList.Refresh()
	} else {
		vList.Selected = versionList[vType.Selected][0]
	}
	vType.OnChanged = func(selection string) {
		vList.Options = versionList[selection]
		vList.Selected = versionList[vType.Selected][0]
		vList.Refresh()
	}
	return vType,vList,versionText
}

func NewGameDirElem(pGamedir string,pSeparate bool) (*widget.Label,*widget.Entry,*widget.Button,*widget.Check) {
	button := widget.NewButton("Browse",func(){})
	text := widget.NewLabel("Game directory")
	entry := widget.NewEntry()
	entry.Validator = fyne.StringValidator(ValidateGameDir())
	separateInstallation := widget.NewCheck("Separate installation",func(value bool){})
	separateInstallation.Disable()
	entry.SetPlaceHolder("Default")
	entry.OnChanged = func(text string){
		if len(text) > 0 {
			separateInstallation.Enable()
		} else {
			separateInstallation.Disable()
		}
	}
	if pGamedir != "" {
		entry.Text = pGamedir
		if pSeparate {
			separateInstallation.Checked = true
		}
		separateInstallation.Enable()
	}
	return text,entry,button,separateInstallation
}

func NewJavaDirElem(pJavaDir string) (*widget.Label,*widget.Entry,*widget.Button) {
	text := widget.NewLabel("Java executable ")
	entry := widget.NewEntry()
	entry.Validator = fyne.StringValidator(ValidateJavaExec())
	entry.SetPlaceHolder("Version default")
	button := widget.NewButton("Browse",func(){})
	if pJavaDir != "" {
		entry.Text = pJavaDir
	}
	return text,entry,button
}

func NewJVMArgsElem(pJVMArgs string) (*widget.Label,*widget.Entry) {
	text := widget.NewLabel("JVM arguments ")
	entry := widget.NewEntry()
	entry.Text = pJVMArgs
	return text,entry
}

func NewResolutionElem(pWidth,pHeight int,pFullscreen bool) (*widget.Label,*widget.Entry,*widget.Entry,*widget.Check) {
	text := widget.NewLabel("Game resolution")
	widthEntry := widget.NewEntry()
	heightEntry := widget.NewEntry()
	widthEntry.SetPlaceHolder("Width")
	heightEntry.SetPlaceHolder("Height")
	fullscreenCheck := widget.NewCheck("Fullscreen",func(value bool){
		if value {
			widthEntry.SetPlaceHolder("Width")
			heightEntry.SetPlaceHolder("Height")
			widthEntry.Text = ""
			heightEntry.Text = ""
			widthEntry.Disable()
			heightEntry.Disable()
		} else {
			widthEntry.Text = "854"
			heightEntry.Text = "480"
			widthEntry.Enable()
			heightEntry.Enable()
		}
	})
	return text,widthEntry,heightEntry,fullscreenCheck
}

func ValidateJavaExec() func(string) error {
	return func(path string) error {
		if path == "" {
			return nil
		}
		f,err := os.Open(path)
		if err != nil {
			return err
		}
		fStat,err := f.Stat()
		if err != nil {
			return err
		}
		if fStat.IsDir() {
			return errors.New("Not a file")
		} else {
			return nil
		}
	}
}

func ValidateGameDir() func(string) error {
	return func(path string) error {
		if path == "" {
			return nil
		}
		f,err := os.Open(path)
		if err != nil {
			return err
		}
		fStat,err := f.Stat()
		if err != nil {
			return err
		}
		if fStat.IsDir() {
			return nil
		} else {
			return errors.New("Not a directory")
		}
	}
}
