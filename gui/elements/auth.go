package elements

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/container"
)

type Auth struct {
	BtnOffline     *widget.Button
	BtnMS          *widget.Button
	BtnList        *widget.Button
	BaseCnt        *fyne.Container
}

func NewAuth() *Auth {
	a := &Auth{}
	heading := widget.NewLabel("Hello!")
	heading.TextStyle = fyne.TextStyle{Bold:true}
	label := widget.NewLabel("Please select an authentication method")
	a.BtnOffline = widget.NewButton("No authentication",func(){})
	a.BtnMS = widget.NewButton("Microsoft authenication",func(){})
	a.BtnMS.Importance = widget.HighImportance
	a.BtnList = widget.NewButtonWithIcon("List accounts",theme.AccountIcon(),func(){})
	a.BaseCnt = container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(heading),
		container.NewCenter(label),
		container.New(
			layout.NewGridLayout(0),
			container.NewCenter(container.NewHBox(a.BtnMS,a.BtnOffline)),
		),
		container.NewHBox(
			layout.NewSpacer(),
			a.BtnList,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	)
	return a
}
