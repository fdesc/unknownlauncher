package elements

import (
	"image"

	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/launcher"
	"fdesc/unknownlauncher/gui/resources"
	"fdesc/unknownlauncher/util/downloadutil"
	"fdesc/unknownlauncher/launcher/profilemanager"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Home struct {
	Account          auth.AccountProperties
	Profile          profilemanager.ProfileProperties
	BtnSettings      *widget.Button
	ListAccountsFunc func()
	ListProfilesFunc func()
	PopUpCanvas      fyne.Canvas
	AccountCnt       *fyne.Container
	ProfileCnt       *fyne.Container
	BtnProfile       *widget.Button
	BtnPlay          *widget.Button
	PlayCnt          *fyne.Container
	SettingsCnt      *fyne.Container
	BtnReadMore      *widget.Button
	BaseCnt          *fyne.Container
}

func NewHome() *Home {
	hs := &Home{}
	settingsLabel := widget.NewLabel("Settings")
	settingsImg := canvas.NewImageFromResource(theme.SettingsIcon())
	settingsImg.FillMode = canvas.ImageFillOriginal
	accountLabel := widget.NewLabel("")
	newsLabel := canvas.NewText(launcher.ContentMessage,theme.ForegroundColor())
	newsLabel.Alignment = fyne.TextAlignCenter
	newsLabel.TextStyle = fyne.TextStyle{Bold: true}
	newsImage := canvas.NewImageFromImage(launcher.ContentImage)
	newsImage.FillMode = canvas.ImageFillOriginal
	playHeading := widget.NewLabel("Play")
	playHeading.TextStyle = fyne.TextStyle{Bold: true}
	playIcon := canvas.NewImageFromResource(theme.MediaPlayIcon())
	profileIcon := canvas.NewImageFromResource(resources.ProfileIcon)
	accountInfoBtn := widget.NewButton("",func(){
		var modal *widget.PopUp
		skinImg := canvas.NewImageFromImage(
			hs.AccountCnt.Objects[0].(*fyne.Container).
			             Objects[1].(*fyne.Container).
		                     Objects[0].(*fyne.Container).
		                     Objects[0].(*fyne.Container).
			             Objects[1].(*fyne.Container).
	                             Objects[0].(*fyne.Container).
		                     Objects[0].(*canvas.Image).Image,
		)
		skinImg.FillMode = canvas.ImageFillOriginal
		nameLabel := widget.NewLabel(hs.Account.Name)
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}
		modal = widget.NewModalPopUp(
			container.NewVBox(
				container.NewCenter(widget.NewLabel("Logged in as")),
				skinImg,
				container.NewCenter(nameLabel),
				layout.NewSpacer(),
				container.NewGridWithRows(
					0,
					widget.NewButton("Ok",func(){ modal.Hide() }),
					widget.NewButton("Change account",func(){
						hs.ListAccountsFunc()
						modal.Hide()
					}),
				),
			),
			hs.PopUpCanvas,
		)
		modal.Show()
	})
	profileHeading := widget.NewLabel("")
	profileHeading.TextStyle = fyne.TextStyle{Bold: true}
	profileLabel := widget.NewLabel(hs.Profile.LastGameType+" "+hs.Profile.LastGameVersion)
	playLabel := widget.NewLabel("")
	progress := widget.NewProgressBarInfinite()
	progress.Hide()
	progressText := canvas.NewText(downloadutil.CurrentFile,theme.ForegroundColor())
	progressText.Hide()
	hs.BtnSettings = widget.NewButton("",func(){})
	hs.BtnReadMore = widget.NewButton("Read more",func(){ launcher.InvokeDefault(launcher.ContentReadMoreLink) })
	hs.BtnProfile = widget.NewButton("",func(){ hs.ListProfilesFunc() })
	hs.BtnPlay = widget.NewButton("",func(){})
	hs.SettingsCnt = NewSquareButtonWithIcon(settingsLabel,settingsImg,hs.BtnSettings,36)
	hs.PlayCnt = NewRectangleButtonWithIcon(playHeading,playLabel,playIcon,hs.BtnPlay,127)
	hs.AccountCnt = NewSquareButtonWithIcon(accountLabel,canvas.NewImageFromResource(nil),accountInfoBtn,36)
	hs.ProfileCnt = NewRectangleButtonWithIcon(profileHeading,profileLabel,profileIcon,hs.BtnProfile,420)
	hs.BaseCnt = container.NewVBox(
		layout.NewSpacer(),
		container.NewBorder(
			nil,
			newsLabel,
			nil,
			nil,
			newsImage,
		),
		container.NewHBox(
			layout.NewSpacer(),
			hs.BtnReadMore,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
		container.NewVBox(
			container.NewPadded(progressText),
			container.NewGridWrap(fyne.NewSize(675,1),progress),
		),
		container.NewHBox(
			hs.AccountCnt,
			hs.SettingsCnt,
			hs.ProfileCnt,
			hs.PlayCnt,
		),
	)
	return hs
}

func (hs *Home) Update(account auth.AccountProperties, profile profilemanager.ProfileProperties) {
	hs.Profile = profile
	hs.Account = account
	lastVersion := hs.Profile.LastGameVersion
	accountName := hs.Account.Name
	if len(accountName) > 6 {
		accountName = accountName[:5]+"..."
	}
	if len(lastVersion) > 6 {
		lastVersion = lastVersion[:6]+"..."
	}
	if hs.Profile.Name != "" {
		hs.ProfileCnt.Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*widget.Label).SetText(hs.Profile.Name)
	} else {
		hs.ProfileCnt.Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*widget.Label).SetText(hs.Profile.Type)
	}

	hs.ProfileCnt.Objects[0].(*fyne.Container).
		      Objects[1].(*fyne.Container).
	              Objects[0].(*fyne.Container).
	              Objects[0].(*fyne.Container).
	              Objects[0].(*fyne.Container).
	              Objects[1].(*widget.Label).SetText(hs.Profile.LastGameType+" "+hs.Profile.LastGameVersion)

	hs.PlayCnt.Objects[0].(*fyne.Container).
		   Objects[1].(*fyne.Container).
	           Objects[0].(*fyne.Container).
	           Objects[0].(*fyne.Container).
	           Objects[0].(*fyne.Container).
	           Objects[1].(*widget.Label).SetText(lastVersion)

	hs.AccountCnt.Objects[0].(*fyne.Container).
	              Objects[1].(*fyne.Container).
	              Objects[0].(*fyne.Container).
	              Objects[0].(*fyne.Container).
	              Objects[0].(*fyne.Container).
		      Objects[0].(*widget.Label).SetText(accountName)
}

func (hs *Home) SetSkinIcon(icon image.Image) {
	if !auth.DefaultSkinIcon {
		hs.AccountCnt.Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0] = canvas.NewImageFromImage(icon)
	} else {
		hs.AccountCnt.Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*canvas.Image).Resource = resources.UnknownSkinIcon

		hs.AccountCnt.Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*canvas.Image).Refresh()
	}
}
