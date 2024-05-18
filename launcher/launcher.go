package launcher

import (
	"bufio"
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/launcher/profilemanager"
	"fdesc/unknownlauncher/launcher/resourcemanager"
	"fdesc/unknownlauncher/launcher/versionmanager"
	"fdesc/unknownlauncher/util/downloadutil"
	"fdesc/unknownlauncher/util/gamepath"
	"fdesc/unknownlauncher/util/logutil"

	"github.com/tidwall/gjson"
	"golang.org/x/image/draw"
)

var OfflineMode         bool
var ContentMessage      string
var ContentReadMoreLink string
var ContentImage        image.Image

type LauncherSettings struct {
	LauncherTheme     string
	LaunchRule        string
	DisableValidation bool
}

type LaunchTask struct {
	Profile     *profilemanager.ProfileProperties
	Account     *auth.AccountProperties
	javaPath    string
	classPath   string
	mainClass   string
	nativesPath string
	logCfgPath  string
	assetID     string
   libsPath    []string
	javaArgs    []string
	gameArgs    []string
}

func GetLauncherContent() {
	logutil.Info("Acquiring launcher content")
	const launcherContentMeta = "https://launchercontent.mojang.com"
   var imgUrl string
	contentData,err := downloadutil.GetData(launcherContentMeta+"/v2/news.json")
	if err != nil {
		OfflineMode = true
		logutil.Error("Failed to get launcher content",err)
		logutil.Info("Switching to offline mode")
		ContentMessage = "Offline Mode, no internet connection detected"
      ContentImage = nil
		return
	}
	gjson.Get(string(contentData),"entries").ForEach(func(_,value gjson.Result) bool {
		if value.Get("category").String() == "Minecraft: Java Edition" && value.Get("tag").Type == gjson.Null {
         if len(value.Get("newsType").Array()) == 2 && value.Get("newsType.0").String() == "Java" && value.Get("newsType.1").String() == "News page" {
            ContentMessage = value.Get("title").String()
            ContentReadMoreLink = value.Get("readMoreLink").String()
            imgUrl = launcherContentMeta+value.Get("playPageImage").Get("url").String()
         }
      }
      return true
	})
   ContentImage,err = scaleContentImage(imgUrl,400,233)
   if err != nil {
      ContentImage = nil
   }
}

func scaleContentImage(url string,width,height int) (image.Image,error) {
   data,err := downloadutil.GetData(url)
   if err != nil { logutil.Error("Failed to get image data",err); return nil,err }
   img,err := jpeg.Decode(bytes.NewReader(data))
   if err != nil { logutil.Error("Failed to decode data into image",err); return nil,err }
   newImg := image.NewRGBA(image.Rect(0,0,width,height))
   draw.BiLinear.Scale(newImg,newImg.Rect,img,img.Bounds(),draw.Over,nil)
   return newImg,err
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

func cleanDuplicateNatives(version string) {
   if OfflineMode { return; }
	logutil.Info("Cleaning duplicate natives")
   counter := 0
	nativesRegex := regexp.MustCompile(`.*-natives-.*`)
	versionDirs,err := os.ReadDir(filepath.Join(gamepath.Versionsdir,version))
	if err != nil { return }
	for _,file := range versionDirs {
      if file.IsDir() && nativesRegex.MatchString(file.Name()) {
         counter++
         if counter >= 1 {
            os.RemoveAll(filepath.Join(gamepath.Versionsdir,version,file.Name()))
            counter--
         }
      }
	}
}

func (t *LaunchTask) Prepare() error {
	logutil.Info("Started job for version "+t.Profile.LastVersion()+" as user "+t.Account.Name)
   if OfflineMode {
      return t.prepareOffline()
   }
	metaUrl,err := versionmanager.SelectVersion(versionmanager.GetVersionType(t.Profile.LastVersion()),t.Profile.LastVersion())
	if err != nil { return err }
	versionData,err := versionmanager.ParseVersion(metaUrl)
	if err != nil { return err }
	versionId := versionData.Get("id").String()
   err = versionmanager.SaveVersionArguments(&versionData)
	if err != nil { return err }
   argsContent,err := os.ReadFile(filepath.Join(gamepath.Versionsdir,versionId,versionId+".json"))
   if err != nil { logutil.Error("Failed to read contents of arguments data file",err); return err }
   var argsData gjson.Result = gjson.Parse(string(argsContent))
	jvmType := argsData.Get("jvmType").String()
	t.assetID = argsData.Get("assets").String()
	t.mainClass = argsData.Get("mainClass").String()
   t.gameArgs = strings.Split(argsData.Get("arguments").String()," ")
	if t.Profile.SeparateInstallation {
		gamepath.SeparateInstallation = true
		gamepath.Gamedir = t.Profile.GameDirectory
		gamepath.Reload()
	}
	cleanDuplicateNatives(versionId)
	err = resourcemanager.Client(&versionData,versionId)
	if err != nil { return err }
	if t.Profile.JavaDirectory != "" {
		t.javaPath = t.Profile.JavaDirectory
	} else if _,err := os.Stat(filepath.Join(gamepath.Runtimesdir,jvmType)); err != nil {
		var err error
		t.javaPath,err = resourcemanager.Runtimes(&versionData)
		if err != nil { return err }
	} else {
		t.javaPath = filepath.Join(gamepath.Runtimesdir,jvmType)
	}
	err = resourcemanager.AssetIndex(resourcemanager.GetAssetProperties(&versionData))
	if err != nil { return err }
	assetsData,err := resourcemanager.ParseAssets()
	if err != nil { return err }
	resourcemanager.Assets(&assetsData)
	t.logCfgPath = resourcemanager.Log4JConfig(&versionData)
	t.libsPath,t.nativesPath = resourcemanager.Libraries(versionId,&versionData)
   err = versionmanager.AppendLibsToArgumentsData(t.libsPath,t.nativesPath,versionId)
   if err != nil { return err }
	resourcemanager.CleanLibraryList()
	return err
}

func (t *LaunchTask) prepareOffline() error {
	content,err := os.ReadFile(filepath.Join(gamepath.Versionsdir,t.Profile.LastVersion(),t.Profile.LastVersion()+".json"))
	if err != nil { logutil.Error("Failed to read version arguments file",err); return err }
	argsData := gjson.Parse(string(content))
	versionId := argsData.Get("id").String()
	jvmType := argsData.Get("jvmType").String()
	t.gameArgs = strings.Split(argsData.Get("arguments").String()," ")
	t.assetID = argsData.Get("assets").String()
	t.mainClass = argsData.Get("mainClass").String()
	if t.Profile.SeparateInstallation {
		gamepath.SeparateInstallation = true
		gamepath.Gamedir = t.Profile.GameDirectory
		gamepath.Reload()
	}
   cleanDuplicateNatives(versionId)
	if t.Profile.JavaDirectory != "" {
		t.javaPath = t.Profile.JavaDirectory
	} else {
		t.javaPath = filepath.Join(gamepath.Runtimesdir,jvmType)
	}
   argsData.Get("libraries").ForEach(func(_, value gjson.Result) bool {
      if value.Get("path").String() != "" {
         t.libsPath = append(t.libsPath,value.Get("path").String())
      }
      if value.Get("native").Exists() {
         t.nativesPath = value.Get("native").String()
      }
      return true
   })
   resourcemanager.CleanLibraryList()
   return err
}

func (t *LaunchTask) CompleteArguments() (*exec.Cmd,string) {
	t.buildClassPath()
	t.buildGameArguments()
	t.buildJvmArgs()
	os.Chdir(gamepath.Gamedir)
	logutil.Info("Java path is "+t.javaPath) // fix empty condition
	previewArgs := make([]string,len(t.javaArgs))
	copy(previewArgs,t.javaArgs)
	for i := range previewArgs {
		switch previewArgs[i] {
		case "-cp":
         previewArgs[i+1] = " | "+previewArgs[i+1]
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
   gameLogPath := filepath.Join(gamepath.Gamedir,"logs","latest.log")
	logutil.Info("Generated command arguments:"+"\n"+strings.Join(previewArgs,"\n"))
	gamepath.SeparateInstallation = false
	gamepath.Reload()
	return exec.Command(t.javaPath,t.javaArgs...),gameLogPath
}

func (t *LaunchTask) buildGameArguments() {
	if t.gameArgs[0] == "default" {
		t.gameArgs = nil
		t.gameArgs = append(
			t.gameArgs,
			"--username",
			t.Account.Name,
			"--version",
			t.Profile.LastVersion(),
			"--gameDir",
			gamepath.Gamedir,
			"--assetIndex",
			t.assetID,
			"--uuid",
			t.Account.AccountUUID,
			"--accessToken",
			t.Account.RefreshToken,
			"--userType",
			t.Account.AccountType,
		)
		if t.Account.AccountType == "offline" {
			t.gameArgs[11] = "0"
			t.gameArgs[13] = "mojang"
		}
		if t.Profile.GameDirectory != "" && !t.Profile.SeparateInstallation {
			t.gameArgs[5] = t.Profile.GameDirectory
		}
	} else {
		var ArgMap = map[string]string {
			"${auth_player_name}":t.Account.Name,
			"${auth_session}":"0",
			"${version_name}":t.Profile.LastVersion(),
			"${game_directory}":gamepath.Gamedir,
			"${game_assets}":gamepath.Assetsdir,
			"${assets_root}":gamepath.Assetsdir,
			"${assets_index_name}":t.assetID,
			"${auth_uuid}":t.Account.AccountUUID,
			"${auth_access_token}":t.Account.RefreshToken,
			"${user_type}":t.Account.AccountType,
			"${user_properties}":"{}",
		}
		if t.Account.AccountType == "offline" {
			ArgMap["${auth_access_token}"] = "0"
			ArgMap["${user_type}"] = "mojang"
		}
		if t.Profile.GameDirectory != "" && !t.Profile.SeparateInstallation {
			ArgMap["${game_directory}"] = t.Profile.GameDirectory
		}
		for i := range t.gameArgs {
			if t.gameArgs[i][0] == '$' {
				t.gameArgs[i] = ArgMap[t.gameArgs[i]]
			}
		}
	}
	if t.Profile.Resolution != nil {
		if !t.Profile.Resolution.Fullscreen {
			t.gameArgs = append(t.gameArgs,"--width",strconv.Itoa(t.Profile.Resolution.Width))
			t.gameArgs = append(t.gameArgs,"--height",strconv.Itoa(t.Profile.Resolution.Height))
		} else {
			t.gameArgs = append(t.gameArgs,"--fullscreen")
		}
	}
}

func (t *LaunchTask) buildJvmArgs() {
	if (t.Profile.JavaDirectory == "" && filepath.Dir(t.javaPath)[len(filepath.Dir(t.javaPath))-1] == 'e') {
		switch runtime.GOOS {
		case "windows":
			t.javaPath = filepath.Join(t.javaPath,"bin","javaw.exe")
		case "linux":
			t.javaPath = filepath.Join(t.javaPath,"bin","java")
		case "darwin":
			t.javaPath = filepath.Join(t.javaPath,"jre.bundle","Contents","Home","bin","java")
		}
	}
	if t.Profile.JVMArgs != "" {
		toSlice := regexp.MustCompile(`[^\s]+`)
		t.javaArgs = append(t.javaArgs,toSlice.FindAllString(t.Profile.JVMArgs,-1)...)
	} else {
		t.javaArgs = append(
			t.javaArgs,
			"-Xdiag",
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:+UseG1GC",
			"-XX:G1NewSizePercent=20",
			"-XX:G1ReservePercent=20",
			"-XX:MaxGCPauseMillis=50",
			"-XX:G1HeapRegionSize=16M",
		)
	}
	if t.logCfgPath != "" {
		t.javaArgs = append(
			t.javaArgs,
			"-Dlog4j2.formatMsgNoLookups=true",
			"-Djava.library.path="+t.nativesPath,
			"-Dlog4j.configurationFile="+t.logCfgPath,
			"-Dlog4j.rootLogger=OFF",
		)
	} else {
		t.javaArgs = append(t.javaArgs,"-Djava.library.path="+t.nativesPath)
	}
	if runtime.GOOS == "darwin" {
		t.javaArgs = append(t.javaArgs, "-XstartOnFirstThread")
	}
	if runtime.GOARCH == "386" {
		t.javaArgs = append(t.javaArgs, "-Xss1M")
	}
	t.javaArgs = append(t.javaArgs, "-cp", t.classPath, t.mainClass)
	t.javaArgs = append(t.javaArgs,t.gameArgs...)
}

func (t *LaunchTask) buildClassPath() {
	if runtime.GOOS == "windows" {
		t.classPath = strings.Join(t.libsPath,";")
		t.classPath = t.classPath+";"+filepath.Join(gamepath.Versionsdir,t.Profile.LastVersion(),t.Profile.LastVersion()+".jar")
	} else {
		t.classPath = strings.Join(t.libsPath,":")
		t.classPath = t.classPath+":"+filepath.Join(gamepath.Versionsdir,t.Profile.LastVersion(),t.Profile.LastVersion()+".jar")
	}
   logutil.Info(t.libsPath[len(t.libsPath)-1])
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
