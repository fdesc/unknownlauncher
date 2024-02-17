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

	"fdesc/unknownlauncher/util/logutil"
	"fdesc/unknownlauncher/launcher"
)

func showGameLog(logPath,gameStdout string,gameStderr error) {
	logutil.Info("Game crashed loading log window")
	logWindow := MainApp.NewWindow("Logs")
	logWindow.Resize(fyne.Size{Height: 300, Width: 600})

	logPathText := canvas.NewText("Game crashed. Latest log path is "+logPath,theme.ForegroundColor())

	crashLog := launcher.GetCrashLog(gameStdout)
	crashMsg := canvas.NewText("Latest crash log:",theme.ForegroundColor())
	crashText := widget.NewTextGridFromString(crashLog)

	stderrMsg := canvas.NewText("Error: "+gameStderr.Error(),theme.ForegroundColor())
	textRectangle := canvas.NewRectangle(color.RGBA{R: 25,G:25, B:25,A: 200})

	closeButton := widget.NewButton("Close",func(){logWindow.Close()})
	launcherLogButton := widget.NewButton("Open the launcher log file",func(){
		launcher.InvokeDefault(filepath.Join(logutil.CurrentLogPath,"launcher_"+logutil.CurrentLogTime+".log"))
	})
	gameStdoutButton := widget.NewButton("Show game output",func(){
		cwd,err := os.Getwd()
		if err != nil { logutil.Error("Failed to get current working directory",err); return }
		file,err := os.Create(filepath.Join(cwd,"logs","latestOutput"))
		if err != nil { logutil.Error("Failed to create temporary file",err); return }
		_,err = file.WriteString(gameStdout)
		if err != nil { logutil.Error("Failed to write into temporary file",err); return }
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

func viewLauncherLogs() {
	logutil.Info("Loading launcher logs to gui")
	logWindow := MainApp.NewWindow("Launcher logs")
	logWindow.Resize(fyne.Size{Height: 600, Width: 600})
	logWindow.SetOnClosed(func(){MainWindow.Show()})
	logWindow.Show()
	logField := widget.NewLabel("")
	logRectangle := canvas.NewRectangle(color.RGBA{R: 25,G:25, B:25,A: 200})
	scrollContainer := container.NewScroll(container.NewStack(logRectangle,logField))
	closeButton := widget.NewButton("Close",func(){
		logWindow.Close()
	})
	openFileButton := widget.NewButton("Open the launcher log file",func(){
		launcher.InvokeDefault(filepath.Join(logutil.CurrentLogPath,"launcher_"+logutil.CurrentLogTime+".log"))
	})
	go func(){
		logutil.Warn("The log viewer is delayed and it can be inaccurate at file integrity checks. It's recommended to check logs from the command output if real time diagnostic is required")
		for {
			if len(logField.Text) > 50000 {
				logField.SetText("")
				logField.SetText("Clearing logs to prevent performance issues its recommended to view the logs from a text editor saved at GAME_DIRECTORY/logs/launcher\n")
			}
			time.Sleep(300 * time.Millisecond)
			logField.SetText(logField.Text+<-logutil.LogChannel)
			scrollContainer.ScrollToBottom()
		}
	}()
	logWindow.SetContent(
		container.NewPadded(
			container.NewBorder(
				nil,
				container.NewHBox(layout.NewSpacer(),closeButton,openFileButton),
				nil,
				nil,
				scrollContainer,
			),
		),
	)
}
