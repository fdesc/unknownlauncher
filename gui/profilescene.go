package gui

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"

	"egreg10us/faultylauncher/launcher/versionmanager"
	"egreg10us/faultylauncher/launcher/profilemanager"
	"egreg10us/faultylauncher/gui/resources"
	"egreg10us/faultylauncher/gui/elements"
	"egreg10us/faultylauncher/util/logutil"
)

func listProfiles(currentCanvas fyne.Canvas) {
	var nameUUIDPairs = make(map[string]string)
	var profileDisplayNames []string
	for k,v := range lProfiles.Profiles {
		if v.Name != "" {
			profileDisplayNames = append(profileDisplayNames, v.Name)
			nameUUIDPairs[v.Name] = k
		} else {
			profileDisplayNames = append(profileDisplayNames, v.Type)
			nameUUIDPairs[v.Type] = k
		}
	}
	profileBind := binding.BindStringList(&profileDisplayNames)

	profileImage := canvas.NewImageFromResource(resources.ProfileIcon)
	profileImage.FillMode = canvas.ImageFillOriginal
	var profileList *widget.List
	profileList = widget.NewListWithData(profileBind,
	func() fyne.CanvasObject {
		label := widget.NewLabel("")
		toolbar := widget.NewToolbar(
				widget.NewToolbarAction(theme.DocumentCreateIcon(), func(){
					profile := lProfiles.Profiles[nameUUIDPairs[label.Text]]
					editProfile(&profile,nameUUIDPairs[label.Text],profileImage,"Edit Profile",currentCanvas)
				}),
				widget.NewToolbarAction(theme.ContentCopyIcon(), func(){
					profile := lProfiles.Profiles[nameUUIDPairs[label.Text]]
					profile.LastUsed = time.Now().Format(time.RFC3339)
					profile.Created = time.Now().Format(time.RFC3339)
					uuid,err := profilemanager.GenerateProfileUUID()
					if err != nil { logutil.Error("",err) }
					lProfiles.Profiles[uuid] = profile
					lProfiles.SaveToFile()
					listProfiles(currentCanvas)
				}),
				widget.NewToolbarAction(theme.DeleteIcon(), func(){
					headingLabel := widget.NewLabel("Remove profile?")
					headingLabel.TextStyle = fyne.TextStyle{Bold:true}
					var modal *widget.PopUp
					modal = widget.NewModalPopUp(
						container.New(
							layout.NewVBoxLayout(),
							headingLabel,
							layout.NewSpacer(),
							widget.NewLabel("This action cannot be undone."),
							layout.NewSpacer(),
							container.New(
								layout.NewGridLayoutWithRows(1),
								widget.NewButton("Cancel",func(){modal.Hide()}),
								widget.NewButton("Confirm",func(){
									delete(lProfiles.Profiles,nameUUIDPairs[label.Text])
									lProfiles.SaveToFile()
									modal.Hide()
									listProfiles(currentCanvas)
								}),
							),
						),
						currentCanvas,
					)
					modal.Show()
				}),
			)
		box := container.New(
			layout.NewPaddedLayout(),
			container.NewBorder(nil,nil,profileImage,toolbar,label),
		)
		profileList.OnSelected = func(id widget.ListItemID) {
			pName,_ := profileBind.GetValue(id)
			lProfiles.LastUsedProfile = nameUUIDPairs[pName]
			lProfiles.SaveToFile()
			setCurrentProfileProperties()
			mainScene(currentCanvas)
		}
		return box
	},
	func(i binding.DataItem, o fyne.CanvasObject) {
		o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).Bind(i.(binding.String))
	})

	newProfileButton := widget.NewButtonWithIcon("New Profile",theme.ContentAddIcon(),func(){
		var newProfile = &profilemanager.ProfileProperties{}
		newProfile.Type = "Profile"
		newProfile.Created = time.Now().Format(time.RFC3339)
		newProfile.LastUsed = time.Now().Format(time.RFC3339)
		newUUID,err := profilemanager.GenerateProfileUUID()
		if err != nil { logutil.Error("",err) }
		editProfile(newProfile,newUUID,profileImage,"New Profile",currentCanvas)
	})

	currentCanvas.SetContent(
		container.NewBorder(nil,container.NewCenter(container.NewHBox(newProfileButton)),nil,nil,profileList),
	)
}

func editProfile (profileData *profilemanager.ProfileProperties,pUUID string,pImage *canvas.Image,headingString string,currentCanvas fyne.Canvas) {
	// Misc elements(decoration)
	separatorLine := canvas.NewLine(theme.OverlayBackgroundColor())
	separatorLine.StrokeWidth = 1
	headingText := widget.NewRichTextFromMarkdown("## "+headingString)
	// Profile elements
	nLabel,nEntry := elements.NewProfileNameElem(profileData.Name)
	var vTypeDisplayName []string
	for k := range versionmanager.VersionList {
		vTypeDisplayName = append(vTypeDisplayName,k)
	}
	vType,vList,vLabel := elements.NewVersionElem(vTypeDisplayName,versionmanager.VersionList,profileData.LastGameVersion,profileData.LastGameType)
	gameDirLabel,gameDirEntry,gameDirButton,gameDirCheck := elements.NewGameDirElem(profileData.GameDirectory,profileData.SeparateInstallation)
	javaDirLabel,javaDirEntry,javaDirButton := elements.NewJavaDirElem(profileData.JavaDirectory)
	jvmArgsLabel,jvmArgsEntry := elements.NewJVMArgsElem(profileData.JVMArgs)
	rLabel,wEntry,hEntry,fullscreenCheck := elements.NewResolutionElem(0,0,false)
	// Additional settings for resolution element
	if profileData.Resolution != nil {
		if (profileData.Resolution.Width != 0 && profileData.Resolution.Height != 0) {
			wEntry.Text = strconv.Itoa(profileData.Resolution.Width)
			hEntry.Text = strconv.Itoa(profileData.Resolution.Height)
		}
		if profileData.Resolution.Fullscreen {
			wEntry.SetPlaceHolder("Width")
			hEntry.SetPlaceHolder("Height")
			wEntry.Text = ""
			hEntry.Text = ""
			wEntry.Disable()
			hEntry.Disable()
		}
		fullscreenCheck.Checked = profileData.Resolution.Fullscreen
	}
	//
	saveButton := widget.NewButtonWithIcon("Save",theme.ConfirmIcon(),func(){
		profileData.Name = nEntry.Text
		profileData.LastGameVersion = vList.Selected
		profileData.LastGameType = vType.Selected
		if (len(jvmArgsEntry.Text) >= 5) {
			profileData.JVMArgs = jvmArgsEntry.Text
		} else if jvmArgsEntry.Text == "" {
			profileData.JVMArgs = jvmArgsEntry.Text
		}
		if fullscreenCheck.Checked {
			profileData.Resolution = &profilemanager.ProfileResolution{Fullscreen:fullscreenCheck.Checked}
		} else if (wEntry.Text == "" || hEntry.Text == "" && !fullscreenCheck.Checked) {
			profileData.Resolution = nil
		} else {
			if (wEntry.Text != "" && hEntry.Text != "") {
				w,err := strconv.Atoi(wEntry.Text)
				if err != nil { logutil.Error("",err) }
				h,err := strconv.Atoi(hEntry.Text)
				if err != nil {	logutil.Error("",err) }
				profileData.Resolution = &profilemanager.ProfileResolution{Width:w,Height:h,Fullscreen:false}
			}
		}
		if elements.ValidateGameDir()(gameDirEntry.Text) != nil {
			return
		}
		profileData.GameDirectory = gameDirEntry.Text
		profileData.SeparateInstallation = gameDirCheck.Checked
		if elements.ValidateJavaExec()(javaDirEntry.Text) != nil {
			return
		}
		profileData.JavaDirectory = javaDirEntry.Text
		lProfiles.Profiles[pUUID] = *profileData
		lProfiles.SaveToFile()
		listProfiles(currentCanvas)
	})
	cancelButton := widget.NewButtonWithIcon("Cancel",theme.ContentClearIcon(),func(){listProfiles(currentCanvas)})

	currentCanvas.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			container.NewPadded(container.NewBorder(nil,nil,pImage,nil,headingText)),
			nLabel,
			container.New(
				layout.NewPaddedLayout(),
				container.New(&MLayout{}, nEntry),
			),
			vLabel,
			container.New(
				layout.NewPaddedLayout(),
				container.New(layout.NewGridLayoutWithRows(1), container.New(layout.NewHBoxLayout(), vType,vList)),
			),
			separatorLine,
			layout.NewSpacer(),
			container.NewBorder(nil,nil,gameDirLabel,container.New(layout.NewHBoxLayout(), gameDirButton,gameDirCheck),gameDirEntry),
			layout.NewSpacer(),
			container.NewBorder(nil,nil,rLabel,fullscreenCheck,container.New(layout.NewGridLayoutWithRows(1),wEntry,hEntry)),
			layout.NewSpacer(),
			container.NewBorder(nil,nil,javaDirLabel,javaDirButton,javaDirEntry),
			layout.NewSpacer(),
			container.NewBorder(nil,nil,jvmArgsLabel,nil,jvmArgsEntry),
			layout.NewSpacer(),
			container.New(layout.NewCenterLayout(), container.New(layout.NewHBoxLayout(), cancelButton,saveButton)),
		),
	)
}
