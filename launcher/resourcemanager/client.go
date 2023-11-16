package resourcemanager

import (
	"path/filepath"
	"os"

	"github.com/tidwall/gjson"
	"egreg10us/faultylauncher/util/logutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/downloadutil"
)

func Client(versiondata *gjson.Result,version string) {
	logutil.Info("Downloading client JAR for version "+version)
	jsonResult := versiondata.Get("downloads")
	downloadutil.DownloadSingle(jsonResult.Get("client").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,version+".jar"))
	if ValidateChecksum(filepath.Join(gamepath.Versionsdir,version,version+".jar"),jsonResult.Get("client").Get("sha1").String()) == false {
		os.Remove(filepath.Join(gamepath.Versionsdir,version,version+".jar"))
		downloadutil.DownloadSingle(jsonResult.Get("client").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,version+".jar"))
	}
	if jsonResult.Get("client_mappings").Exists() {
		downloadutil.DownloadSingle(jsonResult.Get("client_mappings").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,"client.txt"))
		if ValidateChecksum(filepath.Join(gamepath.Versionsdir,version,"client.txt"),jsonResult.Get("client_mappings").Get("sha1").String()) == false {
			os.Remove(filepath.Join(gamepath.Versionsdir,version,"client.txt"))
			downloadutil.DownloadSingle(jsonResult.Get("client_mappings").Get("url").String(),filepath.Join(gamepath.Versionsdir,version,"client.txt"))
		}
	}
	logutil.Info("Task client JAR finished for version "+version)
}
