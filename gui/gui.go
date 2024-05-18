package gui

import (
	"strconv"
   "image/color"
	"time"
	"os"

	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/gui/elements"
	"fdesc/unknownlauncher/gui/resources"
	"fdesc/unknownlauncher/launcher"
	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/util/logutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var lProfiles = &profilemanager.ProfilesRoot{}
var lAccounts = &auth.AccountsRoot{}
var lSettings = &launcher.LauncherSettings{}

type gui struct {
	App      fyne.App
	Window   fyne.Window
	MainCnt  *fyne.Container
	Elements *elements.GuiElements
}

func SetProfilesRoot(p *profilemanager.ProfilesRoot) {
	lProfiles = p
}

func SetAccountsRoot(a *auth.AccountsRoot) {
	lAccounts = a
}

func SetSettings(s *launcher.LauncherSettings) {
	lSettings = s
}

func NewGui() *gui {
	gApp := app.New()
   gui := &gui{
		App: gApp,
		Window: gApp.NewWindow(""),
		MainCnt: container.NewStack(),
		Elements: elements.New(),
	}
   gui.ReloadSettings()
   return gui
}

func (g *gui) Start(title string) {
	g.Window.SetTitle(title)
   g.Window.SetMaster()
	g.Window.Resize(fyne.Size{Height: 480, Width: 680})
	g.Window.SetFixedSize(true)
	g.Window.SetContent(g.MainCnt)
   g.Elements.CrashInformer.InfoWindow = g.App.NewWindow("Logs")
   g.Elements.CrashInformer.InfoWindow.Hide()
   if lAccounts.LastUsed().Name != "" {
      skinData,_ := auth.GetSkinData(auth.InitClient(),lAccounts.LastUsed().AccountUUID)
      skinUrl := auth.GetSkinUrl(skinData)
      g.Elements.HomeScreen.Update(lAccounts.LastUsed(),lProfiles.LastUsed())
      g.Elements.HomeScreen.SetSkinIcon(auth.CropSkinImage(skinUrl))
      g.SetContainer(g.Elements.HomeScreen.BaseCnt)
   } else {
      g.SetContainer(g.Elements.Auth.BaseCnt)
   }
	g.Window.ShowAndRun()
}

func (g *gui) SetContainer(objects ...fyne.CanvasObject) {
	if len(g.MainCnt.Objects) == 0 {
		g.MainCnt.Objects = append(g.MainCnt.Objects,objects[0])
	} else {
		g.MainCnt.Objects[0] = objects[0]
	}
	g.MainCnt.Refresh()
}

func (g *gui) ReloadSettings() {
	var err error
	lSettings,err = launcher.ReadLauncherSettings()
	if err != nil {
		logutil.Error("Failed to read launcher settings",err)
		return
	}
	if lSettings.LauncherTheme == "Light" {
		g.App.Settings().SetTheme(&resources.DefaultLightTheme{})
	} else {
		g.App.Settings().SetTheme(&resources.DefaultDarkTheme{})
	}
	g.Elements.Settings.Update(lSettings.LauncherTheme,lSettings.LaunchRule,lSettings.DisableValidation)
}

func InitialError(message string,err error) {
   smallApp := app.New()
   window := smallApp.NewWindow("Application crashed")
   window.Resize(fyne.Size{Height:480,Width:680})
   messageRect := canvas.NewRectangle(color.RGBA{R:25,G:25,B:25,A:200})
   messageLabel := widget.NewLabel("Application failed to start. To troubleshoot check structure of json files in game directory.")
   textField := widget.NewTextGridFromString(message+"\n"+err.Error())
   quitButton := widget.NewButton("Quit",func() { window.Close(); smallApp.Quit() })
   window.SetContent(
      container.NewBorder(
         messageLabel,
         quitButton,
         nil,
         nil,
         container.NewScroll(
            container.NewStack(messageRect,textField),
         ),
      ),
   )
   window.ShowAndRun()
}

func (g *gui) SetProperties() {
	g.bindAuth()
	g.bindAuthOffline()
	g.bindAccountList()
	g.bindProfileList()
	g.bindProfileEdit()
	g.bindSettings()
	g.bindHomeScreen()
}

func (g *gui) bindHomeScreen() {
	g.Elements.HomeScreen.AppExitFunc = func() {
		g.Elements.HomeScreen.WindowHideFunc()
		g.App.Quit()
		os.Exit(0)
	}
	g.Elements.HomeScreen.WindowHideFunc = func() { g.Window.Hide() }
	g.Elements.HomeScreen.WindowShowFunc = func() { g.Window.Show() }
	g.Elements.HomeScreen.LaunchRuleFunc = func() string { return lSettings.LaunchRule }
	g.Elements.HomeScreen.ListAccountsFunc = func() { g.SetContainer(g.Elements.AccountList.BaseCnt) }
	g.Elements.HomeScreen.BtnSettings.OnTapped = func() { g.SetContainer(g.Elements.Settings.BaseCnt) }
	g.Elements.HomeScreen.BtnProfile.OnTapped = func() { g.SetContainer(g.Elements.ProfileList.BaseCnt) }
   g.Elements.HomeScreen.CrashInformerFunc = func(err error, output string, logPath string) {
      g.Elements.CrashInformer.Start(err,output,logPath)
   }
	g.Elements.HomeScreen.PopUpCanvas = g.Window.Canvas()
}

func (g *gui) bindSettings() {
	g.Elements.Settings.BtnOk.OnTapped = func() {
		lSettings.LauncherTheme = g.Elements.Settings.ThemeRadio.Selected
		switch g.Elements.Settings.LaunchRuleSelect.Selected[0] {
		case 'H':
			lSettings.LaunchRule = "Hide"
		case 'E':
			lSettings.LaunchRule = "Exit"
		case 'D':
			lSettings.LaunchRule = "DoNothing"
		}
		lSettings.DisableValidation = g.Elements.Settings.IntegrityCheck.Checked
		lSettings.SaveToFile()
		g.ReloadSettings()
		g.Elements.Settings.Update(lSettings.LauncherTheme,lSettings.LaunchRule,lSettings.DisableValidation)
		g.SetContainer(g.Elements.HomeScreen.BaseCnt)
	}
	g.Elements.Settings.ThemeRadio.OnChanged = func(option string) {
		switch option {
		case "Dark":
			g.App.Settings().SetTheme(&resources.DefaultDarkTheme{})
		case "Light":
			g.App.Settings().SetTheme(&resources.DefaultLightTheme{})
		}
	}
	g.Elements.Settings.BtnCancel.OnTapped = func() {
		g.ReloadSettings()
		g.Elements.Settings.Update(lSettings.LauncherTheme,lSettings.LaunchRule,lSettings.DisableValidation)
		g.SetContainer(g.Elements.HomeScreen.BaseCnt)
	}
}

func (g *gui) bindProfileEdit() {
	g.Elements.ProfileEdit.SaveProfileFunc = func() {
		lProfiles.AddProfile(g.Elements.ProfileEdit.Profile,g.Elements.ProfileEdit.ProfileUUID)
		if lProfiles.ProfileNameExists(g.Elements.ProfileEdit.Profile.Name,g.Elements.ProfileEdit.Profile.Type) {
			savedProfile := lProfiles.GetProfile(g.Elements.ProfileEdit.ProfileUUID)
			savedProfile.Name = savedProfile.Name+"-"+strconv.Itoa(len(lProfiles.Profiles)+1)
			lProfiles.AddProfile(&savedProfile,g.Elements.ProfileEdit.ProfileUUID)
		}
		lProfiles.SaveToFile()
		g.Elements.ProfileList.Update(lProfiles.GetProfileNames())
		g.Elements.ProfileList.LookupMapRefresh()
		g.Elements.ProfileEdit.Update(&profilemanager.ProfileProperties{},"")
		g.SetContainer(g.Elements.ProfileList.BaseCnt)
	}
	g.Elements.ProfileEdit.BtnCancel.OnTapped = func() {
		g.Elements.ProfileList.Update(lProfiles.GetProfileNames())
		g.Elements.ProfileList.LookupMapRefresh()
		g.Elements.ProfileEdit.Update(&profilemanager.ProfileProperties{},"")
		g.SetContainer(g.Elements.ProfileList.BaseCnt)
	}
}

func (g *gui) bindProfileList() {
	g.Elements.ProfileList.CopyProfileFunc = func(p profilemanager.ProfileProperties) {
		logutil.Info("Copying profile")
		profile := p
		profile.Name = p.Name+"-copy-"+strconv.Itoa(len(lProfiles.Profiles)+1)
		profile.Type = p.Type+"-copy"
		profile.LastUsed = time.Now().Format(time.RFC3339)
		profile.Created = time.Now().Format(time.RFC3339)
		uuid,err := profilemanager.GenerateProfileUUID()
		if err != nil { return }
		lProfiles.AddProfile(&profile,uuid)
		lProfiles.SaveToFile()
		g.Elements.ProfileList.Update(lProfiles.GetProfileNames())
		g.Elements.ProfileList.LookupMapRefresh()
	}
	g.Elements.ProfileList.SelectProfileFunc = func(name string) {
		uuid := g.Elements.ProfileList.LookupMap[name]
		logutil.Info("Selecting profile with UUID: "+uuid)
      selectedProfile := lProfiles.GetProfile(uuid)
      selectedProfile.LastUsed = time.Now().Format(time.RFC3339)
		lProfiles.LastUsedProfile = uuid
		lProfiles.SaveToFile()
		g.Elements.ProfileList.LookupMapRefresh()
		g.Elements.HomeScreen.Update(lAccounts.LastUsed(),lProfiles.GetProfile(uuid))
		g.SetContainer(g.Elements.HomeScreen.BaseCnt)
	}
	g.Elements.ProfileList.CreateProfileFunc = func() (profilemanager.ProfileProperties,string) {
		p := profilemanager.ProfileProperties{}
		p.Name = "profile "+strconv.Itoa(len(lProfiles.Profiles)+1)
		p.Type = "custom-profile"
		p.Created = time.Now().Format(time.RFC3339)
		p.LastUsed = time.Now().Format(time.RFC3339)
		uuid,_ := profilemanager.GenerateProfileUUID()
		return p,uuid
	}
	g.Elements.ProfileList.LookupMapRefresh = func() {
		g.Elements.ProfileList.LookupMap = make(map[string]string)
		for k,v := range lProfiles.Profiles {
			if v.Name != "" {
				g.Elements.ProfileList.LookupMap[v.Name] = k
			} else {
				g.Elements.ProfileList.LookupMap[v.Type] = k
			}
		}
	}
	g.Elements.ProfileList.DelProfileFunc = func(name string) {
		logutil.Info("Removing profile")
		uuid := g.Elements.ProfileList.LookupMap[name]
		lProfiles.DeleteProfile(uuid)
		lProfiles.SaveToFile()
		g.Elements.ProfileList.Update(lProfiles.GetProfileNames())
		g.Elements.ProfileList.LookupMapRefresh()
	}
	g.Elements.ProfileList.GetProfileFunc = func(name string) (profilemanager.ProfileProperties,string) {
		uuid := g.Elements.ProfileList.LookupMap[name]
		return lProfiles.GetProfile(uuid),uuid
	}
	g.Elements.ProfileList.EditProfileFunc = func(p profilemanager.ProfileProperties,uuid string) {
		g.Elements.ProfileEdit.Update(&p,uuid)
		g.SetContainer(g.Elements.ProfileEdit.BaseCnt)
	}
	g.Elements.ProfileList.PopUpCanvas = g.Window.Canvas()
	g.Elements.ProfileList.LookupMapRefresh()
	g.Elements.ProfileList.Update(lProfiles.GetProfileNames())
}

func(g *gui) bindAuthOffline() {
	g.Elements.AuthOffline.AuthFunc = func(name string) error {
		skinimg,err := lAccounts.SaveOfflineAccount(name)
		if err != nil { return err }
		g.Elements.AccountList.Update(lAccounts.GetAccountNames())
		g.Elements.HomeScreen.SetSkinIcon(skinimg)
		g.Elements.HomeScreen.Update(lAccounts.LastUsed(),lProfiles.LastUsed())
		g.SetContainer(g.Elements.HomeScreen.BaseCnt)
		g.Elements.AuthOffline.ResetFields()
		return nil
	}
	g.Elements.AuthOffline.BtnCancel.OnTapped  = func() {
		g.SetContainer(g.Elements.Auth.BaseCnt)
		g.Elements.AuthOffline.ResetFields()
	}
}

func (g *gui) bindAccountList() {
	g.Elements.AccountList.SelectAccountFunc = func(uuid string) {
		logutil.Info("Selecting account with UUID: "+uuid)
		selectedAccount := lAccounts.GetAccount(uuid)
		lAccounts.LastUsedAccount = uuid
		lAccounts.SaveToFile()
      lastProfile := lProfiles.LastUsed()
		if selectedAccount.AccountType == "offline" {
			skinData,_ := auth.GetSkinData(auth.InitClient(),selectedAccount.AccountUUID)
			skinUrl := auth.GetSkinUrl(skinData)
			g.Elements.HomeScreen.SetSkinIcon(auth.CropSkinImage(skinUrl))
			g.Elements.HomeScreen.Update(selectedAccount,lastProfile)
			g.SetContainer(g.Elements.HomeScreen.BaseCnt)
		}
	}
	g.Elements.AccountList.DelAccountFunc = func(name string) {
		lAccounts.DeleteAccount(name)
		lAccounts.SaveToFile()
		g.Elements.AccountList.Update(lAccounts.GetAccountNames())
	}
	g.Elements.AccountList.GetAccountFunc = func(name string) auth.AccountProperties {
		return lAccounts.GetAccountFromName(name)
	}
	g.Elements.AccountList.Update(lAccounts.GetAccountNames())
	g.Elements.AccountList.PopUpCanvas = g.Window.Canvas()
	g.Elements.AccountList.BtnNew.OnTapped = func() { g.SetContainer(g.Elements.Auth.BaseCnt) }
}

func (g *gui) bindAuth() {
	g.Elements.Auth.BtnMS.OnTapped      = func() {}
	g.Elements.Auth.BtnOffline.OnTapped = func() { g.SetContainer(g.Elements.AuthOffline.BaseCnt) }
	g.Elements.Auth.BtnList.OnTapped    = func() { g.SetContainer(g.Elements.AccountList.BaseCnt) }
}
