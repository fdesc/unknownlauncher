package elements

import (
   "slices"

	"fdesc/unknownlauncher/auth"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AccountList struct {
	Data              []string
	BindData          binding.ExternalStringList
	List              *widget.List
	PopUpCanvas       fyne.Canvas
	GetAccountFunc    func(string) auth.AccountProperties
	DelAccountFunc    func(string)
	SelectAccountFunc func(string)
	CItemFunc         func() fyne.CanvasObject
	BtnNew            *widget.Button
	BaseCnt           *fyne.Container
}

func NewAccountList() *AccountList {
	al := &AccountList{}
	var popUp *widget.PopUp
	heading := widget.NewRichTextFromMarkdown("# Select an account")
	icon := canvas.NewImageFromResource(theme.AccountIcon())
	icon.FillMode = canvas.ImageFillOriginal
	al.BindData = binding.BindStringList(&al.Data)
	al.BtnNew = widget.NewButtonWithIcon("Add new",theme.ContentAddIcon(),func(){})
	popUpHeading := widget.NewLabel("")
	popUpHeading.TextStyle = fyne.TextStyle{Bold: true}
	popUpBtnOk := widget.NewButton("Ok",func(){ popUp.Hide() })
	al.CItemFunc = func() fyne.CanvasObject {
		label := widget.NewLabel("")
		toolbar := widget.NewToolbar(
			widget.NewToolbarAction(theme.InfoIcon(),func() {
				popUpHeading.SetText("Account properties")
				popUpBtnOk.SetText("Ok")
				popUp = widget.NewModalPopUp(
					container.NewVBox(
						popUpHeading,
						layout.NewSpacer(),
						widget.NewLabel("Name: "+label.Text),
						widget.NewLabel("Type: "+al.GetAccountFunc(label.Text).AccountType),
						widget.NewLabel("UUID: "+al.GetAccountFunc(label.Text).AccountUUID),
						layout.NewSpacer(),
						container.NewGridWithRows(
							0,
							popUpBtnOk,
						),
					),
					al.PopUpCanvas,
				)
				popUp.Show()
			}),
			widget.NewToolbarAction(theme.DeleteIcon(),func() {
				popUpHeading.SetText("Delete account "+label.Text+"?")
				popUpBtnOk.SetText("No")
				popUp = widget.NewModalPopUp(
					container.NewVBox(
						popUpHeading,
						layout.NewSpacer(),
						widget.NewLabel("This account will be removed from the launcher."),
						layout.NewSpacer(),
						container.NewGridWithRows(
							1,
							popUpBtnOk,
							widget.NewButton("Yes",func(){
								al.DelAccountFunc(al.GetAccountFunc(label.Text).AccountUUID)
								popUp.Hide()
							}),
						),
					),
					al.PopUpCanvas,
				)
				popUp.Show()
			}),
		)
		return container.NewPadded(container.NewBorder(nil,nil,icon,toolbar,label))
	}
	al.List = widget.NewListWithData(
		al.BindData,
		al.CItemFunc,
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).Bind(i.(binding.String))
		},
	)
	al.List.OnSelected = func(id widget.ListItemID) {
		aName,_ := al.BindData.GetValue(id)
		al.SelectAccountFunc(al.GetAccountFunc(aName).AccountUUID)
		al.List.Unselect(id)
	}
	al.BaseCnt = container.NewBorder(
		container.NewCenter(container.NewPadded(heading)),
		container.NewHBox(
			layout.NewSpacer(),
			container.NewPadded(al.BtnNew),
			layout.NewSpacer(),
		),
		nil,
		nil,
		al.List,
	)
	return al
}

func (al *AccountList) Update(data []string) {
   slices.Sort(data)
	al.Data = data
	al.BindData.Reload()
	al.BaseCnt.Objects[0].Refresh()
}
