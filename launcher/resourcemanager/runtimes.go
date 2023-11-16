package resourcemanager

import (
	"path/filepath"
	"os/exec"
	"regexp"
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

const runtimesmeta = "https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json"

func Runtimes(versiondata *gjson.Result) (error,string) {
	var requiredComponent string
	var manifestUrl string
	var targetDir string
	if versiondata.Get("javaVersion").Exists() {
		requiredComponent = versiondata.Get("javaVersion").Get("component").String()
		targetDir = filepath.Join(gamepath.Runtimesdir,requiredComponent)
		logutil.Info("Required jvm runtime for version is "+requiredComponent)
		jsonBytes,err := downloadutil.GetData(runtimesmeta)
		if err != nil { logutil.Error(err.Error()); return err,"" }
		if gamepath.UserOS == "linux" {
			if gamepath.UserArch != "386" {
				manifestUrl = gjson.Get(string(jsonBytes),"linux").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else {
				if gjson.Get(string(jsonBytes),"linux-i386").Get(requiredComponent+".0").Exists() {
					manifestUrl = gjson.Get(string(jsonBytes),"linux-i386").Get(requiredComponent+".0").Get("manifest").Get("url").String()
				} else {
					logutil.Warn("This architecture does not support this jvm runtime"); return nil,getDefaultJavaInstallation()
				}
			}
		} else if gamepath.UserOS == "darwin" {
			if gamepath.UserArch != "arm64" {
				manifestUrl = gjson.Get(string(jsonBytes),"mac-os").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else {
				if gjson.Get(string(jsonBytes),"mac-os-arm64").Get(requiredComponent+".0").Exists() {
					manifestUrl = gjson.Get(string(jsonBytes),"mac-os-arm64").Get(requiredComponent+".0").Get("manifest").Get("url").String()
				} else {
					logutil.Warn("This architecture does not support this jvm runtime"); return nil,getDefaultJavaInstallation()
				}
			}
		} else if gamepath.UserOS == "windows" {
			if gamepath.UserArch == "amd64" {
				manifestUrl = gjson.Get(string(jsonBytes),"windows-x64").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else if gamepath.UserArch == "386" {
				manifestUrl = gjson.Get(string(jsonBytes),"windows-x86").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else {
				if gjson.Get(string(jsonBytes),"windows-arm64").Get(requiredComponent+".0").Exists() {
					manifestUrl = gjson.Get(string(jsonBytes),"windows-arm64").Get(requiredComponent+".0").Get("manifest").Get("url").String()
				} else {
					logutil.Warn("This architecture does not support this jvm runtime"); return nil,getDefaultJavaInstallation()
				}
			}
		}
		runtimeData,err := parseutil.ParseJSON(manifestUrl,true)
		if err != nil { logutil.Error(err.Error()); return err,"" }
		runtimeData.Get("files").ForEach(func(key,value gjson.Result) bool {
			if value.Get("type").Exists() {
				if value.Get("type").String() == "file" {
					if value.Get("downloads").Get("lzma").Exists() {
						        _,err = os.Stat(filepath.Join(targetDir,key.String()))
							if err != nil {
								lzmaData,err := downloadutil.GetData(value.Get("downloads").Get("lzma").Get("url").String())
								if err != nil { logutil.Error(err.Error()) }
								read,err := lzma.NewReader(bytes.NewReader(lzmaData))
								if err != nil { logutil.Error(err.Error()) }
								err = gamepath.Makedir(filepath.Dir(filepath.Join(targetDir,key.String())))
								if err != nil { logutil.Error(err.Error()) }
								file,err := os.Create(filepath.Join(targetDir,key.String()))
								if err != nil { logutil.Error(err.Error()) }
								defer file.Close()
								logutil.Info("Downloaded "+filepath.Base(key.String()))
								if _,err = io.Copy(file,read); err != nil { logutil.Error(err.Error()) }
							}
					} else {
						downloadutil.DownloadSingle(value.Get("downloads").Get("raw").Get("url").String(),filepath.Join(targetDir,key.String()))
					}
				}
			}
			if value.Get("executable").Exists() && (gamepath.UserOS == "linux" || gamepath.UserOS == "darwin") {
				if value.Get("executable").Bool() == true {
					err := os.Chmod(filepath.Join(targetDir,key.String()),0755)
					if err != nil { logutil.Error(err.Error()) }
				}
			}
			return true
		})
	} else {
		logutil.Warn("No required jvm runtime found for version. Using jvm installed in system")
		return nil,getDefaultJavaInstallation()
	}
	return nil,targetDir
}

func getDefaultJavaInstallation() string {
	if gamepath.UserOS == "windows" {
		// https://stackoverflow.com/questions/69990781/how-do-i-find-where-java-is-installed-on-windows-10
		// Not every Java installer will automatically set JAVA_HOME
		findPath := regexp.MustCompile(`^C:\\.*`)
		out,_ := exec.Command(`wmic product where "Name like '%%Java%%'" get installlocation`).Output()
		return string(findPath.Find(out))+"bin"+gamepath.P+"java.exe"
	} else {
		out,_ := exec.Command("which java").Output()
		return string(out)
	}
}
