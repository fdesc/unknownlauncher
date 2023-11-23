package main

import (

	"egreg10us/faultylauncher/launcher/versionmanager"
	"egreg10us/faultylauncher/launcher/resourcemanager"
	"egreg10us/faultylauncher/util/logutil"
)

func main() {
	wantedversion := &versionmanager.GameVersion{Version: "23w46a",VersionType:"snapshot"}
	url,ver,err := versionmanager.SelectVersion(wantedversion)

	versionData,err := versionmanager.ParseVersion(url)
	if err != nil { logutil.Error("Failed to parse version data",err) }
	resourcemanager.Client(&versionData,ver)
	resourcemanager.Runtimes(&versionData)
	arl,aid := resourcemanager.GetAssetProperties(&versionData)
	assetsData,err := resourcemanager.ParseAssets(arl)
	if err != nil { logutil.Error("Failed to parse assets data",err) }
	err = resourcemanager.AssetIndex(arl,aid)
	if err != nil { logutil.Error("Failed to get asset index",err) }
	resourcemanager.Assets(&assetsData,aid)
	resourcemanager.Log4JConfig(&versionData)
	resourcemanager.Libraries(wantedversion.Version,&versionData)
}
