package resourcemanager

import (

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/parseutil"
	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
)

const runtimesmeta = "https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json"

func Runtimes(versiondata *gjson.Result) error {
	var manifestUrl string
	if versiondata.Get("javaVersion").Exists() {
		requiredComponent := versiondata.Get("javaVersion").Get("component").String()
		logutil.Info("Required jvm runtime for version is "+requiredComponent)
		jsonBytes,err := downloadutil.GetJSON(runtimesmeta)
		if err != nil { logutil.Error(err.Error()); return err }
		if gamepath.UserOS == "linux" {
			if gamepath.UserArch != "i386" {
				manifestUrl = gjson.Get(string(jsonBytes),"linux").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else {
				if gjson.Get(string(jsonBytes),"linux").Get(requiredComponent+".0").Exists() {
					manifestUrl = gjson.Get(string(jsonBytes),"linux").Get(requiredComponent+".0").Get("manifest").Get("url").String()
				} else {
					logutil.Warn("This architecture does not support this jvm runtime"); return nil
				}
			}
		} else if gamepath.UserOS == "darwin" {
			if gamepath.UserArch != "arm64" {
				manifestUrl = gjson.Get(string(jsonBytes),"mac-os").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else {
				if gjson.Get(string(jsonBytes),"mac-os-arm64").Get(requiredComponent+".0").Exists() {
					manifestUrl = gjson.Get(string(jsonBytes),"mac-os-arm64").Get(requiredComponent+".0").Get("manifest").Get("url").String()
				} else {
					logutil.Warn("This architecture does not support this jvm runtime"); return nil
				}
			}
		} else if gamepath.UserOS == "windows" {
			if gamepath.UserArch == "amd64" {
				manifestUrl = gjson.Get(string(jsonBytes),"windows-x64").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else if gamepath.UserArch == "i386" {
				manifestUrl = gjson.Get(string(jsonBytes),"windows-x86").Get(requiredComponent+".0").Get("manifest").Get("url").String()
			} else {
				if gjson.Get(string(jsonBytes),"windows-arm64").Get(requiredComponent+".0").Exists() {
					manifestUrl = gjson.Get(string(jsonBytes),"windows-arm64").Get(requiredComponent+".0").Get("manifest").Get("url").String()
				} else {
					logutil.Warn("This architecture does not support this jvm runtime"); return nil
				}
			}
		}
		runtimeData,err := parseutil.ParseJSON(manifestUrl,true) // continue from here
		if err != nil { logutil.Error(err.Error()); return err }
	} else {
		logutil.Warn("No required jvm runtime found for version. Using jvm installed in system")
	}
	return nil
}
