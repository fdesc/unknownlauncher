package elements

import (
	"image"

	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/gui/resources"
	"fdesc/unknownlauncher/launcher"
	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/launcher/versionmanager"
	"fdesc/unknownlauncher/util/downloadutil"
	"fdesc/unknownlauncher/util/logutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Home struct {
	Account           auth.AccountProperties
	Profile           profilemanager.ProfileProperties
	BtnSettings       *widget.Button
	LaunchRuleFunc    func() string
	WindowHideFunc    func()
	WindowShowFunc    func()
	AppExitFunc       func()
	ListAccountsFunc  func()
	ListProfilesFunc  func()
   CrashInformerFunc func(err error,output string,logPath string)
	PopUpCanvas       fyne.Canvas
	AccountCnt        *fyne.Container
	ProfileCnt        *fyne.Container
	BtnProfile        *widget.Button
	BtnPlay           *widget.Button
	PlayCnt           *fyne.Container
	SettingsCnt       *fyne.Container
	BtnReadMore       *widget.Button
	BaseCnt           *fyne.Container
}

func NewHome() *Home {
	hs := &Home{}
	settingsLabel := widget.NewLabel("Settings")
	settingsImg := canvas.NewImageFromResource(theme.SettingsIcon())
	settingsImg.FillMode = canvas.ImageFillOriginal
	accountLabel := widget.NewLabel("")
	newsLabel := canvas.NewText(launcher.ContentMessage,theme.PlaceHolderColor())
	newsLabel.Alignment = fyne.TextAlignCenter
	newsLabel.TextStyle = fyne.TextStyle{Bold: true}
	var newsImage = new(canvas.Image)
	if launcher.ContentImage == nil {
		newsImage = canvas.NewImageFromResource(theme.BrokenImageIcon())
	} else {
		newsImage = canvas.NewImageFromImage(launcher.ContentImage)
	}
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
   var profileLabel *widget.Label
	profileHeading := widget.NewLabel("")
	profileHeading.TextStyle = fyne.TextStyle{Bold: true}
   if !launcher.OfflineMode {
      profileLabel = widget.NewLabel(versionmanager.GetVersionType(hs.Profile.LastVersion())+" "+hs.Profile.LastVersion())
   } else {
      profileLabel = widget.NewLabel("Local "+hs.Profile.LastVersion())
   }
	playLabel := widget.NewLabel("")
	progress := widget.NewProgressBarInfinite()
	progress.Hide()
	progressText := canvas.NewText(downloadutil.CurrentFile,theme.PlaceHolderColor())
	progressText.Hide()
	hs.BtnSettings = widget.NewButton("",func(){})
	hs.BtnReadMore = widget.NewButton("Read more",func(){ launcher.InvokeDefault(launcher.ContentReadMoreLink) })
   if launcher.OfflineMode {
      hs.BtnReadMore.Hide()
   }
	hs.BtnProfile = widget.NewButton("",func(){ hs.ListProfilesFunc() })
	hs.BtnPlay = widget.NewButton("",func() {
		var err error
		progress.Show()
		progressText.Show()
		playHeading.SetText("Loading")
		hs.BtnProfile.Disable()
		hs.BtnPlay.Disable()
		hs.BaseCnt.Refresh()
		first := downloadutil.CurrentFile
		go func() {
			task := &launcher.LaunchTask{Profile: &hs.Profile,Account: &hs.Account}
			err = task.Prepare()
			progress.Hide()
			progressText.Hide()
			playHeading.SetText("Play")
			hs.BtnProfile.Enable()
			hs.BtnPlay.Enable()
			if err != nil {
				logutil.Error("Failed to prepare the task",err)
				return
			} else {
				switch hs.LaunchRuleFunc() {
				case "Hide":
					hs.WindowHideFunc()
					command,logPath := task.CompleteArguments()
               output,err := command.CombinedOutput()
					if err != nil {
						hs.WindowShowFunc()
						logutil.Error("Failed to start the game",err)
						logutil.Info("Game output: \n"+string(output))
                  hs.CrashInformerFunc(err,string(output),logPath)
						return
					} else {
						hs.WindowShowFunc()
						return
					}
				case "Exit":
               command,_ := task.CompleteArguments()
               command.Start()
					hs.AppExitFunc()
					return
				case "DoNothing":
               command,logPath := task.CompleteArguments()
					output,err := command.CombinedOutput()
					if err != nil {
						logutil.Error("Failed to start the game",err)
						logutil.Info("Game output:"+string(output))
                  hs.CrashInformerFunc(err,string(output),logPath)
						return
					}
				}
			}
		}()
		go func() {
			for {
				if progressText.Hidden {
					return
				}
				if first != downloadutil.CurrentFile {
					progressText.Text = downloadutil.CurrentFile+" "+downloadutil.GetJobInfo()
					progressText.Refresh()
				}
			}
		}()
	})
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

func (hs *Home) Update(account auth.AccountProperties,profile profilemanager.ProfileProperties) {
   var lastType string
	hs.Profile = profile
	hs.Account = account
	lastVersion := hs.Profile.LastVersion()
   if !launcher.OfflineMode {
      lastType = versionmanager.GetVersionType(hs.Profile.LastVersion())
   } else {
      lastType = "Local"
   }
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
	              Objects[1].(*widget.Label).SetText(lastType+" "+lastVersion)

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

