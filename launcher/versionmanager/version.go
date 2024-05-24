package versionmanager

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"

	"fdesc/unknownlauncher/util/downloadutil"
	"fdesc/unknownlauncher/util/gamepath"
	"fdesc/unknownlauncher/util/logutil"
)

const versionMeta string = `https://launchermeta.mojang.com/mc/game/version_manifest_v2.json`
var VersionList = make(map[string][]string)
var LatestRelease string
var LatestSnapshot string

type classPathList struct {
   Path string `json:"path,omitempty"`
   Native string `json:"native,omitempty"`
}

type argumentLookup struct {
   JvmType   string `json:"jvmType"`
   Assets    string `json:"assets"`
   Id        string `json:"id"`
   MainClass string `json:"mainClass"`
   Arguments string `json:"arguments"`
   Libraries []classPathList `json:"libraries,omitempty"`
}

func SelectVersion(versionType,version string) (string,error) {
	var versionUrl string
	jsonBytes,err := downloadutil.GetData(versionMeta); if err != nil { logutil.Error("Failed to get data for version",err); return "",err }
	gjson.Get(string(jsonBytes),"versions").ForEach(func(_, value gjson.Result) bool {
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

func SaveVersionArguments(versiondata *gjson.Result) error {
   version := versiondata.Get("id").String()
   os.MkdirAll(filepath.Join(gamepath.Versionsdir,version),os.ModePerm)
	file,err := os.Create(filepath.Join(gamepath.Versionsdir,version,version+".json"))
	if err != nil { logutil.Error("Failed to create arguments to save files",err); return err }
   args := &argumentLookup{
      JvmType: versiondata.Get("javaVersion.component").String(),
      Assets: versiondata.Get("assets").String(),
      Id: version,
      MainClass: versiondata.Get("mainClass").String(),
   }
   if versiondata.Get("minecraftArguments").Exists() {
      args.Arguments = versiondata.Get("minecraftArguments").String()
   } else {
      args.Arguments = "default"
   }
   data,err := json.Marshal(args)
   if err != nil { logutil.Error("Failed to marshal arguments data",err); return err }
	defer file.Close()
	_,err = file.Write(data)
	if err != nil { logutil.Error("Failed to write arguments data to file",err); return err }
	return err
}

func AppendLibsToArgumentsData(libs []string,natives string,version string) error {
   args := &argumentLookup{}
   var pathList = make([]classPathList,len(libs),len(libs))
   file,err := os.Open(filepath.Join(gamepath.Versionsdir,version,version+".json"))
   if err != nil { logutil.Error("Failed to open arguments data file",err); return err }
   data,err := io.ReadAll(file)
   if err != nil { logutil.Error("Failed to read arguments data file",err); return err }
   err = json.Unmarshal(data,args)
   if err != nil { logutil.Error("Failed to unmarshal arguments data",err); return err }
   defer file.Close()
   for i := range libs {
      pathList[i].Path = libs[i]
   }
   args.Libraries = pathList
   pathList[len(pathList)-1].Native = natives
   jsonData,err := json.Marshal(args)
   if err != nil { logutil.Error("Failed to marshal arguments data",err); return err }
   overwrittenFile,err := os.OpenFile(filepath.Join(gamepath.Versionsdir,version,version+".json"),os.O_WRONLY|os.O_CREATE|os.O_TRUNC,os.ModePerm)
   _,err = overwrittenFile.Write(jsonData)
   if err != nil { logutil.Error("Failed to write arguments data to file",err); return err }
   overwrittenFile.Close()
   return err
}

func GetVersionList() error {
	logutil.Info("Acquiring version list")
	jsonBytes,err := downloadutil.GetData(versionMeta); if err != nil {
      searchErr := searchLocalVersions(true)
      if searchErr != nil { logutil.Error("Failed to search local versions",searchErr); return searchErr }
		logutil.Error("Failed to get data for version",err)
		return err
	}
	LatestRelease = gjson.Get(string(jsonBytes),"latest.release").String()
	LatestSnapshot = gjson.Get(string(jsonBytes),"latest.snapshot").String()
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
   err = searchLocalVersions(false)
   if err != nil { logutil.Error("Failed to search local versions",err); return err }
	return err
}

func searchLocalVersions(offlineMode bool) error {
	var names []string
	dirEntry,err := os.ReadDir(gamepath.Versionsdir)
	if err != nil {
		logutil.Error("Failed to read directory contents",err)
		return err
	}
	for _,file := range dirEntry {
		if file.IsDir() {
         versionDir,err := os.ReadDir(filepath.Join(gamepath.Versionsdir,file.Name()))
         if err != nil {
            logutil.Error("Failed to read directory contents",err)
            return err
         }
         for _,f := range versionDir {
            if !f.IsDir() && filepath.Ext(filepath.Join(gamepath.Versionsdir,file.Name(),f.Name())) == ".json" {
               filename := file.Name()
               names = append(names,filename[:len(filename)-5])
               if offlineMode {
                  VersionList["Local"] = names
               } else {
                  var optifine []string
                  for _,e := range names {
                     if strings.Contains(e,"OptiFine") {
                        optifine = append(optifine, e)
                     }
                  }
                  if optifine != nil {
                     VersionList["OptiFine"] = optifine 
                  }
               }
            }
         }
		}
	}
	return err
}

func GetVersionType(version string) string {
	var vertype string
	jsonBytes,err := downloadutil.GetData(versionMeta)
	if err != nil { logutil.Error("Failed to get data for version",err); return "" }
	gjson.Get(string(jsonBytes),"versions").ForEach(func(_, value gjson.Result) bool {
		if value.Get("id").String() == version {
			vertype = value.Get("type").String()
			return true
		}
		return true
	})
	return vertype
}

func SortVersionTypes(slice []string) []string {
	order := map[int]string{
		0:"release",
		1:"snapshot",
		2:"old_beta",
		3:"old_alpha",
	}
   for i := 0; i < len(order); i++ {
		slice[i] = order[i]
	}
	return slice
}

func ParseVersion(url string) (gjson.Result,error) {
	versionData,err := downloadutil.GetData(url)
	if err != nil { logutil.Error("Failed to get version data",err); return gjson.Result{},err }
	return gjson.Parse(string(versionData)),err
}
