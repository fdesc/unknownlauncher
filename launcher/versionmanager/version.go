package versionmanager

import (

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/logutil"
	"egreg10us/faultylauncher/util/parseutil"
	"egreg10us/faultylauncher/util/downloadutil"
)

var metasrc string = `https://launchermeta.mojang.com/mc/game/version_manifest_v2.json`

type GameVersion struct {
	Version string
	VersionType string
	VersionUrl string
}

func SelectVersion(v *GameVersion) (string,string,error) {
	v.VersionUrl = ""
	jsonBytes,err := downloadutil.GetData(metasrc); if err != nil { logutil.Error(err.Error()); return "","",err }
	jsonResult := gjson.Get(string(jsonBytes),"versions")
	jsonResult.ForEach(func(key, value gjson.Result) bool {
		if (v.VersionType == value.Get("type").String()) {
			if (v.Version == value.Get("id").String()) {
				v.VersionUrl = value.Get("url").String()
				return true
			}
		}
		return true
	})
	return v.VersionUrl,v.Version,nil
}

func ParseVersion(url string) (gjson.Result,error) {
	return parseutil.ParseJSON(url,true)
}
