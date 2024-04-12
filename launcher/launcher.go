package launcher

import (
	"path/filepath"
	"os/exec"
	"strconv"
	"strings"
	"runtime"
	"regexp"
	"bufio"
	"os"

	"github.com/tidwall/gjson"
	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/launcher/versionmanager"
	"fdesc/unknownlauncher/launcher/resourcemanager"
	"fdesc/unknownlauncher/util/downloadutil"
	"fdesc/unknownlauncher/util/gamepath"
	"fdesc/unknownlauncher/util/logutil"
)

var OfflineMode         bool
var ContentMessage      string
var ContentReadMoreLink string
var ContentImageLink    string
var TaskStatus          int         // 0 = no tasks, -1 = running task finished, >= 1 = a task is running

type LauncherSettings struct {
	LauncherTheme     string
	LaunchRule        string
	DisableValidation bool
}

type LaunchTask struct {
	Profile     *profilemanager.ProfileProperties
	Account     *auth.AccountProperties
	JavaPath    string
	ClassPath   string
	MainClass   string
	NativesPath string
	LogPath     string
	AssetID     string
        LibsPath    []string
	JavaArgs    []string
	GameArgs    []string
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
			ContentImageLink = launcherContentMeta+value.Get("playPageImage").Get("url").String()
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
			os.Remove(filepath.Join(gamepath.Gamedir,"launcher_settings"))
			return ReadLauncherSettings()
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
	var found string
	scanner := bufio.NewScanner(strings.NewReader(gameStdout))
	for scanner.Scan() {
		if len(scanner.Text()) > 5 && scanner.Text()[:5] == "#@!@#" {
			findPath := regexp.MustCompile(`(/|C:\\).+`)
			found = findPath.FindAllString(scanner.Text(),len(scanner.Text()))[0]
		}
	}
	if found != "" {
		content,err := os.ReadFile(found)
		if err != nil {
			logutil.Error("Failed to read crash log",err)
			return ""
		}
		return string(content)
	}
	return ""
}

func cleanDuplicateNatives() {
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

func (t *LaunchTask) Prepare() error {
	logutil.Info("Started job for version "+t.Profile.LastVersion()+" as user "+t.Account.Name)
	metaUrl,err := versionmanager.SelectVersion(t.Profile.LastType(),t.Profile.LastVersion())
	if err != nil { return err }
	versionData,err := versionmanager.ParseVersion(metaUrl)
	if err != nil { return err }
	err = versionmanager.GetVersionArguments(&versionData)
	if err != nil { return err }
	content,err := os.ReadFile(filepath.Join(gamepath.Assetsdir,"args",t.Profile.LastVersion()+".json"))
	if err != nil { logutil.Error("Failed to read version arguments file",err); return err }
	argsdata := gjson.Parse(string(content))
	versionId := argsdata.Get("id").String()
	jvmType := argsdata.Get("jvmtype").String()
	t.AssetID = argsdata.Get("assets").String()
	t.MainClass = argsdata.Get("mainclass").String()
	t.GameArgs = strings.Split(argsdata.Get("arguments").String()," ")
	if t.Profile.SeparateInstallation {
		gamepath.SeparateInstallation = true
		gamepath.Gamedir = t.Profile.GameDirectory
		gamepath.Reload()
	}
	cleanDuplicateNatives()
	err = resourcemanager.Client(&versionData,versionId)
	if err != nil { return err }
	if t.Profile.JavaDirectory != "" {
		t.JavaPath = t.Profile.JavaDirectory
	} else if _,err := os.Stat(filepath.Join(gamepath.Runtimesdir,jvmType)); err != nil {
		var err error
		t.JavaPath,err = resourcemanager.Runtimes(&versionData)
		if err != nil { return err }
	} else {
		t.JavaPath = filepath.Join(gamepath.Runtimesdir,jvmType)
	}
	err = resourcemanager.AssetIndex(resourcemanager.GetAssetProperties(&versionData))
	if err != nil { return err }
	assetsData,err := resourcemanager.ParseAssets()
	if err != nil { return err }
	resourcemanager.Assets(&assetsData)
	t.LogPath = resourcemanager.Log4JConfig(&versionData)
	t.LibsPath,t.NativesPath = resourcemanager.Libraries(versionId,&argsdata)
	resourcemanager.CleanLibraryList()
	return err
}

func (t *LaunchTask) CompleteArguments() (*exec.Cmd) {
	t.buildClassPath()
	t.buildGameArguments()
	t.buildJvmArgs()
	os.Chdir(gamepath.Gamedir)
	logutil.Info("Java path is "+t.JavaPath)
	previewArgs := make([]string,len(t.JavaArgs))
	copy(previewArgs,t.JavaArgs)
	for i := range previewArgs {
		switch previewArgs[i] {
		case "-cp":
			if runtime.GOOS == "windows" {
				previewArgs[i+1] = strings.Replace(previewArgs[i+1],";","\n | ",len(previewArgs[i+1]))
			} else {
				previewArgs[i+1] = strings.Replace(previewArgs[i+1],":","\n | ",len(previewArgs[i+1]))
			}
		case "--accessToken":
			previewArgs[i+1] = "HIDDEN"
		case "--session":
			previewArgs[i+1] = "HIDDEN"
		default:
			continue
		}
	}
	logutil.Info("Generated command arguments:"+"\n"+strings.Join(previewArgs,"\n"))
	gamepath.SeparateInstallation = false
	gamepath.Reload()
	return exec.Command(t.JavaPath,t.JavaArgs...)
}

func (t *LaunchTask) buildGameArguments() {
	if t.GameArgs[0] == "default" {
		t.GameArgs = nil
		t.GameArgs = append(
			t.GameArgs,
			"--username",
			t.Account.Name,
			"--version",
			t.Profile.LastVersion(),
			"--gameDir",
			gamepath.Gamedir,
			"--assetIndex",
			t.AssetID,
			"--uuid",
			t.Account.AccountUUID,
			"--accessToken",
			t.Account.RefreshToken,
			"--userType",
			t.Account.AccountType,
			"--versionType",
			t.Profile.LastType(),
		)
		if t.Account.AccountType == "offline" {
			t.GameArgs[11] = "0"
			t.GameArgs[13] = "mojang"
		}
		if t.Profile.GameDirectory != "" && !t.Profile.SeparateInstallation {
			t.GameArgs[5] = t.Profile.GameDirectory
		}
	} else {
		var ArgMap = map[string]string {
			"${auth_player_name}":t.Account.Name,
			"${auth_session}":"0",
			"${version_name}":t.Profile.LastVersion(),
			"${game_directory}":gamepath.Gamedir,
			"${game_assets}":gamepath.Assetsdir,
			"${assets_root}":gamepath.Assetsdir,
			"${assets_index_name}":t.AssetID,
			"${auth_uuid}":t.Account.AccountUUID,
			"${auth_access_token}":t.Account.RefreshToken,
			"${user_type}":t.Account.AccountType,
			"${user_properties}":"{}",
			"${version_type}":t.Profile.LastType(),
		}
		if t.Account.AccountType == "offline" {
			ArgMap["${auth_access_token}"] = "0"
			ArgMap["${user_type}"] = "mojang"
		}
		if t.Profile.GameDirectory != "" && !t.Profile.SeparateInstallation {
			ArgMap["${game_directory}"] = t.Profile.GameDirectory
		}
		for i := range t.GameArgs {
			if i % 2 == 1 {
				t.GameArgs[i] = ArgMap[t.GameArgs[i]]
			}
		}
	}
	if t.Profile.Resolution != nil {
		if !t.Profile.Resolution.Fullscreen {
			t.GameArgs = append(t.GameArgs,"--width",strconv.Itoa(t.Profile.Resolution.Width))
			t.GameArgs = append(t.GameArgs,"--height",strconv.Itoa(t.Profile.Resolution.Height))
		} else {
			t.GameArgs = append(t.GameArgs,"--fullscreen")
		}
	}
}

func (t *LaunchTask) buildJvmArgs() {
	if (t.Profile.JavaDirectory == "" && filepath.Dir(t.JavaPath)[len(filepath.Dir(t.JavaPath))-1] == 'e') {
		switch runtime.GOOS {
		case "windows":
			t.JavaPath = filepath.Join(t.JavaPath,"bin","javaw.exe")
		case "linux":
			t.JavaPath = filepath.Join(t.JavaPath,"bin","java")
		case "darwin":
			t.JavaPath = filepath.Join(t.JavaPath,"jre.bundle","Contents","Home","bin","java")
		}
	}
	if t.Profile.JVMArgs != "" {
		toSlice := regexp.MustCompile(`[^\s]+`)
		t.JavaArgs = append(t.JavaArgs,toSlice.FindAllString(t.Profile.JVMArgs,-1)...)
	} else {
		t.JavaArgs = append(
			t.JavaArgs,
			"-Xdiag",
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:+UseG1GC",
			"-XX:G1NewSizePercent=20",
			"-XX:G1ReservePercent=20",
			"-XX:MaxGCPauseMillis=50",
			"-XX:G1HeapRegionSize=16M",
		)
	}
	if t.LogPath != "" {
		t.JavaArgs = append(
			t.JavaArgs,
			"-Dlog4j2.formatMsgNoLookups=true",
			"-Djava.library.path="+t.NativesPath,
			"-Dlog4j.configurationFile="+t.LogPath,
			"-Dlog4j.rootLogger=OFF",
		)
	} else {
		t.JavaArgs = append(t.JavaArgs,"-Djava.library.path="+t.NativesPath)
	}
	if runtime.GOOS == "darwin" {
		t.JavaArgs = append(t.JavaArgs, "-XstartOnFirstThread")
	}
	if runtime.GOARCH == "386" {
		t.JavaArgs = append(t.JavaArgs, "-Xss1M")
	}
	t.JavaArgs = append(t.JavaArgs, "-cp", t.ClassPath, t.MainClass)
	t.JavaArgs = append(t.JavaArgs,t.GameArgs...)
}

func (t *LaunchTask) buildClassPath() {
	if runtime.GOOS == "windows" {
		t.ClassPath = strings.Join(t.LibsPath,";")
		t.ClassPath = t.ClassPath+";"+filepath.Join(gamepath.Versionsdir,t.Profile.LastVersion(),t.Profile.LastVersion()+".jar")
	} else {
		t.ClassPath = strings.Join(t.LibsPath,":")
		t.ClassPath = t.ClassPath+":"+filepath.Join(gamepath.Versionsdir,t.Profile.LastVersion(),t.Profile.LastVersion()+".jar")
	}
}

// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func InvokeDefault(url string) error {
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
	if err != nil { logutil.Error("Failed to invoke default application",err); return err }
	return err
}
