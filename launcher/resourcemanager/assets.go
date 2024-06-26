package resourcemanager

import (
   "path/filepath"
   "os"
   "io"

   "github.com/tidwall/gjson"
   "fdesc/unknownlauncher/util/downloadutil"
   "fdesc/unknownlauncher/util/gamepath"
   "fdesc/unknownlauncher/util/logutil"
)

var assetID string

func GetAssetProperties(versiondata *gjson.Result) string {
   assetID = versiondata.Get("assetIndex.id").String()
   assetsUrl := versiondata.Get("assetIndex.url").String()
   return assetsUrl
}

func ParseAssets() (gjson.Result,error) {
   assets,err := os.ReadFile(filepath.Join(gamepath.Assetsdir,"indexes",assetID+".json"))
   assetsdata := gjson.Parse(string(assets))
   if err != nil { logutil.Error("Failed to parse assets file",err) }
   return assetsdata.Get("objects"),err
}

func Assets(assetsdata *gjson.Result) {
   logutil.Info("Downloading assets")
   var hashPathMap = map[string]string{}
   var hashSlice []string
   var urlSlice []string
   var pathSlice []string
   var fileNameSlice []string
   var objectsDir string = filepath.Join(gamepath.Assetsdir,"objects")
   assetsdata.ForEach(func(key,value gjson.Result) bool {
      if assetID == "legacy" || assetID == "pre-1.6" {
         fileNameSlice = append(fileNameSlice, key.String())
      }
      hashPathMap[value.Get("hash").String()] = filepath.Join(value.Get("hash").String()[:2],value.Get("hash").String())
      return true
   })
   for k,v := range hashPathMap {
      hashSlice = append(hashSlice, k)
      pathSlice = append(pathSlice, filepath.Join(objectsDir,v))
      urlSlice = append(urlSlice, "https://resources.download.minecraft.net/"+v)
   }
   downloadutil.DownloadMultiple(urlSlice,pathSlice)
   for i := range pathSlice {
      if !ValidateChecksum(pathSlice[i],hashSlice[i]) {
         for c := 0; c < 3; c++ {
            os.Remove(pathSlice[i])
            downloadutil.DownloadSingle(urlSlice[i],pathSlice[i])
            if ValidateChecksum(pathSlice[i],hashSlice[i]) {
               break
            }
            if c == 2 {
               logutil.Warn("Failed to validate checksum after 3 tries. Skipping this file")
               break
            }
         }
      } else {
         continue
      }
   }
   if (assetID == "legacy" || assetID == "pre-1.6") {
      legacyAssets(pathSlice,fileNameSlice,assetID)
   }
   logutil.Info("Task downloading assets finished")
   downloadutil.ResetJobCount()
}

func legacyAssets(pathSlice []string,fileNameSlice []string,assetid string) error {
   var targetDir string
   if assetid == "legacy" {
      targetDir = filepath.Join(gamepath.Assetsdir,"virtual","legacy")
      gamepath.Assetsdir = targetDir
   } else {
      targetDir = filepath.Join(gamepath.Gamedir,"resources")
      gamepath.Assetsdir = targetDir
   }
   for i := range pathSlice {
      file,err := os.Open(pathSlice[i])
      if err != nil { logutil.Error("Failed to open file",err); return err }
      defer file.Close()
      err = os.MkdirAll(filepath.Join(targetDir,fileNameSlice[i]),os.ModePerm)
      if err != nil { logutil.Error("Failed to create directory",err); return err }
      destination,err := os.Create(filepath.Join(targetDir,fileNameSlice[i]))
      if err != nil { logutil.Error("Failed to create file",err); return err }
      defer destination.Close()
      _,err = io.Copy(destination,file)
      if err != nil { logutil.Error("Failed to copy data",err); return err }
   }
   return nil
}

func AssetIndex(url string) error {
   return downloadutil.DownloadSingle(url,filepath.Join(gamepath.Assetsdir,"indexes",assetID+".json"))
}

