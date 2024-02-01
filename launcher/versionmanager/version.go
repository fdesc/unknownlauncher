package versionmanager

import (
	"path/filepath"
	"os"

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
	"egreg10us/faultylauncher/util/parseutil"
	"egreg10us/faultylauncher/util/downloadutil"
)

const versionMeta string = `https://launchermeta.mojang.com/mc/game/version_manifest_v2.json`
var VersionList = make(map[string][]string)
var LatestRelease string
var LatestSnapshot string

func SelectVersion(versionType,version string) (string,error) {
	var versionUrl string
	jsonBytes,err := downloadutil.GetData(versionMeta); if err != nil { logutil.Error("Failed to get data for version",err); return "",err }
	gjson.Get(string(jsonBytes),"versions").ForEach(func(key, value gjson.Result) bool {
		if (versionType == value.Get("type").String()) {
			if (version == value.Get("id").String()) {
				versionUrl = value.Get("url").String()
				return true
			}
		}
		return true
	})
	return versionUrl,nil
}

func GetVersionArguments(versiondata *gjson.Result) error {
	var data []byte
	os.MkdirAll(filepath.Join(gamepath.Assetsdir,"args"),os.ModePerm)
	os.Create(filepath.Join(gamepath.Assetsdir,"args",versiondata.Get("id").String()+".json"))
	if versiondata.Get("minecraftArguments").Exists() {
		data = []byte(`{"assets":`+versiondata.Get("assets").Raw+`,"id":`+versiondata.Get("id").Raw+`,"mainclass":`+versiondata.Get("mainClass").Raw+`,"libraries":`+versiondata.Get("libraries").Raw+`,"arguments":`+versiondata.Get("minecraftArguments").Raw+`}`)
	} else {
		data = []byte(`{"assets":`+versiondata.Get("assets").Raw+`,"id":`+versiondata.Get("id").Raw+`,"mainclass":`+versiondata.Get("mainClass").Raw+`,"libraries":`+versiondata.Get("libraries").Raw+`,"arguments":"default"}`)
	}
	err := os.WriteFile(filepath.Join(gamepath.Assetsdir,"args",versiondata.Get("id").String()+".json"),data,os.ModePerm)
	if err != nil { logutil.Error("Failed to save version data",err); return err }
	return err
}

func GetVersionList() error {
	logutil.Info("Acquiring version list")
	jsonBytes,err := downloadutil.GetData(versionMeta); if err != nil {
		searchLocalVersions()
		logutil.Error("Failed to get data for version",err)
		return err
	}
	LatestRelease = gjson.Get(string(jsonBytes),"latest").Get("release").String()
	LatestSnapshot = gjson.Get(string(jsonBytes),"latest").Get("snapshot").String()
	gjson.Get(string(jsonBytes),"versions").ForEach(func(_,value gjson.Result) bool {
		if _,ok := VersionList[value.Get("type").String()]; !ok {
			VersionList[value.Get("type").String()] = []string{}
		}
		for k,v := range VersionList {
			if k == value.Get("type").String() {
				if value.Get("id").Exists() {
					v = append(v,value.Get("id").String())
					VersionList[k] = v
				}
			}
		}
		return true
	})
	return err
}

func searchLocalVersions() error {
	var names []string
	dirEntry,err := os.ReadDir(filepath.Join(gamepath.Assetsdir,"args"))
	if err != nil { 
		logutil.Error("Failed to read directory contents",err)
		return err
	}
	for _,file := range dirEntry {
		if !file.IsDir() {
			filename := file.Name()
			names = append(names,filename[:len(filename)-5])
			VersionList["Local"] = names
		}
	}
	return err
}

func ParseVersion(url string) (gjson.Result,error) {
	return parseutil.ParseJSON(url,true)
}
