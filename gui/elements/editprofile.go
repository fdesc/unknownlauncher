package elements

import (
   "errors"
   "os"
   "strconv"

   "fdesc/unknownlauncher/gui/resources"
   "fdesc/unknownlauncher/launcher"
   "fdesc/unknownlauncher/launcher/profilemanager"
   "fdesc/unknownlauncher/launcher/versionmanager"
   "fdesc/unknownlauncher/util/logutil"

   "github.com/sqweek/dialog"

   "fyne.io/fyne/v2"
   "fyne.io/fyne/v2/canvas"
   "fyne.io/fyne/v2/container"
   "fyne.io/fyne/v2/layout"
   "fyne.io/fyne/v2/theme"
   "fyne.io/fyne/v2/widget"
)

type ProfileEdit struct {
   ProfileUUID             string
   Profile                 *profilemanager.ProfileProperties
   NameEntry               *widget.Entry
   VersionTSelect          *widget.Select
   TypeSlice               []string
   VersionSelect           *widget.Select
   GameDirEntry            *widget.Entry
   SeparateInstallation    *widget.Check
   ResolutionHEntry        *widget.Entry
   ResolutionWEntry        *widget.Entry
   FullscreenCheck         *widget.Check
   JavaExecEntry           *widget.Entry
   JavaArgsEntry           *widget.Entry
   BtnOk                   *widget.Button
   BtnCancel               *widget.Button
   SaveProfileFunc         func()
   BaseCnt                 *fyne.Container
}

func NewProfileEdit() *ProfileEdit {
   pe := &ProfileEdit{}
   pe.Profile = &profilemanager.ProfileProperties{}
   icon := canvas.NewImageFromResource(resources.ProfileIcon)
   icon.FillMode = canvas.ImageFillOriginal
   heading := widget.NewRichTextFromMarkdown("# Edit profile")
   nameLabel := widget.NewLabel("Name")
   nameLabel.TextStyle = fyne.TextStyle{Bold:true}
   versionLabel := widget.NewLabel("Version")
   separatorLine := canvas.NewLine(theme.OverlayBackgroundColor())
   separatorLine.StrokeWidth = 1
   gameDirLabel := widget.NewLabel("Game directory")
   gameDirButton := widget.NewButton("Browse",func(){
      pe.GameDirEntry.Disable()
      filename,err := dialog.Directory().Title("Select game directory").Browse()
      if err != nil && err != dialog.ErrCancelled {
         logutil.Error("Failed to load directory",err)
         pe.GameDirEntry.Enable()
         return
      }
      pe.GameDirEntry.SetText(filename)
      pe.GameDirEntry.Enable()
   })
   resolutionLabel := widget.NewLabel("Resolution")
   javaDirLabel := widget.NewLabel("Java executable ")
   javaDirButton := widget.NewButton("Browse",func(){
      pe.JavaExecEntry.Disable()
      filename,err := dialog.File().Load()
      if err != nil && err != dialog.ErrCancelled {
         logutil.Error("Failed to load file",err)
         pe.JavaExecEntry.Enable()
         return
      }
      pe.JavaExecEntry.SetText(filename)
      pe.JavaExecEntry.Enable()
   })
   javaArgsLabel := widget.NewLabel("JVM arguments")
   pe.NameEntry = widget.NewEntry()
   for k := range versionmanager.VersionList {
      pe.TypeSlice = append(pe.TypeSlice, k)
   }
   if len(pe.TypeSlice) > 2 {
      pe.TypeSlice = versionmanager.SortVersionTypes(pe.TypeSlice)
   }
   pe.VersionTSelect = widget.NewSelect(pe.TypeSlice,func(s string){
      pe.VersionSelect.Options = versionmanager.VersionList[s]
      pe.VersionSelect.Selected = versionmanager.VersionList[s][0]
      pe.VersionSelect.Refresh()
   })
   pe.VersionTSelect.Selected = pe.TypeSlice[0]
   pe.VersionTSelect.Refresh()
   pe.VersionSelect = widget.NewSelect(versionmanager.VersionList[pe.VersionTSelect.Selected],func(string){})
   if !launcher.OfflineMode {
      pe.VersionTSelect.Selected = versionmanager.GetVersionType(pe.Profile.LastVersion())
   }
   pe.GameDirEntry = widget.NewEntry()
   // user should not be able to use the separate installation feature if the game directory entry is empty
   // the Disable() and Enable() functions of the check is provided by fyne api these functions does not change the behaviour of the feature
   pe.GameDirEntry.Validator = fyne.StringValidator(func(path string) error {
      pe.SeparateInstallation.Disable()
      if path == "" {
         return nil
      }
      f,err := os.Open(path)
      if err != nil {
         return err
      }
      fStat,err := f.Stat()
      if err != nil {
         return err
      }
      if fStat.IsDir() {
         pe.SeparateInstallation.Enable()
         return nil
      } else {
         return errors.New("Not a directory")
      }
   })
   pe.SeparateInstallation = widget.NewCheck("Separate installation",func(checked bool){})
   pe.SeparateInstallation.Disable()
   pe.ResolutionHEntry = widget.NewEntry()
   pe.ResolutionHEntry.SetPlaceHolder("Height")
   pe.ResolutionHEntry.Validator = fyne.StringValidator(func(val string) error {
      for _,c := range val {
         switch c {
         case '1','2','3','4','5','6','7','8','9','0':
            continue
         default:
            return errors.New("NaN")
         }
      }
      return nil
   })
   pe.ResolutionWEntry = widget.NewEntry()
   pe.ResolutionWEntry.SetPlaceHolder("Width")
   pe.ResolutionWEntry.Validator = pe.ResolutionHEntry.Validator
   pe.FullscreenCheck = widget.NewCheck("Fullscreen",func(checked bool){
      if checked {
         pe.ResolutionWEntry.Disable()
         pe.ResolutionHEntry.Disable()
      } else {
         pe.ResolutionWEntry.Enable()
         pe.ResolutionHEntry.Enable()
         pe.ResolutionWEntry.Text = "854"
         pe.ResolutionHEntry.Text = "480"
         pe.ResolutionWEntry.FocusGained()
         pe.ResolutionHEntry.FocusGained()
      }
   })
   pe.JavaExecEntry = widget.NewEntry()
   pe.JavaExecEntry.Validator = fyne.StringValidator(func(path string) error {
      if path == "" { return nil }
      f,err := os.Open(path)
      if err != nil { return err }
      fStat,err := f.Stat()
      if err != nil { return err }
      if fStat.IsDir() {
         return errors.New("Not a file")
      } else {
         return nil
      }
   })
   pe.JavaArgsEntry = widget.NewEntry()
   pe.BtnCancel = widget.NewButton("Cancel",func(){})
   pe.BtnOk = widget.NewButton("Save",func() {
      if pe.ResolutionWEntry.Validate() != nil {
         return
      }
      if pe.ResolutionHEntry.Validate() != nil {
         return
      }
      if pe.GameDirEntry.Validate() != nil {
         return
      }
      if pe.JavaExecEntry.Validate() != nil {
         return
      }
      pe.Profile.Name = pe.NameEntry.Text
      pe.Profile.LastGameVersion = pe.VersionSelect.Selected
      if len(pe.JavaArgsEntry.Text) >= 5 || pe.JavaArgsEntry.Text == "" {
         pe.Profile.JVMArgs = pe.JavaArgsEntry.Text
      }
      if pe.FullscreenCheck.Checked {
         pe.Profile.Resolution = &profilemanager.ProfileResolution{Fullscreen:pe.FullscreenCheck.Checked}
      } else if pe.ResolutionHEntry.Text != "" && pe.ResolutionWEntry.Text != "" && !pe.FullscreenCheck.Checked {
         h,err := strconv.Atoi(pe.ResolutionHEntry.Text)
         if err != nil { 
            logutil.Error("Failed to do conversion (string -> int)",err)
            return
         }
         w,err := strconv.Atoi(pe.ResolutionWEntry.Text)
         if err != nil {
            logutil.Error("Failed to do conversion (string -> int)",err)
            return
         }
         pe.Profile.Resolution = &profilemanager.ProfileResolution{Width:w,Height:h,Fullscreen:false}
      } else {
         pe.Profile.Resolution = nil
      }
      pe.Profile.GameDirectory = pe.GameDirEntry.Text
      pe.Profile.JavaDirectory = pe.JavaExecEntry.Text
      pe.Profile.SeparateInstallation = pe.SeparateInstallation.Checked
      pe.SaveProfileFunc()
   })
   pe.BaseCnt = container.NewVBox(
      container.NewPadded(container.NewBorder(nil,nil,icon,nil,heading)),
      nameLabel,
      container.New(&HalfLayout{}, pe.NameEntry),
      versionLabel,
      container.NewPadded(
         container.New(
            layout.NewGridLayoutWithRows(1),
            container.NewHBox(pe.VersionTSelect,pe.VersionSelect),
         ),
      ),
      separatorLine,
      layout.NewSpacer(),
      container.NewBorder(nil,nil,gameDirLabel,container.NewHBox(gameDirButton,pe.SeparateInstallation),pe.GameDirEntry),
      layout.NewSpacer(),
      container.NewBorder(
         nil,
         nil,
         resolutionLabel,
         pe.FullscreenCheck,
         container.New(
            layout.NewGridLayoutWithRows(1),
            pe.ResolutionWEntry,
            pe.ResolutionHEntry,
         ),
      ),
      layout.NewSpacer(),
      container.NewBorder(nil,nil,javaDirLabel,javaDirButton,pe.JavaExecEntry),
      layout.NewSpacer(),
      container.NewBorder(nil,nil,javaArgsLabel,nil,pe.JavaArgsEntry),
      layout.NewSpacer(),
      container.NewCenter(container.NewHBox(pe.BtnCancel,pe.BtnOk)),
   )
   return pe
}

func (pe *ProfileEdit) Update(profile *profilemanager.ProfileProperties,uuid string) {
   pe.Profile = profile
   pe.ProfileUUID = uuid
   pe.NameEntry.SetText("")
   if pe.Profile.Name != "" {
      pe.NameEntry.SetText(pe.Profile.Name)
   }
   if pe.Profile.LastVersion() != "" {
      if !launcher.OfflineMode {
         pe.VersionTSelect.Selected = versionmanager.GetVersionType(pe.Profile.LastVersion())
         pe.VersionSelect.Options = versionmanager.VersionList[versionmanager.GetVersionType(pe.Profile.LastVersion())]
         pe.VersionSelect.Selected = pe.Profile.LastVersion()
      } else {
         pe.VersionTSelect.Selected = pe.TypeSlice[0]
         pe.VersionSelect.Selected = pe.Profile.LastVersion()
      }
   } else {
      pe.VersionTSelect.Selected = pe.TypeSlice[0]
      pe.VersionSelect.Selected = versionmanager.VersionList[pe.VersionTSelect.Selected][0]
   }
   pe.JavaArgsEntry.SetText(pe.Profile.JVMArgs)
   if pe.Profile.Resolution != nil {
      if pe.Profile.Resolution.Fullscreen {
         pe.FullscreenCheck.Checked = pe.Profile.Resolution.Fullscreen
         pe.ResolutionHEntry.Disable()
         pe.ResolutionWEntry.Disable()
      }
      if pe.Profile.Resolution.Width != 0 && pe.Profile.Resolution.Height != 0 {
         pe.ResolutionHEntry.SetText(strconv.Itoa(pe.Profile.Resolution.Height))
         pe.ResolutionWEntry.SetText(strconv.Itoa(pe.Profile.Resolution.Width))
      }
   }
   pe.GameDirEntry.SetText(pe.Profile.GameDirectory)
   pe.JavaExecEntry.SetText(pe.Profile.JavaDirectory)
   pe.JavaArgsEntry.SetText(pe.Profile.JVMArgs)
   pe.SeparateInstallation.Checked = pe.Profile.SeparateInstallation
   if pe.GameDirEntry.Text != "" {
      pe.SeparateInstallation.Enable()
   } else {
      pe.SeparateInstallation.Checked = false
      pe.SeparateInstallation.Disable()
   }
}
