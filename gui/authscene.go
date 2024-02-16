package gui

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"

	"egreg10us/unknownlauncher/auth"
	"egreg10us/unknownlauncher/util/logutil"
)

func NewAccountScene(currentCanvas fyne.Canvas) {
	if lAccounts.LastUsed != "" && previewAccountName == "" {
		lastUsed := lAccounts.Accounts[lAccounts.LastUsed]
		if lastUsed.AccountType == "offline" {
			logutil.Info("Loading the last used account with name: "+lastUsed.Name)
			skinUrl,uuid := auth.PerformOfflineAuthentication(lastUsed.Name)
			if uuid == "" {
				logutil.Warn("Failed to get UUID")
				setCurrentAccountProperties(skinUrl)
				setCurrentProfileProperties()
				return
			}
			if lAccounts.Accounts[lAccounts.LastUsed].Name == lastUsed.Name && lAccounts.Accounts[lAccounts.LastUsed].AccountUUID != uuid {
				delete(lAccounts.Accounts,lAccounts.Accounts[lAccounts.LastUsed].AccountUUID)
			}
			lAccounts.Accounts[uuid] = auth.SaveOfflineAccount(lastUsed.Name,uuid)
			lAccounts.LastUsed = uuid
			lAccounts.SaveToFile()
			setCurrentAccountProperties(skinUrl)
			setCurrentProfileProperties()
			mainScene(currentCanvas)
			return
		}
	}
	welcomeText := widget.NewLabel("Welcome!")
	welcomeText.TextStyle = fyne.TextStyle{Bold:true}
	explanationText := widget.NewLabel("Please select an authentication method.")

	buttonListAccounts := widget.NewButtonWithIcon("List Accounts",theme.AccountIcon(),func(){
		listAccounts(currentCanvas)
	})
	authButtonMS := widget.NewButton("Microsoft Authentication",func(){})
	authButtonMS.Importance = widget.HighImportance
	authButtonNONE := widget.NewButton("No Authentication",func(){ offlineAuthScene(currentCanvas) })

	currentCanvas.SetContent(container.New(
		layout.NewVBoxLayout(),
		layout.NewSpacer(),
		container.NewCenter(welcomeText),
		container.NewCenter(explanationText),
		container.New(layout.NewGridLayout(0), container.NewCenter(container.NewHBox(authButtonMS,authButtonNONE))),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buttonListAccounts, layout.NewSpacer()),
		layout.NewSpacer(),
	))
}

func offlineAuthScene(currentCanvas fyne.Canvas) {
	nameField := widget.NewLabel("Username")
	explanationField := canvas.NewText("Please enter an username for the account.",theme.ForegroundColor())
	nameField.TextStyle = fyne.TextStyle{Bold:true} 
	nameField.Alignment = fyne.TextAlignCenter
	explanationField.Alignment = fyne.TextAlignCenter
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	usernameEntryConfirm := widget.NewButtonWithIcon("Submit",theme.ConfirmIcon(),func() {
		if len(usernameEntry.Text) < 2 || len(usernameEntry.Text) > 16 {
			explanationField.Color = theme.ErrorColor()
			explanationField.Text = "Username cant have length lower than 2 or higher than 16"
			currentCanvas.Refresh(explanationField)
			return
		}
		skinUrl,uuid := auth.PerformOfflineAuthentication(usernameEntry.Text)
		if uuid == "" {
			explanationField.Color = theme.ErrorColor()
			explanationField.Text = "Failed to get UUID please wait and try again later"
			currentCanvas.Refresh(explanationField)
			return
		}
		logutil.Info("Saved offline account with name: "+usernameEntry.Text)
		if lAccounts.Accounts == nil {
			lAccounts.Accounts = make(map[string]auth.AccountProperties)
		}
		for _,v := range lAccounts.Accounts {
			if v.Name == usernameEntry.Text {
				return
			}
		}
		lAccounts.Accounts[uuid] = auth.SaveOfflineAccount(usernameEntry.Text,uuid)
		lAccounts.LastUsed = uuid
		lAccounts.SaveToFile()
		setCurrentAccountProperties(skinUrl)
		setCurrentProfileProperties()
		mainScene(currentCanvas)
		MainWindow.Resize(fyne.Size{Height: 480, Width: 680})
	})
	usernameEntryConfirm.Importance = widget.HighImportance
	usernameEntryCancel := widget.NewButtonWithIcon("Cancel",theme.CancelIcon(),func(){
		NewAccountScene(currentCanvas)
	})

	currentCanvas.SetContent(container.New(
		layout.NewVBoxLayout(),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), nameField, layout.NewSpacer()),
		container.New(layout.NewVBoxLayout(), explanationField, layout.NewSpacer()),
		container.New(layout.NewPaddedLayout(), usernameEntry),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), usernameEntryCancel,usernameEntryConfirm, layout.NewSpacer()),
		layout.NewSpacer(),
	))
}

func listAccounts(currentCanvas fyne.Canvas) {
	logutil.Info("Listing accounts")
	var accountsData = make(map[string]auth.AccountProperties)
	var displayNameSlices []string
        heading := widget.NewRichTextFromMarkdown("# Select an account")
        for _,v := range lAccounts.Accounts {
		accountsData[v.Name] = v
                displayNameSlices = append(displayNameSlices,v.Name)
        }
        data := binding.BindStringList(&displayNameSlices)
        list := widget.NewListWithData(data,
                func() fyne.CanvasObject {
			label := widget.NewLabel("")
                        toolbar := widget.NewToolbar(
                                widget.NewToolbarAction(theme.InfoIcon(), func(){
					headingLabel := widget.NewLabel("Account Properties")
					headingLabel.TextStyle = fyne.TextStyle{Bold:true}
					var modal *widget.PopUp
					modal = widget.NewModalPopUp(
						container.New(
							layout.NewVBoxLayout(),
							headingLabel,
							layout.NewSpacer(),
							widget.NewLabel("Name: "+label.Text),
							widget.NewLabel("Type: "+accountsData[label.Text].AccountType),
							widget.NewLabel("UUID: "+accountsData[label.Text].AccountUUID),
							layout.NewSpacer(),
							container.New(
								layout.NewGridLayoutWithRows(0),
								widget.NewButton("Ok",func(){modal.Hide()}),
							),
						),
						currentCanvas,
					)
					modal.Show()
				}),
                                widget.NewToolbarAction(theme.DeleteIcon(), func(){
					headingLabel := widget.NewLabel("Delete account "+label.Text+"?")
					headingLabel.TextStyle = fyne.TextStyle{Bold:true}
					var modal *widget.PopUp
					modal = widget.NewModalPopUp(
						container.New(
							layout.NewVBoxLayout(),
							headingLabel,
							layout.NewSpacer(),
							widget.NewLabel("This account will be removed from the launcher."),
							layout.NewSpacer(),
							container.New(
								layout.NewGridLayoutWithRows(1),
								widget.NewButton("No",func(){modal.Hide()}),
								widget.NewButton("Yes",func(){
									logutil.Info("Removing account")
									delete(lAccounts.Accounts,accountsData[label.Text].AccountUUID)
									lAccounts.SaveToFile()
									listAccounts(currentCanvas)
									modal.Hide()
								}),
							),
						),
						currentCanvas,
					)
				modal.Show()
				}),
                        )
                        img := canvas.NewImageFromResource(theme.AccountIcon())
                        img.FillMode = canvas.ImageFillOriginal
                        return container.NewPadded(
                                container.NewBorder(nil,nil,img,toolbar,label),
                        )
                },
                func(i binding.DataItem, o fyne.CanvasObject) {
                        o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).Bind(i.(binding.String))
                },
        )
	list.OnSelected = func(i widget.ListItemID) {
		val,_ := data.GetValue(i)
		lAccounts.LastUsed = accountsData[val].AccountUUID
		logutil.Info("Setting the last used account to the account with name: "+lAccounts.Accounts[lAccounts.LastUsed].Name)
		lAccounts.SaveToFile()
		if accountsData[val].AccountType == "offline" {
			skinData,_ := auth.GetSkinData(auth.InitializeClient(),accountsData[val].AccountUUID)
			skinUrl := auth.GetSkinUrl(skinData)
			setCurrentAccountProperties(skinUrl)
			setCurrentProfileProperties()
			mainScene(currentCanvas)
			MainWindow.Resize(fyne.Size{Height: 480, Width: 680})
		}
	}
        newAccountButton := widget.NewButtonWithIcon("Add New",theme.ContentAddIcon(),func(){NewAccountScene(currentCanvas)})
        currentCanvas.SetContent(
                container.NewBorder(
                        container.NewCenter(container.NewPadded(heading)),
                        container.NewHBox(layout.NewSpacer(), container.NewPadded(newAccountButton), layout.NewSpacer()),
                        nil,
                        nil,
                        list,
                ),
        )
}
