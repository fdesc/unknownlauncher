package elements

import (
   "image/color"
   "path/filepath"
   "os"

   "fdesc/unknownlauncher/launcher"
   "fdesc/unknownlauncher/util/logutil"

   "fyne.io/fyne/v2"
   "fyne.io/fyne/v2/canvas"
   "fyne.io/fyne/v2/container"
   "fyne.io/fyne/v2/layout"
   "fyne.io/fyne/v2/theme"
   "fyne.io/fyne/v2/widget"
)

type CrashInformer struct {
   InfoWindow     fyne.Window
}

func NewCrashInformer() *CrashInformer {
   return &CrashInformer{}
}

func (cinf *CrashInformer) Start(err error,output,logPath string) {
   cinf.InfoWindow.Resize(fyne.Size{Height:300,Width:600})
   logPathText := canvas.NewText("Game crashed. latest log path is "+logPath,theme.PlaceHolderColor())
   crashText := widget.NewTextGridFromString("")
   crashLog := launcher.GetCrashLog(output)
   if crashLog != "" {
      crashText.SetText(crashLog)
   }
   errorMessage := canvas.NewText("Error: "+err.Error(),theme.PlaceHolderColor())
   textRect := canvas.NewRectangle(color.RGBA{R:25,G:25,B:25,A:200})

   closeButton := widget.NewButton("Close",func(){ cinf.InfoWindow.Hide() })
   launcherLogButton := widget.NewButton("View launcher logs",func(){
      launcher.InvokeDefault(filepath.Join(logutil.CurrentLogPath,"launcher_"+logutil.CurrentLogDate+".log"))
   })
   gameOutputButton := widget.NewButton("Show game output",func(){
      cwd,err := os.Getwd()
      if err != nil { logutil.Error("Failed to get current working directory",err); return }
      file,err := os.Create(filepath.Join(cwd,"latestOutput"))
      if err != nil { logutil.Error("Failed to create temporary file",err); return }
      _,err = file.WriteString(output)
      if err != nil { logutil.Error("Failed to write into temporary file",err); return }
      launcher.InvokeDefault(file.Name())
      file.Close()
   })
   if crashLog != "" {
      cinf.InfoWindow.SetContent(
         container.NewBorder(container.NewPadded(container.NewVBox(
            container.NewPadded(logPathText),
            container.NewPadded(errorMessage),
            ),
            ),
            container.NewPadded(
               container.NewHBox(
                  layout.NewSpacer(),
                  widget.NewButton("Copy crash log to clipboard",func(){ cinf.InfoWindow.Clipboard().SetContent(crashLog) }),
                  gameOutputButton,
                  launcherLogButton,
                  closeButton,
                  ),
               ),
            nil,
            nil,
            container.NewScroll(container.NewPadded(container.NewStack(textRect,crashText))),
            ),
         )
   } else {
      cinf.InfoWindow.SetContent(
         container.NewBorder(container.NewPadded(container.NewVBox(
            container.NewPadded(logPathText),
            container.NewPadded(errorMessage),
            ),
            ),
            container.NewPadded(container.NewHBox(layout.NewSpacer(),gameOutputButton,launcherLogButton,closeButton)),
            nil,
            nil,
            nil,
            ),
         )
   }
   cinf.InfoWindow.Show()
}
