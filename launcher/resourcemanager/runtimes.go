package resourcemanager

import (
	"path/filepath"
	"strings"
	"strconv"
	"os/exec"
	"bytes"
	"os"
	"io"

	"github.com/ulikunitz/xz/lzma"
	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/parseutil"
	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
)

const runtimesMeta = "https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json"

func Runtimes(versiondata *gjson.Result) (string,error) {
	setIdentifier()
	var requiredComponent string
	var manifestUrl string
	var targetDir string
	if versiondata.Get("javaVersion").Exists() {
		requiredComponent = versiondata.Get("javaVersion").Get("component").String()
		targetDir = filepath.Join(gamepath.Runtimesdir,requiredComponent)
		logutil.Info("Required jvm runtime for version is "+requiredComponent)
		jsonBytes,err := downloadutil.GetData(runtimesMeta)
		if err != nil { logutil.Error("Failed to get jvm runtime data",err); return "",err }
		if identifier == "osx" {
			identifier = "mac-os"
		}
		if (identifier == "windows") {
			manifestUrl = gjson.Get(string(jsonBytes),identifier+"-"+identifierArch).Get(requiredComponent+".0").Get("manifest").Get("url").String()
		}
		if (identifierArch == "i386" || identifierArch == "arm64") {
			if gjson.Get(string(jsonBytes),identifier+"-"+identifierArch).Get(requiredComponent+".0").Exists() {
				manifestUrl = gjson.Get(string(jsonBytes),identifier+"-"+identifierArch).Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else {
				logutil.Warn("This architecture does not support this jvm runtime"); return getDefaultJavaInstallation(),nil
			}
		} else {
			manifestUrl = gjson.Get(string(jsonBytes),identifier).Get(requiredComponent+".0").Get("manifest").Get("url").String()
		}
		runtimeData,err := parseutil.ParseJSON(manifestUrl,true)
		if err != nil { logutil.Error("Failed to parse runtime data",err); return "",err }
		runtimeData.Get("files").ForEach(func(key,value gjson.Result) bool {
			if value.Get("type").String() == "file" {
				if value.Get("downloads").Get("lzma").Exists() {
					_,err = os.Stat(filepath.Join(targetDir,key.String()))
					if err != nil {
						lzmaData,err := downloadutil.GetData(value.Get("downloads").Get("lzma").Get("url").String())
						if err != nil { logutil.Error("Failed to get runtime lzma archive",err) }
						read,err := lzma.NewReader(bytes.NewReader(lzmaData))
						if err != nil { logutil.Error("Failed to read runtime lzma archive",err) }
						err = gamepath.Makedir(filepath.Dir(filepath.Join(targetDir,key.String())))
						if err != nil { logutil.Error("Failed to create directory",err) }
						file,err := os.Create(filepath.Join(targetDir,key.String()))
						if err != nil { logutil.Error("Failed to create file",err) }
						defer file.Close()
						logutil.Info("Downloaded "+filepath.Base(key.String()))
						if _,err = io.Copy(file,read); err != nil { logutil.Error("Failed to copy lzma data",err) }
					}
				} else {
					downloadutil.DownloadSingle(value.Get("downloads").Get("raw").Get("url").String(),filepath.Join(targetDir,key.String()))
				}
			}
			if (value.Get("executable").Exists() && (gamepath.UserOS == "linux" || gamepath.UserOS == "darwin")) {
				if value.Get("executable").Bool() == true {
					err := os.Chmod(filepath.Join(targetDir,key.String()),0755)
					if err != nil { logutil.Error("Failed to chmod executable file",err) }
				}
			}
			return true
		})
	} else {
		logutil.Warn("No required jvm runtime found for version. Using jvm installed in system")
		return getDefaultJavaInstallation(),nil
	}
	return targetDir,nil
}

func getDefaultJavaInstallation() string {
	logutil.Info("Trying to obtain java installation path. This might take a while depending on your OS")
	if gamepath.UserOS == "windows" {
		// https://stackoverflow.com/questions/69990781/how-do-i-find-where-java-is-installed-on-windows-10
		// Not every Java installer will automatically set JAVA_HOME
		out,_ := exec.Command("powershell","wmic","product","where",strconv.Quote(`Name like '%%Java%%'`),"get","installlocation").Output()
		return strings.TrimSpace(strings.ReplaceAll(string(out),"InstallLocation",""))+"bin"+gamepath.P+"java.exe"
	} else {
		out,_ := exec.Command("which","java").Output()
		return string(out)
	}
}
