package gui

// TODO: move these to auth pkg after implementation
import (
	"fdesc/unknownlauncher/auth"
	"fdesc/unknownlauncher/launcher/profilemanager"
)

func GetAccountNames(accountsMap map[string]auth.AccountProperties) []string {
	s := []string{}
	for _,v := range accountsMap {
		s = append(s, v.Name)
	}
	return s
}

func GetAccountFromName(accountsMap map[string]auth.AccountProperties,name string) auth.AccountProperties {
	var data auth.AccountProperties
	for _,v := range accountsMap {
		if v.Name == name {
			data = v
		}
	}
	return data
}

func DeleteAccount(aRoot *auth.AccountsRoot,uuid string) error {
	delete(aRoot.Accounts,uuid)
	return aRoot.SaveToFile()
}

func GetProfileUUID(pMap *map[string]profilemanager.ProfileProperties,pData *profilemanager.ProfileProperties) string {
	var uuid string
	for k,v := range *pMap {
		if v.Created == pData.Created && v.LastUsed == pData.LastUsed {
			uuid = k
		}
	}
	return uuid
}

func GetProfileNames(pMap *map[string]profilemanager.ProfileProperties) []string {
	s := []string{}
	for _,v := range *pMap {
		if v.Name != "" {
			s = append(s,v.Name)
		} else {
			s = append(s,v.Type)
		}
	}
	return s
}

func DeleteProfile(pRoot *profilemanager.ProfilesRoot,uuid string) error {
	delete(pRoot.Profiles,uuid)
	return pRoot.SaveToFile()
}
