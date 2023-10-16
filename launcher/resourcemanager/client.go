package resourcemanager

import (

	"github.com/tidwall/gjson"
//	"egreg10us/faultylauncher/launcher/versionmanager"
	"egreg10us/faultylauncher/util/logutil"
	"egreg10us/faultylauncher/util/gamepath"
	"egreg10us/faultylauncher/util/downloadutil"
)

func Client(versiondata *gjson.Result,ver string) (string,error) {
	logutil.Info("Downloading client JAR for version "+ver)
	jsonResult := versiondata.Get("downloads")
	downloadutil.DownloadSingle(jsonResult.Get("client").Get("url").String(),gamepath.Versionsdir+gamepath.P+ver+gamepath.P+ver+".jar",false)
	logutil.Info("Task client JAR finished for version "+ver)
	return jsonResult.Get("client").Get("sha1").String(),nil
}
