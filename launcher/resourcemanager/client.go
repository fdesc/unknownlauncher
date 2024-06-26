package resourcemanager

import (
   "path/filepath"
   "os"

   "github.com/tidwall/gjson"
   "fdesc/unknownlauncher/util/logutil"
   "fdesc/unknownlauncher/util/gamepath"
   "fdesc/unknownlauncher/util/downloadutil"
)

func Client(versiondata *gjson.Result,version string) error {
   logutil.Info("Downloading client JAR for version "+version)
   jsonResult := versiondata.Get("downloads")
   err := downloadutil.DownloadSingle(jsonResult.Get("client.url").String(),filepath.Join(gamepath.Versionsdir,version,version+".jar"))
   if err != nil { return err }
   if !ValidateChecksum(filepath.Join(gamepath.Versionsdir,version,version+".jar"),jsonResult.Get("client").Get("sha1").String()) {
      os.Remove(filepath.Join(gamepath.Versionsdir,version,version+".jar"))
      err = downloadutil.DownloadSingle(jsonResult.Get("client.url").String(),filepath.Join(gamepath.Versionsdir,version,version+".jar"))
      if err != nil { return err }
   }
   if jsonResult.Get("client_mappings").Exists() {
      err = downloadutil.DownloadSingle(jsonResult.Get("client_mappings.url").String(),filepath.Join(gamepath.Versionsdir,version,"client.txt"))
      if err != nil { return err }
      if !ValidateChecksum(filepath.Join(gamepath.Versionsdir,version,"client.txt"),jsonResult.Get("client_mappings").Get("sha1").String()) {
         os.Remove(filepath.Join(gamepath.Versionsdir,version,"client.txt"))
         err = downloadutil.DownloadSingle(jsonResult.Get("client_mappings.url").String(),filepath.Join(gamepath.Versionsdir,version,"client.txt"))
         if err != nil { return err }
      }
   }
   logutil.Info("Task client JAR finished for version "+version)
   downloadutil.ResetJobCount()
   return err
}

func Log4JConfig(versiondata *gjson.Result) string {
   if versiondata.Get("logging").Exists() {
      url := versiondata.Get("logging.client.file.url").String()
      id := versiondata.Get("logging.client.file.id").String()
      downloadutil.DownloadSingle(url,filepath.Join(gamepath.Assetsdir,"log_configs",id))
      downloadutil.ResetJobCount()
      return filepath.Join(gamepath.Assetsdir,"log_configs",id)
   } else {
      return ""
   }
}
