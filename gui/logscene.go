package gui

import(
	"path/filepath"
	"image/color"
	"time"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/container"

	"egreg10us/faultylauncher/util/logutil"
	"egreg10us/faultylauncher/launcher"
)

func showGameLog(logPath,gameStdout string,gameStderr error) {
	logWindow := MainApp.NewWindow("Logs")
	logWindow.Resize(fyne.Size{Height: 300, Width: 600})

	logPathText := canvas.NewText("Game crashed. Latest log path is "+logPath,theme.ForegroundColor())

	crashLog := launcher.GetCrashLog(gameStdout)
	crashMsg := canvas.NewText("Latest crash log:",theme.ForegroundColor())
	crashText := widget.NewTextGridFromString(crashLog)

	stderrMsg := canvas.NewText("Error: "+gameStderr.Error(),theme.ForegroundColor())
	textRectangle := canvas.NewRectangle(color.RGBA{R: 101,G:101, B:101,A: 20})

	closeButton := widget.NewButton("Close",func(){logWindow.Close()})
	launcherLogButton := widget.NewButton("Show launcher logs",func(){
		launcher.InvokeDefault(logutil.Save(filepath.Dir(filepath.Dir(logPath)),time.Now()))
	})
	gameStdoutButton := widget.NewButton("Show game output",func(){
		cwd,err := os.Getwd()
		if err != nil { logutil.Error("Failed to get current working directory",err) }
		file,err := os.Create(filepath.Join(cwd,"logs","latestOutput"))
		if err != nil { logutil.Error("Failed to create temporary file",err) }
		_,err = file.WriteString(gameStdout)
		if err != nil { logutil.Error("Failed to write into temporary file",err) }
		launcher.InvokeDefault(file.Name())
		file.Close()
	})

	if crashLog != "" {
		logWindow.SetContent(
			container.NewBorder(		
				container.NewPadded(
					container.NewVBox(
						container.NewPadded(logPathText),
						container.NewPadded(stderrMsg),
						container.NewPadded(crashMsg),
					),
				),
				container.NewPadded(container.NewHBox(layout.NewSpacer(), gameStdoutButton,launcherLogButton,closeButton)),
				nil,
				nil,
				container.NewScroll(container.NewPadded(container.NewStack(textRectangle,crashText))),
			),
		)
	} else {
		logWindow.SetContent(
			container.NewBorder(		
				container.NewPadded(container.NewVBox(container.NewPadded(logPathText),container.NewPadded(stderrMsg))),
				container.NewPadded(container.NewHBox(layout.NewSpacer(), gameStdoutButton,launcherLogButton,closeButton)),
				nil,
				nil,
			),
		)
	}
	logWindow.Show()
}
