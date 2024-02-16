package profilemanager

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"bytes"
	"time"
	"os"
	"io"

	"github.com/google/uuid"
	"egreg10us/unknownlauncher/launcher/versionmanager"
	"egreg10us/unknownlauncher/util/logutil"
	"egreg10us/unknownlauncher/util/gamepath"
)

type ProfilesRoot struct {
	Profiles map[string]ProfileProperties `json:"profiles"`
	LastUsedProfile string `json:"currentProfile,omitempty"`
}

type ProfileResolution struct {
	Height int `json:"height,omitempty"`
	Width int `json:"width,omitempty"`
	Fullscreen bool `json:"fullscreen,omitempty"`
}

type ProfileProperties struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Created string `json:"created"`
	LastUsed string `json:"lastUsed"`
	LastGameVersion string `json:"lastVersionId,omitempty"`
	LastGameType string `json:"lastVersionType,omitempty"`
	JVMArgs string `json:"javaArgs,omitempty"`
	Resolution *ProfileResolution `json:"resolution,omitempty"`
	GameDirectory string `json:"gameDir,omitempty"`
	SeparateInstallation bool `json:"separateInstallation,omitempty"`
	JavaDirectory string `json:"javaDir,omitempty"`
}

func ReadProfilesRoot() (ProfilesRoot,error) {
	logutil.Info("Reading the profiles file (launcher_profiles.json)")
	if gamepath.SeparateInstallation {
		gamepath.Mcdir()
	}
	file,err := os.Open(filepath.Join(gamepath.Gamedir,"launcher_profiles.json"))
	if err != nil {
		if os.IsNotExist(err) {
			logutil.Info("No profile file found. Creating launcher_profiles.json")
			file,err := os.Create(filepath.Join(gamepath.Gamedir,"launcher_profiles.json"))
			if err != nil {
				logutil.Error("Failed to create profile file",err)
				return ProfilesRoot{},err
			}
			defaultProfiles,err := CreateDefaultProfiles()
			if err != nil { 
				logutil.Error("Failed to create default profiles",err)
				return ProfilesRoot{},err
			}
			lastUsedProfileUUID := GetProfileUUID(&defaultProfiles,"latest-release")
			var pRoot = &ProfilesRoot{Profiles:defaultProfiles,LastUsedProfile:lastUsedProfileUUID}
			out,err := json.MarshalIndent(pRoot,"","  ")
			if err != nil { 
				logutil.Error("Failed to write json",err)
				return ProfilesRoot{},err
			}
			io.Copy(file,bytes.NewReader(out))
			return *pRoot,nil
		}
		logutil.Error("Failed to open launcher_profiles.json",err); return ProfilesRoot{},err
	}
	defer file.Close()
	read,err := io.ReadAll(file)
	if err != nil { logutil.Error("Failed to read contents of launcher_profiles.json",err); return ProfilesRoot{},err }
	readProfilesRoot := ProfilesRoot{}
	err = json.Unmarshal(read,&readProfilesRoot)
	if err != nil { logutil.Error("Failed to unmarshal the json data",err); return ProfilesRoot{},err }
	return readProfilesRoot,err
}

func CreateDefaultProfiles() (map[string]ProfileProperties,error) {
	logutil.Info("Creating default profiles")
	releaseProfile := &ProfileProperties{
		Name:"",
		Type:"latest-release",
		LastGameType: "release",
		LastGameVersion: versionmanager.LatestRelease,
		Created:time.Now().Format(time.RFC3339),
		LastUsed:time.Now().Format(time.RFC3339),
	}
	snapshotProfile := &ProfileProperties{
		Name:"",
		Type:"latest-snapshot",
		LastGameType: "snapshot",
		LastGameVersion: versionmanager.LatestSnapshot,
		Created:time.Now().Format(time.RFC3339),
		LastUsed:time.Now().Format(time.RFC3339),
	}
	rUUID,err := GenerateProfileUUID()
	if err != nil { return nil,err }
	sUUID,err := GenerateProfileUUID()
	if err != nil { return nil,err }
	profilesMap := map[string]ProfileProperties{
		rUUID:*releaseProfile,
		sUUID:*snapshotProfile,
	}
	return profilesMap,err
}

func (pRoot *ProfilesRoot) SaveToFile() error {
	jsonData,err := json.MarshalIndent(pRoot,"","  ")
	if err != nil { return err }
	file,err := os.Create(filepath.Join(gamepath.Gamedir,"launcher_profiles.json"))
	if err != nil { return err }
	defer file.Close()
	_,err = file.Write(jsonData)
	if err != nil { return err }
	return err
}

func GetProfileUUID(pData *map[string]ProfileProperties,profileInfo string) string {
	var uuid string
	for k,v := range *pData {
		if (v.Type == profileInfo || v.Name == profileInfo || v.Created == profileInfo) {
			uuid = k
		}
	}
	return uuid
}

func GenerateProfileUUID() (string,error) {
	generatedUUID,err := uuid.NewRandom()
	if err != nil { logutil.Error("Failed to generate random UUID for profile",err); return "",err }
	formattedUUID := strings.ReplaceAll(generatedUUID.String(),"-","")
	return formattedUUID,err
}

func (pData *ProfileProperties) SaveProfile() (map[string]ProfileProperties,error) {
	logutil.Info("Generating a new profile")
	pUUID,err := GenerateProfileUUID()
	if err != nil { return nil,err }
	savedProfile := map[string]ProfileProperties {pUUID:*pData}
	return savedProfile,err
}
