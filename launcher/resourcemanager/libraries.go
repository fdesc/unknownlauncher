package resourcemanager

import (
	"path/filepath"
	"archive/zip"
	"strconv"
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
	pathSlice []string
	urlSlice []string
	hashSlice []string
)

func libraryRules(versiondata *gjson.Result) {  
	logutil.Info("Acquiring OS dependent library rules")
	versiondata.Get("libraries").ForEach(func(_,value gjson.Result) bool {
		if value.Get("rules.1").Exists() {
			if gamepath.UserOS == "darwin" {
				if value.Get("rules.1").Get("os").Get("name").String() == "osx" { 
					if value.Get("rules.1").Get("action").String() == "allow" {
						if value.Get("downloads").Get("artifact").Exists() {
							pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
							hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
							urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
						}	
					} else {
						if gamepath.UserOS == value.Get("rules.1").Get("os").Get("name").String() {
							if value.Get("rules.1").Get("action").String() == "allow" {
								if value.Get("downloads").Get("artifact").Exists() {
									pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
									hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
									urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
								}
							}
						}
					}
				}
			}
			if value.Get("rules.0").Exists() {
				if value.Get("rules.0").Get("action").String() == "allow" {
					pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
					hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
					urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
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
	versiondata.Get("libraries").ForEach(func (_,value gjson.Result) bool{
		if value.Get("downloads").Get("classifiers").Exists() {
			if gamepath.UserOS == "darwin" {
				if value.Get("downloads").Get("classifiers").Get("natives-osx").Exists() {
					pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("classifiers").Get("natives-osx").Get("path").String())
					hashSlice = append(hashSlice, value.Get("downloads").Get("classifiers").Get("natives-osx").Get("sha1").String())
					urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("classifiers").Get("natives-osx").Get("path").String())
				}
			} else if value.Get("downloads").Get("classifiers").Get("natives-"+gamepath.UserOS).Exists() {
				pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("classifiers").Get("natives-"+gamepath.UserOS).Get("path").String())
				hashSlice = append(hashSlice, value.Get("downloads").Get("classifiers").Get("natives-"+gamepath.UserOS).Get("sha1").String())
				urlSlice = append(urlSlice, librariesUrl+"/"+value.Get("downloads").Get("classifiers").Get("natives-"+gamepath.UserOS).Get("path").String())
			}
		}
		return true
	})
	return pathSlice // slice of natives path
}

func unpackNatives(version *string,nativesSlice []string) (string) {
	logutil.Info("Unpacking natives")
	path := filepath.Join(gamepath.Versionsdir,*version,*version+"-natives-"+strconv.Itoa(time.Now().Nanosecond()))
	err := gamepath.Makedir(path); if err != nil { logutil.Error(err.Error()) }
	for i := range nativesSlice {
		zipReader,err := zip.OpenReader(pathSlice[i])
		if err != nil { logutil.Error(err.Error()) }
		defer zipReader.Close()
		for _,f := range zipReader.File {
			openedFiles, err := f.Open(); if err != nil { logutil.Error(err.Error()) }
			defer openedFiles.Close()
			if f.FileInfo().IsDir() {
				os.MkdirAll(filepath.Join(path,f.Name), f.Mode())
			} else {
				f, err := os.OpenFile(filepath.Join(path,f.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil { logutil.Error(err.Error()) }
				defer f.Close()
				if _,err = io.Copy(f,openedFiles); err != nil { logutil.Error(err.Error()) }
			}
		}
	}
	err = os.RemoveAll(filepath.Join(path,"META-INF")); if err != nil { logutil.Error(err.Error()) }
	return path
}

func Libraries(version *string, versiondata *gjson.Result) ([]string,[]string) {
	logutil.Info("Downloading libraries")
	nativesPath := nativeLibraries(versiondata)
	downloadutil.DownloadMultiple(urlSlice,pathSlice)
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
