package launcher

import (
	"path/filepath"
	"image/jpeg"
	"os/exec"
	"strconv"
	"strings"
	"runtime"
	"regexp"
	"bufio"
	"bytes"
	"image"
	"os"

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/auth"
	"egreg10us/faultylauncher/launcher/profilemanager"
	"egreg10us/faultylauncher/launcher/versionmanager"
	"egreg10us/faultylauncher/launcher/resourcemanager"
	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/util/parseutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
)

var OfflineMode 	bool	
var ContentMessage 	string
var ContentReadMoreLink string
var ContentImage 	image.Image
var TaskStatus	 	int	   // 0 = no tasks, -1 = running task finished, >= 1 = a task is running

type LauncherSettings struct {
	LauncherTheme 	  string
	LaunchRule 	  string
	DisableValidation bool
}

func GetLauncherContent() {
	logutil.Info("Acquiring launcher content")
	const launcherContentMeta = "https://launchercontent.mojang.com"
	contentData,err := downloadutil.GetData(launcherContentMeta+"/news.json")
	if err != nil {
		OfflineMode = true
		logutil.Error("Failed to get launcher content",err)
		logutil.Info("Switching to offline mode")
		ContentMessage = "Offline Mode, no internet connection detected"
		return
	}
	gjson.Get(string(contentData),"entries").ForEach(func(_,value gjson.Result) bool {
		if value.Get("category").String() == "Minecraft: Java Edition" {
			ContentMessage = value.Get("text").String()
			ContentReadMoreLink = value.Get("readMoreLink").String()
			imageData,err := downloadutil.GetData(launcherContentMeta+value.Get("playPageImage").Get("url").String())
			if err != nil { logutil.Error("Failed to get version preview image",err); }
			ContentImage,_ = jpeg.Decode(bytes.NewReader(imageData))
		} else {
			return true
		}
		return false
	})
}

func ReadLauncherSettings() (*LauncherSettings,error) {
	logutil.Info("Reading launcher settings file")
	var file *os.File
	var err error
	file,err = os.Open(filepath.Join(gamepath.Gamedir,"launcher_settings"))
	if err != nil {
		if os.IsNotExist(err) {
			logutil.Info("No launcher_settings file. Creating default settings file")
			file,err = os.Create(filepath.Join(gamepath.Gamedir,"launcher_settings"))
			if err != nil { 
				logutil.Error("Failed to create launcher settings file",err)
				return &LauncherSettings{},err
			}
			_,err = file.WriteString("Theme=Dark\nRule=Hide\nDisableValidation=False")
			if err != nil {
				logutil.Error("Failed to write default settings to settings file",err)
				return &LauncherSettings{},err
			}
		}
	}
	defer file.Close()
	settings := &LauncherSettings{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentLineSplitted := strings.Split(scanner.Text(),"=")
		switch currentLineSplitted[0] {
		case "Theme":
			settings.LauncherTheme = currentLineSplitted[1]
		case "Rule":
			settings.LaunchRule = currentLineSplitted[1]
		case "DisableValidation":
			parsed,err := strconv.ParseBool(currentLineSplitted[1]) 
			if err != nil { continue }
			settings.DisableValidation = parsed
		default:
			continue
		}
	}
	resourcemanager.DisableValidation = settings.DisableValidation
	return settings,err
}

func (ls *LauncherSettings) SaveToFile() error {
	logutil.Info("Saving launcher settings")
	file,err := os.OpenFile(filepath.Join(gamepath.Gamedir,"launcher_settings"),os.O_TRUNC|os.O_WRONLY,os.ModePerm)
	if err != nil { logutil.Error("Failed to open launcher settings file",err); return err }
	defer file.Close()
	switch ls.DisableValidation {
	case true:
		_,err = file.WriteString("Theme="+ls.LauncherTheme+"\n"+"Rule="+ls.LaunchRule+"\n"+"DisableValidation=True")
	case false:
		_,err = file.WriteString("Theme="+ls.LauncherTheme+"\n"+"Rule="+ls.LaunchRule+"\n"+"DisableValidation=False")
	}
	if err != nil { logutil.Error("Failed to write into file",err); return err }
	return err
}

func GetCrashLog(gameStdout string) string {
	findPathRegex := regexp.MustCompile(`[^#]+$`)
	found := findPathRegex.FindAllString(gameStdout,len(gameStdout))
	if len(found) > 1 && found[0] != "" {
		removeWhitespace := regexp.MustCompile(`[^\S]`)
		path := removeWhitespace.ReplaceAllString(found[0],"")
		content,err := os.ReadFile(path)
		if err != nil { logutil.Error("Failed to get contents of crash log",err) }
		return string(content)
	} else {
		return ""
	}
}

func CleanDuplicateNatives() {
	logutil.Info("Cleaning duplicate natives")
	nativesRegex := regexp.MustCompile(`.*-natives-.*`)
	versionDirs,err := os.ReadDir(gamepath.Versionsdir)
	if err != nil { return }
	for _,file := range versionDirs {
		if file.IsDir() {
			os.Chdir(filepath.Join(gamepath.Versionsdir,file.Name()))
			filepath.Walk(filepath.Join(gamepath.Versionsdir,file.Name()),func(path string, native os.FileInfo,walkerr error) error {
				if walkerr == nil && nativesRegex.MatchString(native.Name()) {
					os.RemoveAll(filepath.Join(gamepath.Versionsdir,file.Name(),native.Name()))
				}
				return nil
			})
		}
	}
}

func NewLaunchTask(accountData *auth.AccountProperties,profileData *profilemanager.ProfileProperties) (error,string,string) {
	TaskStatus += 1
	if TaskStatus > 1 {
		return nil,"",""
	}
	logutil.Info("Started job for version "+profileData.LastGameVersion+" as user "+accountData.Name)
	var runtimesPath string
	versionUrl,err := versionmanager.SelectVersion(profileData.LastGameType,profileData.LastGameVersion)
	if err != nil { logutil.Error("Failed to get version",err); return err,"","" }
	versionData,err := versionmanager.ParseVersion(versionUrl)
	if err != nil { logutil.Error("Failed to parse version",err); return err,"","" }
	err = versionmanager.GetVersionArguments(&versionData)
	if err != nil { logutil.Error("Failed to save arguments",err); return err,"","" }
	file,err := os.ReadFile(filepath.Join(gamepath.Assetsdir,"args",profileData.LastGameVersion+".json"))
	if err != nil { logutil.Error("Failed to read arguments for version",err); return err,"","" }
	arguments,err := parseutil.ParseJSON(string(file),false)
	if profileData.SeparateInstallation {
		gamepath.SeparateInstallation = true
		gamepath.Gamedir = profileData.GameDirectory
		gamepath.Reload()
	}
	CleanDuplicateNatives()
	versionString := arguments.Get("id").String()
	if err != nil { logutil.Error("Failed to parse arguments for version",err); return err,"","" }
	resourcemanager.Client(&versionData,versionString)
	if profileData.JavaDirectory != "" {
		runtimesPath = profileData.JavaDirectory
	} else {
		runtimesPath,err = resourcemanager.Runtimes(&versionData)
	}
	if err != nil { logutil.Error("Failed to get runtimes",err); return err,"","" }
	assetsUrl := resourcemanager.GetAssetProperties(&versionData)
	err = resourcemanager.AssetIndex(assetsUrl)
	if err != nil { logutil.Error("Failed to get assets index",err); return err,"","" }
	assetsData,err := resourcemanager.ParseAssets()
	if err != nil { logutil.Error("Failed to parse assets data",err); return err,"","" }
	resourcemanager.Assets(&assetsData)
	logConfigPath := resourcemanager.Log4JConfig(&versionData)
	librariesPath,nativesPath := resourcemanager.Libraries(versionString,&arguments) 
	resourcemanager.CleanLibraryList()
	finalCommand,logPath := generateArguments(accountData,profileData,&arguments,runtimesPath,nativesPath,logConfigPath,librariesPath)
	stdout,err := finalCommand.CombinedOutput()
	runtimesPath = ""
	logutil.Info("Command output is:"+"\n"+string(stdout))
	gamepath.SeparateInstallation = false
	gamepath.Reload()
	if err != nil {
		logutil.Error("Stderr of the command is:",err)
		return err,string(stdout),logPath
	}
	return err,string(stdout),logPath
}

// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func InvokeDefault(url string) error {
	logutil.Info(gamepath.Gamedir)
	logutil.Info("Starting default application for operation")
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open",url).Start()
	case "windows":
		err = exec.Command("rundll32","url.dll","FileProtocolHandler",url).Start()
	case "darwin":
		err = exec.Command("open",url).Start()
	}
	if err != nil { logutil.Error("Failed to invoke default application:",err); return err }
	return err
}

func generateArguments(accountData *auth.AccountProperties,profileData *profilemanager.ProfileProperties,argumentsData *gjson.Result,runtimesPath,nativesPath,logConfigPath string,librariesPath []string) (*exec.Cmd,string) {
	logutil.Info("Generating game arguments")
	var jvmArgs []string
	var gameArgs []string
	var classPath string
	gameLogPath := filepath.Join(gamepath.Gamedir,"logs","latest.log")
	jvmArgs = []string{"-Xdiag","-XX:+UnlockExperimentalVMOptions","-XX:+UseG1GC","-XX:G1NewSizePercent=20","-XX:G1ReservePercent=20","-XX:MaxGCPauseMillis=50","-XX:G1HeapRegionSize=16M"}
	if runtime.GOOS == "darwin" {
		jvmArgs = append(jvmArgs, "-XstartOnFirstThread")
	}
	if runtime.GOARCH == "386" {
		jvmArgs = append(jvmArgs, "-Xss1M")
	}
	if logConfigPath != "" {
		jvmArgs = append(jvmArgs,"-Dlog4j2.formatMsgNoLookups=true","-Djava.library.path="+nativesPath,"-Dlog4j.configurationFile="+logConfigPath,"-Dlog4j.rootLogger=OFF")
	} else {
		jvmArgs = append(jvmArgs,"-Djava.library.path="+nativesPath)
	}
	if profileData.JVMArgs != "" {
		splitRegex := regexp.MustCompile(`[^\s]+`)
		jvmArgs = append(jvmArgs,splitRegex.FindAllString(profileData.JVMArgs,-1)...)
	}
	if runtime.GOOS == "windows" {
		classPath = strings.Join(librariesPath,";")
		classPath = classPath+";"+filepath.Join(gamepath.Versionsdir,profileData.LastGameVersion,profileData.LastGameVersion+".jar")
	} else {
		classPath = strings.Join(librariesPath,":")
		classPath = classPath+":"+filepath.Join(gamepath.Versionsdir,profileData.LastGameVersion,profileData.LastGameVersion+".jar")
	}
	mainClass := argumentsData.Get("mainclass").String()
	if argumentsData.Get("arguments").String() == "default" {
		gameArgs = []string {
			"--username",
			accountData.Name,
			"--version",
			profileData.LastGameVersion,
			"--gameDir",
			gamepath.Gamedir,
			"--assetIndex",
			argumentsData.Get("assets").String(),
			"--uuid",
			accountData.AccountUUID,
			"--accessToken",
			accountData.RefreshToken,
			"--userType",
			accountData.AccountType,
			"--versionType",
			profileData.LastGameType,
		}
		if accountData.AccountType == "offline" {
			gameArgs[11] = "0"
			gameArgs[13] = "mojang"
		}
		if profileData.GameDirectory != "" && !profileData.SeparateInstallation {
			gameArgs[5] = profileData.GameDirectory
		}
		if profileData.Resolution != nil {
			if !profileData.Resolution.Fullscreen {
				gameArgs = append(gameArgs,"--width",strconv.Itoa(profileData.Resolution.Width))
				gameArgs = append(gameArgs,"--height",strconv.Itoa(profileData.Resolution.Height))
			} else {
				gameArgs = append(gameArgs,"--fullscreen")
			}
		}
	} else {
		splitRegex := regexp.MustCompile(`[^\s]+`)
		gameArgs = splitRegex.FindAllString(argumentsData.Get("arguments").String(),-1)
		replaceValues := []string{"${auth_player_name}","${version_name}","${version_type}","${game_directory}","${assets_root}","${game_assets}","${assets_index_name}","${auth_uuid}","${auth_access_token}","${auth_session}","${user_properties}","${user_type}"}
		replaceWith := []string{accountData.Name,profileData.LastGameVersion,profileData.LastGameType,gamepath.Gamedir,gamepath.Assetsdir,gamepath.Assetsdir,argumentsData.Get("assets").String(),accountData.AccountUUID,accountData.RefreshToken,"","{}",accountData.AccountType}
		if accountData.AccountType == "offline" {
			replaceWith[8] = "0"
			replaceWith[9] = "0"
		}
		if profileData.GameDirectory != "" && !profileData.SeparateInstallation {
			gameLogPath = filepath.Join(profileData.GameDirectory,"logs","latest.log")
			replaceWith[3] = profileData.GameDirectory
		}
		for i := 0; i < len(replaceValues); i++ {
			for j := 0; j < len(gameArgs); j++ {
				gameArgs[j] = strings.Replace(gameArgs[j],replaceValues[i],replaceWith[i],len(replaceValues[i]))
			}
		}
		if profileData.Resolution != nil {
			if !profileData.Resolution.Fullscreen {
				gameArgs = append(gameArgs,"--width",strconv.Itoa(profileData.Resolution.Width))
				gameArgs = append(gameArgs,"--height",strconv.Itoa(profileData.Resolution.Height))
			} else {
				gameArgs = append(gameArgs,"--fullscreen")
			}
		}
	}
	if (profileData.JavaDirectory == "" && filepath.Dir(runtimesPath)[len(filepath.Dir(runtimesPath))-1] == 'e') {
		switch runtime.GOOS {
		case "windows":
			runtimesPath = filepath.Join(runtimesPath,"bin","javaw.exe")
		case "linux":
			runtimesPath = filepath.Join(runtimesPath,"bin","java")
		case "darwin":
			runtimesPath = filepath.Join(runtimesPath,"jre.bundle","Contents","Home","bin","java")
		}
	}
	jvmArgs = append(jvmArgs,"-cp",classPath,mainClass)
	jvmArgs = append(jvmArgs,gameArgs...)
	os.Chdir(gamepath.Gamedir)
	logutil.Info("Java path is "+runtimesPath)
	logutil.Info("Generated command arguments:"+"\n"+strings.Join(jvmArgs," "))
	TaskStatus = -1
	return exec.Command(runtimesPath,jvmArgs...),gameLogPath
}
