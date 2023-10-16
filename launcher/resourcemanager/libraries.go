package resourcemanager

import (

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
)

const librariesUrl = "https://libraries.minecraft.net"

func LibraryRules(versiondata *gjson.Result) error {  
	var osDependentPathSlice []string
	var osDependentUrlSlice []string
	var hashSlice []string
	logutil.Info("Acquiring OS dependent library rules")
	versiondata.Get("libraries").ForEach(func(_,value gjson.Result) bool {
		if value.Get("rules.1").Exists() {
			if gamepath.UserOS == "darwin" {
				if value.Get("rules.1").Get("os").Get("name").String() == "osx" { 
					if value.Get("rules.1").Get("action").String() == "allow" {
						if value.Get("downloads").Get("artifact").Exists() {
							osDependentPathSlice = append(osDependentPathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
							hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
							osDependentUrlSlice = append(osDependentUrlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
						}	
					} else {
						if gamepath.UserOS == value.Get("rules.1").Get("os").Get("name").String() {
							if value.Get("rules.1").Get("action").String() == "allow" {
								if value.Get("downloads").Get("artifact").Exists() {
									osDependentPathSlice = append(osDependentPathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
									hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
									osDependentUrlSlice = append(osDependentUrlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
								}
							}
						}
					}
				}
			}
			if value.Get("rules.0").Exists() {
				if value.Get("rules.0").Get("action").String() == "allow" {
					osDependentPathSlice = append(osDependentPathSlice, gamepath.Librariesdir+gamepath.P+value.Get("downloads").Get("artifact").Get("path").String())
					hashSlice = append(hashSlice, value.Get("downloads").Get("artifact").Get("sha1").String())
					osDependentUrlSlice = append(osDependentUrlSlice, librariesUrl+"/"+value.Get("downloads").Get("artifact").Get("path").String())
				}
			}
		}
		return true
	})
	if osDependentPathSlice != nil {
		downloadutil.DownloadMultiple(osDependentUrlSlice,osDependentPathSlice)
		return nil
	} else {
		return nil
	}
}

func DefaultLibraries(versiondata *gjson.Result) {
	var pathSlice []string
	var urlSlice []string
	var hashSlice []string
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
	downloadutil.DownloadMultiple(urlSlice,pathSlice)
}

func NativeLibraries(versiondata *gjson.Result) {
	var pathSlice []string
	var urlSlice []string
	var hashSlice []string
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
				// unpackNatives(value.Get()...)
			}
		}
		return true
	})
	for i := range pathSlice { // TODO: fix this
		downloadutil.DownloadSingle(urlSlice[i],pathSlice[i],false)
	}
}

func Libraries(versiondata *gjson.Result) {
	DefaultLibraries(versiondata)
	LibraryRules(versiondata)
	NativeLibraries(versiondata)
}
