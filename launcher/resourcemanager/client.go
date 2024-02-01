package resourcemanager

import (
	"path/filepath"
	"os"

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/logutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/downloadutil"
)

func Client(versiondata *gjson.Result,version string) error {
	logutil.Info("Downloading client JAR for version "+version)
	jsonResult := versiondata.Get("downloads")
	err := downloadutil.DownloadSingle(jsonResult.Get("client").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,version+".jar"))
	if err != nil { return err }
	if !ValidateChecksum(filepath.Join(gamepath.Versionsdir,version,version+".jar"),jsonResult.Get("client").Get("sha1").String()) {
		os.Remove(filepath.Join(gamepath.Versionsdir,version,version+".jar"))
		err = downloadutil.DownloadSingle(jsonResult.Get("client").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,version+".jar"))
		if err != nil { return err }
	}
	if jsonResult.Get("client_mappings").Exists() {
		err = downloadutil.DownloadSingle(jsonResult.Get("client_mappings").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,"client.txt"))
		if err != nil { return err }
		if !ValidateChecksum(filepath.Join(gamepath.Versionsdir,version,"client.txt"),jsonResult.Get("client_mappings").Get("sha1").String()) {
			os.Remove(filepath.Join(gamepath.Versionsdir,version,"client.txt"))
			err = downloadutil.DownloadSingle(jsonResult.Get("client_mappings").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,"client.txt"))
			if err != nil { return err }
		}
	}
	logutil.Info("Task client JAR finished for version "+version)
	return err
}

func Log4JConfig(versiondata *gjson.Result) string {
	if versiondata.Get("logging").Exists() {
		url := versiondata.Get("logging").Get("client").Get("file").Get("url").String()
		id := versiondata.Get("logging").Get("client").Get("file").Get("id").String()
		downloadutil.DownloadSingle(url,filepath.Join(gamepath.Assetsdir,"log_configs",id))
		return filepath.Join(gamepath.Assetsdir,"log_configs",id)
	} else {
		return ""
	}
}
