package elements

import (
   "slices"

   "fdesc/unknownlauncher/gui/resources"
   "fdesc/unknownlauncher/launcher/profilemanager"

   "fyne.io/fyne/v2"
   "fyne.io/fyne/v2/canvas"
   "fyne.io/fyne/v2/container"
   "fyne.io/fyne/v2/data/binding"
   "fyne.io/fyne/v2/layout"
   "fyne.io/fyne/v2/theme"
   "fyne.io/fyne/v2/widget"
)

type ProfileList struct {
   LookupMap         map[string]string
   LookupMapRefresh  func()
   Data              []string
   BindData          binding.ExternalStringList
   List              *widget.List
   BtnNew            *widget.Button
   PopUpCanvas       fyne.Canvas
   EditProfileFunc   func(profilemanager.ProfileProperties,string)
   CopyProfileFunc   func(profilemanager.ProfileProperties)
   CreateProfileFunc func() (profilemanager.ProfileProperties,string)
   GetProfileFunc    func(string) (profilemanager.ProfileProperties,string)
   SelectProfileFunc func(string)
   DelProfileFunc    func(string)
   CItemFunc         func() fyne.CanvasObject
   BaseCnt           *fyne.Container
}

func NewProfileList() *ProfileList {
   pl := &ProfileList{}
   icon := canvas.NewImageFromResource(resources.ProfileIcon)
   icon.FillMode = canvas.ImageFillOriginal
   pl.BindData = binding.BindStringList(&pl.Data)
   pl.BtnNew = widget.NewButtonWithIcon("Add New",theme.ContentAddIcon(),func(){
      pl.EditProfileFunc(pl.CreateProfileFunc())
   })
   pl.CItemFunc = func() fyne.CanvasObject {
      label := widget.NewLabel("")
      toolbar := widget.NewToolbar(
         widget.NewToolbarAction(theme.DocumentCreateIcon(),func(){
            pl.EditProfileFunc(pl.GetProfileFunc(label.Text))
         }),
         widget.NewToolbarAction(theme.ContentCopyIcon(),func(){
            p,_ := pl.GetProfileFunc(label.Text)
            pl.CopyProfileFunc(p)
         }),
         widget.NewToolbarAction(theme.DeleteIcon(),func() {
            var modal *widget.PopUp
            heading := widget.NewLabel("Remove profile?")
            heading.TextStyle = fyne.TextStyle{Bold:true}
            modal = widget.NewModalPopUp(
               container.NewVBox(
                  heading,
                  layout.NewSpacer(),
                  widget.NewLabel("This action cannot be undone."),
                  layout.NewSpacer(),
                  container.NewGridWithRows(
                     1,
                     widget.NewButton("Cancel",func(){ modal.Hide() }),
                     widget.NewButton("Confirm",func(){
                        pl.DelProfileFunc(label.Text)
                        modal.Hide()
                     }),
                  ),
               ),
               pl.PopUpCanvas,
            )
            modal.Show()
         }),
      )
      pl.List.OnSelected = func(id widget.ListItemID) {
         pName,_ := pl.BindData.GetValue(id)
         pl.SelectProfileFunc(pName)
         pl.List.Unselect(id)
      }
      box := container.NewPadded(container.NewBorder(nil,nil,icon,toolbar,label))
      return box
   }
   pl.List = widget.NewListWithData (
      pl.BindData,
      pl.CItemFunc,
      func(i binding.DataItem, o fyne.CanvasObject) {
         o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).Bind(i.(binding.String))
      },
   )
   pl.BaseCnt = container.NewPadded(
      container.NewBorder(
         nil,
         container.NewCenter(container.NewHBox(pl.BtnNew)),
         nil,
         nil,
         pl.List,
      ),
   )
   return pl
}

func (pl *ProfileList) Update(data []string) {
   slices.Sort(data)
   pl.Data = data
   pl.BindData.Reload()
   pl.BaseCnt.Objects[0].(*fyne.Container).Objects[0].Refresh()
}

