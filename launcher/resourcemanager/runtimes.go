package resourcemanager

import (

	"github.com/ulikunitz/xz/lzma"
	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/parseutil"
	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
	"bytes"
	"os"
	"io"
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
		runtimeData,err := parseutil.ParseJSON(manifestUrl,true)
		if err != nil { logutil.Error(err.Error()); return err }
		runtimeData.Get("files").ForEach(func(key,value gjson.Result) bool { // ...
			if value.Get("type").Exists() {
				if value.Get("type").String() == "directory" {
					gamepath.Makedir(gamepath.Runtimesdir+gamepath.P+requiredComponent+gamepath.P+key.String())
				}
			}
			return true 
		})
		runtimeData.Get("files").ForEach(func(key,value gjson.Result) bool {
			if value.Get("type").Exists() {
				if value.Get("type").String() == "file" {
					if value.Get("downloads").Get("lzma").Exists() {
						        _,err := os.Stat(gamepath.Runtimesdir+gamepath.P+requiredComponent+gamepath.P+key.String())
							if err != nil {
								lzmaData,err := downloadutil.GetJSON(value.Get("downloads").Get("lzma").Get("url").String())
								if err != nil { logutil.Error(err.Error()) }
								read,err := lzma.NewReader(bytes.NewReader(lzmaData))
								if err != nil { logutil.Error(err.Error()) }
								file,err := os.Create(gamepath.Runtimesdir+gamepath.P+requiredComponent+gamepath.P+key.String())
								if err != nil { logutil.Error(err.Error()) }
								logutil.Info("Downloaded "+key.String())
								defer file.Close()
								if _,err := io.Copy(file,read); err != nil { logutil.Error(err.Error()) }
							}
					} else {
						downloadutil.DownloadSingle(value.Get("downloads").Get("raw").Get("url").String(),gamepath.Runtimesdir+gamepath.P+requiredComponent+gamepath.P+key.String(),false)
					}
				}
			}
			if value.Get("executable").Exists() && (gamepath.UserOS == "linux" || gamepath.UserOS == "darwin") {
				if value.Get("executable").Bool() == true {
					err := os.Chmod(gamepath.Runtimesdir+gamepath.P+requiredComponent+gamepath.P+key.String(), 0755)
					if err != nil { logutil.Error(err.Error()) }
				}
			}
			return true
		})
	} else {
		logutil.Warn("No required jvm runtime found for version. Using jvm installed in system")
	}
	return nil
}
