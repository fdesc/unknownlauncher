package elements

import(
   "fyne.io/fyne/v2"
   "fyne.io/fyne/v2/widget"
   "fyne.io/fyne/v2/layout"
   "fyne.io/fyne/v2/container"
)

type Settings struct {
   ThemeRadio       *widget.RadioGroup
   LaunchRuleSelect *widget.Select
   IntegrityCheck   *widget.Check
   BtnOk            *widget.Button
   BtnCancel        *widget.Button
   BaseCnt          *fyne.Container
}

func NewSettings() *Settings {
   s := &Settings{}
   heading := widget.NewRichTextFromMarkdown("# Launcher settings")
   appearanceLabel := widget.NewLabel("Appearance")
   appearanceLabel.TextStyle = fyne.TextStyle{Bold:true}
   themeLabel := widget.NewLabel("Theme style")
   launcherLabel := widget.NewLabel("Launcher")
   launcherLabel.TextStyle = fyne.TextStyle{Bold:true}
   launchRuleLabel := widget.NewLabel("When game starts")
   s.ThemeRadio = widget.NewRadioGroup([]string{"Light","Dark"},func(option string){})
   s.ThemeRadio.Horizontal = true
   s.LaunchRuleSelect = widget.NewSelect([]string{"Hide the launcher","Exit the launcher","Do nothing"},func(string){})
   s.IntegrityCheck = widget.NewCheck("Disable file integrity checks(not recommended)",func(bool){})
   s.BtnOk = widget.NewButton("Save",func(){})
   s.BtnCancel = widget.NewButton("Cancel",func(){})
   s.BaseCnt = container.NewVBox(
      heading,
      appearanceLabel,
      themeLabel,
      container.NewHBox(s.ThemeRadio),
      launcherLabel,
      container.New(
         layout.NewFormLayout(),
         launchRuleLabel,
         container.New(&HalfLayout{},s.LaunchRuleSelect),
      ),
      s.IntegrityCheck,
      layout.NewSpacer(),
      container.NewPadded(
         container.NewHBox(
            layout.NewSpacer(),
            s.BtnCancel,
            s.BtnOk,
         ),
      ),
   )
   return s
}

func (s *Settings) Update(theme string,launchRule string,integrityCheck bool) {
   s.ThemeRadio.Selected = theme
   s.IntegrityCheck.Checked = integrityCheck
   switch launchRule[0] {
   case 'H':
      s.LaunchRuleSelect.Selected = s.LaunchRuleSelect.Options[0]
   case 'E':
      s.LaunchRuleSelect.Selected = s.LaunchRuleSelect.Options[1]
   case 'D':
      s.LaunchRuleSelect.Selected = s.LaunchRuleSelect.Options[2]
   }
}
