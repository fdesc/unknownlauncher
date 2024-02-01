package gui

import (
	"time"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"	
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/container"

	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/gui/resources"
	"egreg10us/faultylauncher/gui/elements"
	"egreg10us/faultylauncher/launcher"
	"egreg10us/faultylauncher/auth"
)

func mainScene(currentCanvas fyne.Canvas) {
	settingsText := widget.NewLabel("Settings")
	settingsImg := canvas.NewImageFromResource(theme.SettingsIcon())
	settingsImg.FillMode = canvas.ImageFillOriginal

	accountsText := widget.NewLabel(previewAccountName)
	if len(previewAccountName) > 6 {
		accountsText = widget.NewLabel(previewAccountName[:5]+"...")
	}
	accountsImg := canvas.NewImageFromResource(resources.UnknownSkinIcon)
	accountsImg.FillMode = canvas.ImageFillOriginal
	if !auth.DefaultSkinIcon {
		accountsImg = canvas.NewImageFromImage(previewAccountSkin)
	}
	accountsButton := widget.NewButton("",func() {
		secondarySkinImg := canvas.NewImageFromResource(resources.UnknownSkinIcon)
		if !auth.DefaultSkinIcon {
			secondarySkinImg = canvas.NewImageFromImage(previewAccountSkin)
		}
		secondarySkinImg.FillMode = canvas.ImageFillOriginal
		nameLabel := widget.NewLabel(previewAccountName)
		nameLabel.TextStyle = fyne.TextStyle{Bold:true}
		var modal *widget.PopUp
		modal = widget.NewModalPopUp(
			container.New(
				layout.NewVBoxLayout(),
				container.NewCenter(widget.NewLabel("Logged in as")),
				secondarySkinImg,
				container.NewCenter(nameLabel),
				layout.NewSpacer(),
				container.New(
					layout.NewGridLayoutWithRows(0),
					widget.NewButton("Ok",func(){modal.Hide()}),	
					widget.NewButton("Change account",func(){
						listAccounts(currentCanvas)
						modal.Hide()
					}),
				),
			),
			currentCanvas,
		)
		modal.Show()
	})

	settingsButton := widget.NewButton("",func(){ settingsScene(currentCanvas) })

	listHeading := widget.NewLabel(previewProfileName)
	listHeading.TextStyle = fyne.TextStyle{Bold:true}
	listContent := widget.NewLabel(previewProfileVersionType+" "+previewProfileVersion)
	listImg := canvas.NewImageFromResource(resources.ProfileIcon)
	listButton := widget.NewButton("",func(){ listProfiles(currentCanvas) })

	playVersionLabel := widget.NewLabel(previewProfileVersion)
	if len(previewProfileVersion) > 6 {
		playVersionLabel = widget.NewLabel(previewProfileVersion[:6]+"...")
	}

	progress := widget.NewProgressBarInfinite()
	progress.Hide()
	progressMsg := canvas.NewText(downloadutil.CurrentFile,theme.ForegroundColor())
	progressMsg.Hide()

	playHeader := widget.NewLabel("Play")
	playHeader.TextStyle = fyne.TextStyle{Bold:true}
	playIcon := canvas.NewImageFromResource(theme.MediaPlayIcon())
	playButton := widget.NewButton("",func(){})
	playContainer := elements.NewRectangleButtonWithIcon(playHeader,playVersionLabel,playIcon,playButton,127)
	playButton.OnTapped = func() {
		launcher.TaskStatus = 0
		playHeader.SetText("Pending")
		playIcon = canvas.NewImageFromResource(theme.MoreHorizontalIcon())
		playContainer.Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0] = playIcon
		playContainer.Refresh()
		playButton.Disable()
		playVersionLabel.Hide()
		progressMsg.Show()	
		progress.Show()
		var exitErr error
		var exitStdout,logPath string
		var procStat *os.ProcessState
		go func() {
			profile := lProfiles.Profiles[lProfiles.LastUsedProfile]
			account := lAccounts.Accounts[previewAccountUUID]
			go func() {
				exitErr,exitStdout,logPath,procStat = launcher.NewLaunchTask(&account,&profile)
				if exitErr != nil {
					showGameLog(logPath,exitStdout,exitErr)
					progressMsg.Hide()
					playHeader.SetText("Play")
					playIcon = canvas.NewImageFromResource(theme.MediaPlayIcon())
					playContainer.Objects[0].(*fyne.Container).
						      Objects[1].(*fyne.Container).
						      Objects[0].(*fyne.Container).
						      Objects[0].(*fyne.Container).
						      Objects[1].(*fyne.Container).
						      Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = playIcon
					playContainer.Refresh()
					playButton.Enable()
					playVersionLabel.Show()
					progress.Hide()
					return
				}
				mainScene(currentCanvas)
			}()
			first := downloadutil.CurrentFile
			progressMsg.Text = first
			progressMsg.Refresh()
			go func() {
				for {
					if downloadutil.CurrentFile != first {
						time.Sleep(time.Millisecond * 250)
						progressMsg.Text = downloadutil.CurrentFile
						progressMsg.Refresh()
					} 
					if launcher.TaskStatus == -1 {
						playVersionLabel.Show()
						progressMsg.Hide()
						progress.Hide()
						switch lSettings.LaunchRule {
						case "Hide":
							MainWindow.Hide()
						case "Exit":
							MainWindow.Hide()
							os.Exit(0)
						case "DoNothing":
							continue
						}
						if procStat != nil && procStat.Exited() {
							MainWindow.Show()
							playHeader.SetText("Play")
							playIcon = canvas.NewImageFromResource(theme.MediaPlayIcon())
							playContainer.Objects[0].(*fyne.Container).
								      Objects[1].(*fyne.Container).
								      Objects[0].(*fyne.Container).
								      Objects[0].(*fyne.Container).
								      Objects[1].(*fyne.Container).
								      Objects[0].(*fyne.Container).
								      Objects[0].(*fyne.Container).
								      Objects[0] = playIcon
							playContainer.Refresh()
							playButton.Enable()
							return 
						}
					}
				}
			}()
		}()
		time.Sleep(500 * time.Millisecond)
	}
	if launcher.TaskStatus == 1 {
		playHeader.SetText("Pending")
		playIcon = canvas.NewImageFromResource(theme.MoreHorizontalIcon())
		playContainer.Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[1].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0].(*fyne.Container).
			      Objects[0] = playIcon
		playContainer.Refresh()
		playButton.Disable()
		playVersionLabel.Hide()
		progress.Show()
		progressMsg.Show()
		go func() {
			for {
				if launcher.TaskStatus == -1 {
					playVersionLabel.Show()
					progressMsg.Hide()
					progress.Hide()
					playHeader.SetText("Play")
					playIcon = canvas.NewImageFromResource(theme.MediaPlayIcon())
					playContainer.Objects[0].(*fyne.Container).
						      Objects[1].(*fyne.Container).
						      Objects[0].(*fyne.Container).
						      Objects[0].(*fyne.Container).
						      Objects[1].(*fyne.Container).
						      Objects[0].(*fyne.Container).
						      Objects[0].(*fyne.Container).
						      Objects[0] = playIcon
					playContainer.Refresh()
					playButton.Enable()
					return 
				}
			}
		}()
	}

	readMoreButton := widget.NewButton("Read More",func(){launcher.InvokeDefault(launcher.ContentReadMoreLink)})
	if launcher.OfflineMode {
		readMoreButton.Hide()
	}
	newsMessage := canvas.NewText(launcher.ContentMessage,theme.ForegroundColor())
	newsMessage.Alignment = fyne.TextAlignCenter
	newsMessage.TextStyle = fyne.TextStyle{Bold:true}
	newsImage := canvas.NewImageFromImage(launcher.ContentImage)
	newsImage.FillMode = canvas.ImageFillOriginal

	currentCanvas.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			layout.NewSpacer(),
			container.NewBorder(
				nil,
				newsMessage,
				nil,
				nil,
				newsImage,
			),
			container.NewHBox(layout.NewSpacer(), readMoreButton, layout.NewSpacer()),
			layout.NewSpacer(),
			container.NewVBox(container.NewPadded(progressMsg),container.NewGridWrap(fyne.NewSize(675,1),progress)),
			container.NewHBox(
				elements.NewSquareButtonWithIcon(accountsText,accountsImg,accountsButton,36),
				elements.NewSquareButtonWithIcon(settingsText,settingsImg,settingsButton,36),
				elements.NewRectangleButtonWithIcon(listHeading,listContent,listImg,listButton,420),
				playContainer,
			),
		),
	)
}
