package auth

import (
	"image"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"fdesc/unknownlauncher/util/logutil"
)

type OfflineProfileResponse struct {
	Timestamp   int64  `json:"timestamp"`
	ProfileID   string `json:"profileId"`
	ProfileName string `json:"profileName"`
	Textures    struct {
		SKIN struct {
			URL      string `json:"url"`
			Metadata struct {
				Model string `json:"model"`
			} `json:"metadata"`
		} `json:"SKIN"`
		CAPE struct {
			URL string `json:"url"`
		} `json:"CAPE"`
	} `json:"textures"`
}

type UsernameResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ProfileData struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Properties []struct {
		Name      string `json:"name"`
		Value     string `json:"value"`
		Signature string `json:"signature"`
	} `json:"properties"`
	Legacy bool `json:"legacy"`
}

func InitializeClient() *http.Client {
	return &http.Client{Timeout: 30*time.Second}
}

func (aRoot *AccountsRoot) SaveOfflineAccount(username string) (image.Image,error) {
	skinsrc,uuid := PerformOfflineAuthentication(username)
	if uuid == "" {
		return nil,errors.New("Failed to get UUID please wait and try again later")
	}
	if aRoot.Accounts == nil {
		aRoot.Accounts = make(map[string]AccountProperties)
	}
	aRoot.Accounts[uuid] = AccountProperties{
		Name: username,
		AccountType: "offline",
		AccountUUID: uuid,
	}
	aRoot.LastUsed = uuid
	aRoot.SaveToFile()
	logutil.Info("Saved offline account with name: "+username)
	if skinsrc == "" {
		return nil,nil
	}
	return CropSkinImage(skinsrc),nil
}

func PerformOfflineAuthentication(username string) (string,string) {
	client := InitializeClient()
	uInformation,uStatus := GetUUIDFromUsername(client,username)
	if uStatus == 429 {
		client.CloseIdleConnections()
		return "",""
	}
	if uInformation.Id == "" {
		// use default skin
		logutil.Info("No skin found. using the default skin icon")
		DefaultSkinIcon = true
		client.CloseIdleConnections()
		return "",NewUUIDFromUsername("OfflinePlayer:"+username)
	} else {
		profileData,profileStatus := GetSkinData(client,uInformation.Id)
		if profileStatus != 200 {
			logutil.Warn("Failed to get skin data expected HTTP response 200")
			client.CloseIdleConnections()
			return "",uInformation.Id
		} else {
			client.CloseIdleConnections()
			return GetSkinUrl(profileData),uInformation.Id
		}
	}
}

// https://github.com/openjdk-mirror/jdk7u-jdk/blob/master/src/share/classes/java/util/UUID.java#L163
func NewUUIDFromUsername(username string) string {
	digestedHash := md5.Sum([]byte(username))
	digestedHash[6] = digestedHash[6] & 0x0f | 0x30 // clear version and set version to 3
	digestedHash[8] = digestedHash[8] & 0x3f | 0x80 // clear variant and set variant to IETF
	encoded := hex.EncodeToString(digestedHash[:])
	return encoded
}

func GetSkinUrl(pData ProfileData) string {
	if len(pData.Properties) < 1 {
		return ""
	}
	skinData,err := base64.StdEncoding.DecodeString(pData.Properties[0].Value)
	if err != nil { logutil.Error("Failed to decode profile properties of account",err); return "" }
	var skinUrl OfflineProfileResponse
	err = json.Unmarshal(skinData,&skinUrl)
	if err != nil { logutil.Error("Failed to unmarshal json data",err); return "" }
	return skinUrl.Textures.SKIN.URL
}

func GetUUIDFromUsername(client *http.Client, username string) (UsernameResponse,int) {
	req,err := http.NewRequest("GET","https://api.mojang.com/users/profiles/minecraft/"+username,nil)
	if err != nil { logutil.Error("Failed to create a GET request for getting UUID from username",err) }
	resp,err := client.Do(req)
	if err != nil { logutil.Error("Failed to perform created GET request",err); return UsernameResponse{},1 }
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		logutil.Warn("Too much requests for UUID")
		return UsernameResponse{},resp.StatusCode
	}
	var uResponse UsernameResponse
	jsonDecode := json.NewDecoder(resp.Body).Decode(&uResponse)
	if jsonDecode != nil { logutil.Error("Failed to decode response",err) }
	return uResponse,resp.StatusCode
}

func GetSkinData(client *http.Client,uuid string) (ProfileData,int) {
	if uuid == "" {
		return ProfileData{},404
	}
	req,err := http.NewRequest("GET","https://sessionserver.mojang.com/session/minecraft/profile/"+uuid,nil)
	if err != nil { logutil.Error("Failed to create GET request for getting profile from UUID",err) }
	resp,err := client.Do(req)
	if err != nil { logutil.Error("Failed to perform created GET request",err); return ProfileData{},404 }
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return ProfileData{},204
	}
	var pResponse ProfileData
	jsonDecode := json.NewDecoder(resp.Body).Decode(&pResponse)
	if jsonDecode != nil { logutil.Warn("Failed to decode response"); }
	return pResponse,resp.StatusCode
}
