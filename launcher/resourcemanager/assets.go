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
var objectsDir string = filepath.Join(gamepath.Assetsdir,"objects")+gamepath.P
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

func Assets(assetsdata *gjson.Result,assetid string) { 
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
}

func legacyAssets(pathSlice []string,fileNameSlice []string,assetid string) error {
	var targetDir string
	if assetid == "legacy" {
		targetDir = filepath.Join(gamepath.Assetsdir,"virtual","legacy")
	} else {
		targetDir = filepath.Join(gamepath.Gamedir,"resources")
	}
	for i := range pathSlice {
		file,err := os.Open(pathSlice[i])
		if err != nil { logutil.Error(err.Error()); return err }
		defer file.Close()
		err = gamepath.Makedir(filepath.Dir(filepath.Join(targetDir,fileNameSlice[i])))
		if err != nil { logutil.Error(err.Error()); return err }
		destination,err := os.Create(filepath.Join(targetDir,fileNameSlice[i]))
		if err != nil { logutil.Error(err.Error()); return err }
		defer destination.Close()
		_,err = io.Copy(destination,file)
		if err != nil { logutil.Error(err.Error()); return err }
	}
	return nil
}

func AssetIndex(url string,assetid string) error {
	return downloadutil.DownloadSingle(url,filepath.Join(gamepath.Assetsdir,"indexes",assetid+".json"))
}

func Log4JConfig(versiondata *gjson.Result) string {
	if versiondata.Get("logging").Exists() {
		url := versiondata.Get("logging").Get("client").Get("file").Get("url").String()
		id := versiondata.Get("logging").Get("client").Get("file").Get("id").String()
		downloadutil.DownloadSingle(url,filepath.Join(gamepath.Assetsdir,"log_configs",id))
		return id
	} else {
		return ""
	}
}

