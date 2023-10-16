package resourcemanager

import (
	"path/filepath"
	"os"
	"io"

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/downloadutil"
	"egreg10us/faultylauncher/util/parseutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/logutil"
)

const resourcesUrl = "https://resources.download.minecraft.net/"
var objectsDir string = gamepath.Assetsdir+gamepath.P+"objects"+gamepath.P
var AssetID string

func GetAssetProperties(versiondata *gjson.Result) (string,string) {
	AssetID = versiondata.Get("assetIndex").Get("id").String()
	assetsUrl := versiondata.Get("assetIndex").Get("url").String()
	return assetsUrl,AssetID
}

func ParseAssets(url string) (gjson.Result,error) {
	assetdata,err := parseutil.ParseJSON(url,true)
	return assetdata.Get("objects"),err
}

func Assets(assetsdata *gjson.Result,assetid string) error { // add asset downloading via assetid (legacy,pre-1.6) as separate function
	var hashSlice []string
	var urlSlice []string
	var pathSlice []string
	var fileNameSlice []string
	assetsdata.ForEach(func(key,value gjson.Result) bool {
		if assetid == "legacy" || assetid == "pre-1.6" {
			fileNameSlice = append(fileNameSlice, key.String())
		}
		hashSlice = append(hashSlice,value.Get("hash").String())
		return true
	})
	for i := range hashSlice {
		urlSlice = append(urlSlice, resourcesUrl+hashSlice[i][:2]+"/"+hashSlice[i])
		pathSlice = append(pathSlice, objectsDir+hashSlice[i][:2]+gamepath.P+hashSlice[i])
	}
	logutil.Info("Downloading assets")
	downloadutil.DownloadMultiple(urlSlice,pathSlice)
	if assetid == "legacy" || assetid == "pre-1.6" {
		legacyAssets(pathSlice,fileNameSlice,assetid)
	}
	logutil.Info("Task downloading assets finished")
	return nil
}

func legacyAssets(pathSlice []string,fileNameSlice []string,assetid string) error {
	var targetDir string
	if assetid == "legacy" {
		targetDir = gamepath.Assetsdir+gamepath.P+"virtual"+gamepath.P+"legacy"
	} else {
		targetDir = gamepath.Gamedir+gamepath.P+"resources"
	}
	for i := range pathSlice {
		file,err := os.Open(pathSlice[i])
		if err != nil { logutil.Error(err.Error()); return err }
		defer file.Close()
		err = gamepath.Makedir(filepath.Dir(targetDir+gamepath.P+fileNameSlice[i]))
		if err != nil { logutil.Error(err.Error()); return err }
		destination,err := os.Create(targetDir+gamepath.P+fileNameSlice[i])
		if err != nil { logutil.Error(err.Error()); return err }
		defer destination.Close()
		_,err = io.Copy(destination,file)
		if err != nil { logutil.Error(err.Error()); return err }
	}
	return nil
}

func AssetIndex(url string,assetid string) error {
	return downloadutil.DownloadSingle(url,gamepath.Assetsdir+gamepath.P+"indexes"+gamepath.P+assetid+".json",false)
}

func Log4JConfig(versiondata *gjson.Result) string {
	if versiondata.Get("logging").Exists() {
		url := versiondata.Get("logging").Get("client").Get("file").Get("url").String()
		id := versiondata.Get("logging").Get("client").Get("file").Get("id").String()
		downloadutil.DownloadSingle(url,gamepath.Assetsdir+gamepath.P+"log_configs"+gamepath.P+id,false)
		return id
	} else {
		return ""
	}
}

