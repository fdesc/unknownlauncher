package elements

import (

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AuthOffline struct {
	Entry     *widget.Entry
	BtnOk     *widget.Button
	BtnCancel *widget.Button
	AuthFunc  func(string) error
	BaseCnt   *fyne.Container
}

func NewAuthOffline() *AuthOffline {
	ao := &AuthOffline{}
	heading := widget.NewLabel("Username")
	heading.TextStyle = fyne.TextStyle{Bold:true}
	label := canvas.NewText("Please select an username for the account",theme.ForegroundColor())
	ao.Entry = widget.NewEntry()
	ao.BtnOk = widget.NewButton("Ok",func(){
		if len(ao.Entry.Text) < 2 || len(ao.Entry.Text) > 16 {
			label.Color = theme.ErrorColor()
			label.Text = "Username cant have length lower than 2 or higher than 16"
			label.Refresh()
			return
		}
		err := ao.AuthFunc(ao.Entry.Text)
		if err != nil {
			label.Color = theme.ErrorColor()
			label.Text = err.Error()
			label.Refresh()
		}
	})
	ao.BtnOk.Importance = widget.HighImportance
	ao.BtnCancel = widget.NewButton("Cancel",func(){})
	ao.BaseCnt = container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			heading,
			layout.NewSpacer(),
		),
		container.NewVBox(container.NewCenter(label)),
		container.NewPadded(ao.Entry),
		container.NewHBox(
			layout.NewSpacer(),
			ao.BtnCancel,
			ao.BtnOk,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	)
	return ao
}

func (ao *AuthOffline) ResetEntry() {
	ao.Entry.SetText("")
}
