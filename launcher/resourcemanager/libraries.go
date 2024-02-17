package resourcemanager

import (
	"path/filepath"
	"archive/zip"
	"strconv"
	"strings"
	"regexp"
	"time"
	"os"
	"io"

	"github.com/tidwall/gjson"
	"fdesc/unknownlauncher/util/downloadutil"
	"fdesc/unknownlauncher/util/gamepath"
	"fdesc/unknownlauncher/util/logutil"
)

var (
	pathSlice []string
	urlSlice []string
	hashSlice []string
)

func libraryRules(versiondata *gjson.Result) {  
	logutil.Info("Acquiring OS dependent library rules")
	versiondata.Get("libraries").ForEach(func(_,value gjson.Result) bool {
		if (value.Get("rules.1").Exists() && !value.Get("natives").Exists()) {
			if value.Get("rules.1.os.name").String() == identifier {
					if value.Get("rules.1.action").String() == "allow" {
						updateSlices(getPkgPath(value.Get("name").String()))
						hashSlice = append(hashSlice, value.Get("downloads.artifact.sha1").String())
					}
				} else {
					updateSlices(getPkgPath(value.Get("name").String()))
					hashSlice = append(hashSlice, value.Get("downloads.artifact.sha1").String())
				}
				if !(value.Get("rules.0").Get("action").String() == "allow" && value.Get("rules.0").Get("os").Get("name").String() == identifier) {
					pkgname := getLastPkgElem(value.Get("name").String())
					if (pkgname == "natives-"+identifier &&
					    pkgname != "natives-"+identifier+"-"+identifierArch &&
					    pkgname != identifier+"-"+identifierArch) {
						updateSlices(getPkgPath(value.Get("name").String()))
						hashSlice = append(hashSlice, value.Get("downloads.artifact.sha1").String())
					}
				}
			} 
		return true
	})
}

func defaultLibraries(versiondata *gjson.Result) {
	logutil.Info("Acquiring default libraries")
	versiondata.Get("libraries").ForEach(func (_,value gjson.Result) bool{
		if (!value.Get("rules").Exists() && value.Get("downloads.artifact").Exists()) {
			updateSlices(getPkgPath(value.Get("name").String()))
			hashSlice = append(hashSlice, value.Get("downloads.artifact.sha1").String())
		}
		return true
	})
}

func nativeLibraries(versiondata *gjson.Result) []string {
	logutil.Info("Acquiring native libraries")
	versiondata.Get("libraries").ForEach(func (_,value gjson.Result) bool {
		if value.Get("natives").Exists() {
			nativeKey := value.Get("natives").Get(identifier).String()
			replaceIdentifier := regexp.MustCompile(`-\${arch}`)
			replaced := replaceIdentifier.ReplaceAllString(nativeKey,"-"+identifierArchOld)
			if value.Get("rules.1.action").String() == "disallow" && value.Get("rules.1.os.name").String() != identifier {
				updateSlices(getPkgPath(value.Get("name").String()+":"+replaced))
				hashSlice = append(hashSlice, value.Get("downloads.classifiers").Get(replaced).Get("sha1").String())
			} else if value.Get("rules.0.os.name").String() == identifier {
				updateSlices(getPkgPath(value.Get("name").String()+":"+replaced))
				hashSlice = append(hashSlice, value.Get("downloads.classifiers").Get(replaced).Get("sha1").String())
			}
			if !value.Get("rules").Exists() {
				updateSlices(getPkgPath(value.Get("name").String()+":"+nativeKey))
				hashSlice = append(hashSlice, value.Get("downloads.classifiers").Get(nativeKey).Get("sha1").String())
			}
		}
		if (value.Get("rules.0.action").String() == "allow" && value.Get("rules.0.os.name").String() == identifier) {
			pkgname := getLastPkgElem(value.Get("name").String())
			if (pkgname == "natives-"+identifier ||
			    pkgname == "natives-"+identifier+"-"+identifierArch ||
			    pkgname == identifier+"-"+identifierArch) {
				updateSlices(getPkgPath(value.Get("name").String()))
				hashSlice = append(hashSlice, value.Get("downloads.artifact.sha1").String())
			}
		}
		return true
	})
	return pathSlice // slice of natives path
}

func unpackNatives(version string,nativesSlice []string) (string) {
	logutil.Info("Unpacking natives")
	target := filepath.Join(gamepath.Versionsdir,version,version+"-natives-"+strconv.Itoa(time.Now().Nanosecond()))
	for i := range nativesSlice {
		jarFile,err := zip.OpenReader(pathSlice[i])
		if err != nil { logutil.Error("Failed to read JAR file",err) }
		defer jarFile.Close()
		for _,file := range jarFile.File {
			openedFiles, err := file.Open(); if err != nil { logutil.Error("Failed to open file",err) }
			defer openedFiles.Close()
			os.MkdirAll(filepath.Dir(filepath.Join(target,file.Name)),os.ModePerm)
			if file.FileInfo().IsDir() {
				continue
			} else {
				f, err := os.Create(filepath.Join(target,filepath.Base(file.Name)))
				if err != nil { logutil.Error("Failed to create file",err) }
				defer f.Close()
				if _,err = io.Copy(f,openedFiles); err != nil { logutil.Error("Failed to copy data",err) }
			}
		}
	}
	fileList,err := os.ReadDir(target)
	if err != nil { logutil.Error("Failed to read directory contents",err) }
	for _,file := range fileList {
		currentFile := filepath.Join(target,file.Name())
		if file.IsDir() {
			os.RemoveAll(currentFile)
		}
		if (filepath.Ext(currentFile) != ".so" && filepath.Ext(currentFile) != ".dll" && filepath.Ext(currentFile) != ".dylib") {
			os.RemoveAll(currentFile)
		}
	}
	return target
}

func updateSlices(path string) error {
	pathSlice = append(pathSlice, gamepath.Librariesdir+gamepath.P+path+".jar")
	urlSlice = append(urlSlice, "https://libraries.minecraft.net/"+path+".jar")
	return nil
}

func getLastPkgElem(pkgname string) string {
	pkgNameSplitted := strings.Split(pkgname,":")
	return pkgNameSplitted[len(pkgNameSplitted)-1]
}

func getPkgPath(pkgname string) string {
	pkgNameSplitted := strings.Split(pkgname,":")
	pkgNameSplitted[0] = strings.ReplaceAll(pkgNameSplitted[0],".","/")
	if (strings.Contains(pkgNameSplitted[len(pkgNameSplitted)-1],"natives") || strings.Contains(pkgNameSplitted[len(pkgNameSplitted)-1],identifier)) {
		nativeKey := pkgNameSplitted[len(pkgNameSplitted)-1]
		pkgNameSplitted = append(pkgNameSplitted[:len(pkgNameSplitted)-1], pkgNameSplitted[len(pkgNameSplitted):]...)
		pkgNameSplitted = append(pkgNameSplitted, strings.Join(pkgNameSplitted[1:], "-")+"-"+nativeKey)
		path := strings.Join(pkgNameSplitted, "/")
		return path
	}
	pkgNameSplitted = append(pkgNameSplitted, strings.Join(pkgNameSplitted[1:],"-"))
	path := strings.Join(pkgNameSplitted,"/")
	return path
}

func Libraries(version string, versiondata *gjson.Result) ([]string,string) {
	setIdentifier()
	logutil.Info("Downloading libraries")
	natives := nativeLibraries(versiondata)
	defaultLibraries(versiondata)
	libraryRules(versiondata)
	downloadutil.DownloadMultiple(urlSlice,pathSlice)
	for i := range pathSlice {
		if !ValidateChecksum(pathSlice[i],hashSlice[i]) {
			os.Remove(pathSlice[i])
			downloadutil.DownloadSingle(urlSlice[i],pathSlice[i])
		} else {
			continue
		}
	}
	var nativesPath string
	if natives != nil {
		nativesPath = unpackNatives(version,natives)
	}
	logutil.Info("Finished downloading libraries")
	return pathSlice,nativesPath
}

func CleanLibraryList() {
	pathSlice = nil
	urlSlice = nil
	hashSlice = nil
}
