package resourcemanager

import (
	"path/filepath"
	"archive/zip"
	"strconv"
	"strings"
	"time"
	"os"
	"io"

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
)

const librariesUrl = "https://libraries.minecraft.net"
var (
	identifier string
	identifierArch string
	identifierArchOld string
	pathSlice []string
	urlSlice []string
	hashSlice []string
)

func getIdentifier() {
	if gamepath.UserOS == "darwin" {
		identifier = "osx"
		if gamepath.UserArch == "arm64" {
			identifierArch = "arm64"
		}
	} else if gamepath.UserOS == "windows" {
		identifier = "windows"
		if gamepath.UserArch == "amd64" {
			identifierArch = "x64"
			identifierArchOld = "64"
		} else if gamepath.UserArch == "386" {
			identifierArch = "x86"
			identifierArchOld = "32"
		} else if gamepath.UserArch == "arm64" {
			identifierArch = "arm64"
		}
	} else if gamepath.UserOS == "linux" {
		identifier = "linux"
		if gamepath.UserArch == "arm64" {
			identifierArch = "aarch_64"
		} else if gamepath.UserArch == "amd64" {
			identifierArch = "x86_64"
		}
	}
}

func libraryRules(versiondata *gjson.Result) {  
	logutil.Info("Acquiring OS dependent library rules")
	versiondata.Get("libraries").ForEach(func(_,value gjson.Result) bool {
		if value.Get("rules.1").Exists() {
			if value.Get("rules.1").Get("os").Get("name").String() == identifier {
					if value.Get("rules.1").Get("action").String() == "allow" {
						if value.Get("downloads").Get("artifact").Exists() {
							pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
							hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
							urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
						}
					}
				}
			}
		if value.Get("rules.0").Exists() {
			if value.Get("rules.0").Get("action").String() == "allow" {
				pkgName := strings.Split(value.Get("name").String(),":")
				if (pkgName[len(pkgName)-1] != "natives-"+identifier && pkgName[len(pkgName)-1] != "natives-"+identifier+"-"+identifierArch && pkgName[len(pkgName)-1] != identifier+"-"+identifierArch) == false {
					if value.Get("downloads").Get("artifact").Exists() {
						pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
						hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
						urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
					}
				}
			}
		}
		return true
	})
}

func defaultLibraries(versiondata *gjson.Result) []string {
	logutil.Info("Acquiring default libraries")
	versiondata.Get("libraries").ForEach(func (_,value gjson.Result) bool{
		if value.Get("rules").Exists() {
			// literally do nothing(cant use continue)
		} else if value.Get("downloads").Get("artifact").Exists() {	
			pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
			hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
			urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
		}
		return true
	})
	return pathSlice
}

func nativeLibraries(versiondata *gjson.Result) []string {
	logutil.Info("Acquiring native libraries")
	versiondata.Get("libraries").ForEach(func (_,value gjson.Result) bool {
		if value.Get("downloads").Get("classifiers").Exists() {
			if value.Get("downloads").Get("classifiers").Get("natives-"+identifier).Exists() {
				if identifierArchOld != "" {
					if value.Get("downloads").Get("classifiers").Get("natives-"+identifier+"-"+identifierArchOld).Exists() {
						pathSlice = append(pathSlice, filepath.Join(gamepath.Librariesdir,value.Get("downloads").Get("classifiers").Get("natives-"+identifier+"-"+identifierArchOld).Get("path").String()))
						hashSlice = append(hashSlice, value.Get("downloads").Get("classifiers").Get("natives-"+identifier+"-"+identifierArchOld).Get("sha1").String())
						urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("classifiers").Get("natives-"+identifier+"-"+identifierArchOld).Get("path").String())
					}
				}
				pathSlice = append(pathSlice, filepath.Join(gamepath.Librariesdir,value.Get("downloads").Get("classifiers").Get("natives-"+identifier).Get("path").String()))
				hashSlice = append(hashSlice, value.Get("downloads").Get("classifiers").Get("natives-"+identifier).Get("sha1").String())
				urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("classifiers").Get("natives-"+identifier).Get("path").String())
			}
		}
		if value.Get("rules.0").Exists() {
			if (value.Get("rules.0").Get("action").String() == "allow" && value.Get("rules.0").Get("os").Get("name").String() == identifier) {
				pkgName := strings.Split(value.Get("name").String(),":")
				if (pkgName[len(pkgName)-1] == "natives-"+identifier || pkgName[len(pkgName)-1] == "natives-"+identifier+"-"+identifierArch || pkgName[len(pkgName)-1] == identifier+"-"+identifierArch) {
					pathSlice = append(pathSlice, filepath.Join(gamepath.Librariesdir,value.Get("downloads").Get("artifact").Get("path").String()))
					hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
					urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
				}
			}
		}
		return true
	})
	return pathSlice // slice of natives path
}

func unpackNatives(version string,nativesSlice []string) (string) {
	logutil.Info("Unpacking natives")
	target := filepath.Join(gamepath.Versionsdir,version,version+"-natives-"+strconv.Itoa(time.Now().Nanosecond()))
	for i := range nativesSlice {
		jarFile,err := zip.OpenReader(pathSlice[i])
		if err != nil { logutil.Error("O"+err.Error()) }
		defer jarFile.Close()
		for _,file := range jarFile.File {
			openedFiles, err := file.Open(); if err != nil { logutil.Error("O2"+err.Error()) }
			defer openedFiles.Close()
			os.MkdirAll(filepath.Dir(filepath.Join(target,file.Name)),os.ModePerm)
			if file.FileInfo().IsDir() {
				continue
			} else {
				f, err := os.Create(filepath.Join(target,filepath.Base(file.Name)))
				if err != nil { logutil.Error("O3"+err.Error()) }
				defer f.Close()
				if _,err = io.Copy(f,openedFiles); err != nil { logutil.Error("O4"+err.Error()) }
			}
		}
	}
	fileList,err := os.ReadDir(target)
	if err != nil { logutil.Error(err.Error()) }
	for _,file := range fileList {
		currentFile := filepath.Join(target,file.Name())
		if file.IsDir() {
			os.RemoveAll(currentFile)
		}
		if (filepath.Ext(currentFile) != ".so" && filepath.Ext(currentFile) != ".dll" && filepath.Ext(currentFile) != ".dylib") {
			os.RemoveAll(currentFile)
		}
	}
	return target
}

func Libraries(version string, versiondata *gjson.Result) ([]string,[]string) {
	getIdentifier()
	logutil.Info("Downloading libraries")
	nativesPath := nativeLibraries(versiondata)
	defaultLibraries(versiondata)
	libraryRules(versiondata)
	downloadutil.DownloadMultiple(urlSlice,pathSlice)
	for i := range pathSlice {
		if ValidateChecksum(pathSlice[i],hashSlice[i]) == false {
			os.Remove(pathSlice[i])
			downloadutil.DownloadSingle(urlSlice[i],pathSlice[i])
		} else {
			continue
		}
	}
	unpackNatives(version,nativesPath)
	logutil.Info("Finished downloading libraries")
	return pathSlice,nativesPath
}
