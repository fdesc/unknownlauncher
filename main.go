package main

import (

	"egreg10us/faultylauncher/launcher/versionmanager"
	"egreg10us/faultylauncher/launcher/resourcemanager"
	"egreg10us/faultylauncher/util/logutil"
//	"egreg10us/faultylauncher/util/gamepath"
//	"egreg10us/faultylauncher/util/downloadutil"
)

// implement runtimes + fix problems + implement checker

func main() {
	url,_,err := versionmanager.SelectVersion(&versionmanager.GameVersion{Version: "1.8.9",VersionType:"release"})
	if err != nil || url == "" {
		logutil.Error("Unknown game version or type "+err.Error())
	}
	versionData,err := versionmanager.ParseVersion(url)
	/*if err != nil { logutil.Error(err.Error()) }
	arl,aid := resourcemanager.GetAssetProperties(&versionData)
	assetsData,err := resourcemanager.ParseAssets(arl)
	if err != nil { logutil.Error(err.Error()) }
	err = resourcemanager.Assets(&assetsData,aid)
	if err != nil { logutil.Error(err.Error()) }
	err = resourcemanager.AssetIndex(arl,aid)
	if err != nil { logutil.Error(err.Error()) }
	resourcemanager.Log4JConfig(&versionData)*/
	resourcemanager.Libraries(&versionData)
	// resourcemanager.Client(&versionData,version)
}
