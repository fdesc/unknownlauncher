package gui

import (
	"strconv"
	"time"

	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/launcher"
	"fdesc/unknownlauncher/gui/elements"
	"fdesc/unknownlauncher/gui/resources"
	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/util/logutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

var lProfiles  = &profilemanager.ProfilesRoot{}
var lAccounts  = &auth.AccountsRoot{}
var lSettings  = &launcher.LauncherSettings{}

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
	return &gui{
		App: gApp,
		Window: gApp.NewWindow(""),
		MainCnt: container.NewStack(),
		Elements: elements.New(),
	}
}

func (g *gui) Start(title string) {
	g.SetContainer(g.Elements.Auth.BaseCnt)
	g.Window.SetTitle(title)
	g.ReloadSettings()
	g.Window.Resize(fyne.Size{Height: 480, Width: 680})
	g.Window.SetFixedSize(true)
	g.Window.SetContent(g.MainCnt)
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

func (g *gui) SetProperties() {
	au := g.Elements.Auth
	ao := g.Elements.AuthOffline
	al := g.Elements.AccountList
	pl := g.Elements.ProfileList
	pe := g.Elements.ProfileEdit
	se := g.Elements.Settings
	hs := g.Elements.HomeScreen
	al.GetAccountFunc = func(name string) auth.AccountProperties {
		return GetAccountFromName(lAccounts.Accounts,name)
	}
	al.DelAccountFunc = func(name string) {
		DeleteAccount(lAccounts,name)
		al.Update(GetAccountNames(lAccounts.Accounts))
	}
	al.SelectAccountFunc = func(uuid string) {
		logutil.Info("Selecting account with the UUID: "+uuid)
		selectedAccount := lAccounts.Accounts[uuid]
		lAccounts.LastUsed = uuid
		lAccounts.SaveToFile()
		if selectedAccount.AccountType == "offline" {
			skinData,_ := auth.GetSkinData(auth.InitializeClient(),selectedAccount.AccountUUID)
			skinUrl := auth.GetSkinUrl(skinData)
			hs.SetSkinIcon(auth.CropSkinImage(skinUrl))
			hs.Update(selectedAccount,lProfiles.Profiles[lProfiles.LastUsedProfile])
			g.SetContainer(hs.BaseCnt)
		}
	}
	ao.AuthFunc = func(name string) error {
		skinimg,err := lAccounts.SaveOfflineAccount(name)
		if err != nil { return err }
		al.Update(GetAccountNames(lAccounts.Accounts))
		hs.SetSkinIcon(skinimg)
		hs.Update(lAccounts.Accounts[lAccounts.LastUsed],lProfiles.Profiles[lProfiles.LastUsedProfile])
		g.SetContainer(hs.BaseCnt)
		ao.ResetEntry()
		return nil
	}
	pl.LookupMapRefresh = func() {
		pl.LookupMap = make(map[string]string)
		for k,v := range lProfiles.Profiles {
			if v.Name != "" {
				pl.LookupMap[v.Name] = k
			} else {
				pl.LookupMap[v.Type] = k
			}
		}
	}
	pl.CreateProfileFunc = func() (profilemanager.ProfileProperties,string) {
		p := profilemanager.ProfileProperties{}
		p.Name = "Profile "+strconv.Itoa(len(lProfiles.Profiles)+1)
		p.Type = "custom-profile"
		p.Created = time.Now().Format(time.RFC3339)
		p.LastUsed = time.Now().Format(time.RFC3339)
		uuid,err := profilemanager.GenerateProfileUUID()
		if err != nil { /**/ }
		return p,uuid
	}
	pl.GetProfileFunc = func(name string) (profilemanager.ProfileProperties,string) {
		uuid := pl.LookupMap[name]
		return lProfiles.Profiles[uuid],uuid
	}
	pl.EditProfileFunc = func(p profilemanager.ProfileProperties,uuid string) {
		pe.Update(&p,uuid)
		g.SetContainer(pe.BaseCnt)
	}
	pl.CopyProfileFunc = func(p profilemanager.ProfileProperties) {
		logutil.Info("Copying profile")
		profile := p
		profile.Name = p.Name+"-copy-"+strconv.Itoa(len(lProfiles.Profiles)+1)
		profile.Type = p.Type+"-copy"
		profile.LastUsed = time.Now().Format(time.RFC3339)
		profile.Created = time.Now().Format(time.RFC3339)
		uuid,err := profilemanager.GenerateProfileUUID()
		if err != nil { return }
		lProfiles.Profiles[uuid] = profile
		lProfiles.SaveToFile()
		pl.Update(GetProfileNames(&lProfiles.Profiles))
		pl.LookupMapRefresh()
	}
	pl.DelProfileFunc = func(name string) {
		logutil.Info("Removing profile")
		uuid := pl.LookupMap[name]
		DeleteProfile(lProfiles,uuid)
		pl.Update(GetProfileNames(&lProfiles.Profiles))
		pl.LookupMapRefresh()
	}
	pl.SelectProfileFunc = func(name string) {
		uuid := pl.LookupMap[name]
		logutil.Info("Selecting profile with UUID: "+uuid)
		lProfiles.LastUsedProfile = uuid
		lProfiles.SaveToFile()
		pl.LookupMapRefresh()
		hs.Update(lAccounts.Accounts[lAccounts.LastUsed],lProfiles.Profiles[uuid])
		g.SetContainer(hs.BaseCnt)
	}
	pe.SaveProfileFunc = func() {
		if lProfiles.ProfileNameExists(pe.Profile.Name,pe.Profile.Type) {
			pe.Profile.Name = pe.Profile.Type+"-"+strconv.Itoa(len(lProfiles.Profiles)+1)
		}
		lProfiles.Profiles[pe.ProfileUUID] = *pe.Profile
		lProfiles.SaveToFile()
		pl.Update(GetProfileNames(&lProfiles.Profiles))
		pl.LookupMapRefresh()
		pe.Update(&profilemanager.ProfileProperties{},"")
		g.SetContainer(pl.BaseCnt)
	}
	pe.BtnCancel.OnTapped = func() {
		pl.Update(GetProfileNames(&lProfiles.Profiles))
		pl.LookupMapRefresh()
		pe.Update(&profilemanager.ProfileProperties{},"")
		g.SetContainer(pl.BaseCnt)
	}
	se.ThemeRadio.OnChanged = func(option string) {
		switch option {
		case "Dark":
			g.App.Settings().SetTheme(&resources.DefaultDarkTheme{})
		case "Light":
			g.App.Settings().SetTheme(&resources.DefaultLightTheme{})
		}
	}
	se.BtnOk.OnTapped = func() {
		lSettings.LauncherTheme = se.ThemeRadio.Selected
		switch se.LaunchRuleSelect.Selected[0] {
		case 'H':
			lSettings.LaunchRule = "Hide"
		case 'E':
			lSettings.LaunchRule = "Exit"
		case 'D':
			lSettings.LaunchRule = "DoNothing"
		}
		lSettings.DisableValidation = se.IntegrityCheck.Checked
		lSettings.SaveToFile()
		g.ReloadSettings()
		se.Update(lSettings.LauncherTheme,lSettings.LaunchRule,lSettings.DisableValidation)
		g.SetContainer(hs.BaseCnt)
	}
	se.BtnCancel.OnTapped = func() {
		g.ReloadSettings()
		se.Update(lSettings.LauncherTheme,lSettings.LaunchRule,lSettings.DisableValidation)
		g.SetContainer(hs.BaseCnt)
	}
	hs.ListAccountsFunc = func() { g.SetContainer(al.BaseCnt) }
	hs.BtnSettings.OnTapped = func() { g.SetContainer(se.BaseCnt) }
	hs.BtnProfile.OnTapped = func() { g.SetContainer(pl.BaseCnt) }
	hs.PopUpCanvas = g.Window.Canvas()
	au.BtnMS.OnTapped      = func() {}
	au.BtnOffline.OnTapped = func() { g.SetContainer(ao.BaseCnt) }
	au.BtnList.OnTapped    = func() { g.SetContainer(al.BaseCnt) }
	ao.BtnCancel.OnTapped  = func() {
		g.SetContainer(au.BaseCnt)
		ao.ResetEntry()
	}
	al.BtnNew.OnTapped     = func() { g.SetContainer(au.BaseCnt) }
	al.Update(GetAccountNames(lAccounts.Accounts))
	al.PopUpCanvas = g.Window.Canvas()
	pl.PopUpCanvas = g.Window.Canvas()
	pl.LookupMapRefresh()
	pl.Update(GetProfileNames(&lProfiles.Profiles))
}
