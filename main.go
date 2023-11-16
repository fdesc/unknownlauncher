package main

import (

	"egreg10us/faultylauncher/launcher/versionmanager"
	"egreg10us/faultylauncher/launcher/resourcemanager"
	"egreg10us/faultylauncher/util/logutil"
)

func main() {
	wantedversion := &versionmanager.GameVersion{Version: "1.16.5",VersionType:"release"}
	url,ver,err := versionmanager.SelectVersion(wantedversion)
	if err != nil || url == "" {
		logutil.Error("Unknown game version or type "+err.Error())
	}

	versionData,err := versionmanager.ParseVersion(url)
	if err != nil { logutil.Error(err.Error()) }
	resourcemanager.Client(&versionData,ver)
	resourcemanager.Runtimes(&versionData)
	arl,aid := resourcemanager.GetAssetProperties(&versionData)
	assetsData,err := resourcemanager.ParseAssets(arl)
	if err != nil { logutil.Error(err.Error()) }
	err = resourcemanager.AssetIndex(arl,aid)
	if err != nil { logutil.Error(err.Error()) }
	resourcemanager.Assets(&assetsData,aid)
	resourcemanager.Log4JConfig(&versionData)
	resourcemanager.Libraries(wantedversion.Version,&versionData)
}
