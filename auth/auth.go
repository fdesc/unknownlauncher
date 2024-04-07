package auth

import (
	"path/filepath"
	"encoding/json"
	"io"
	"os"

	"fdesc/unknownlauncher/util/logutil"
	"fdesc/unknownlauncher/util/gamepath"
)

type AccountsRoot struct {
	Accounts map[string]AccountProperties `json:"accounts"`
	LastUsed string `json:"lastUsed,omitempty"`
}

type AccountProperties struct {
	Name string `json:"name"`
	AccountType string `json:"type"`
	AccountUUID string `json:"uuid"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

func ReadAccountsRoot() (AccountsRoot,error) {
	logutil.Info("Reading accounts.json")
	file,err := os.Open(filepath.Join(gamepath.Gamedir,"accounts.json"))
	if err != nil {
		if os.IsNotExist(err) {
			emptyroot := AccountsRoot{Accounts:nil}
			_,err := os.Create(filepath.Join(gamepath.Gamedir,"accounts.json"))
			if err != nil { logutil.Error("Failed to create accounts file",err); return AccountsRoot{},err }
			emptyroot.SaveToFile()
			return emptyroot,nil
		}
		logutil.Error("Failed to open accounts file",err); return AccountsRoot{},err
	}
	defer file.Close()
	read,err := io.ReadAll(file)
	if err != nil { logutil.Error("Failed to read contents of accounts.json",err); return AccountsRoot{},err }
	readAccountsRoot := AccountsRoot{}
	err = json.Unmarshal(read,&readAccountsRoot)
	if err != nil {
		logutil.Error("Failed to unmarshal accounts.json",err); return AccountsRoot{},err
	}
	return readAccountsRoot,err
}

func (aRoot *AccountsRoot) SaveToFile() error {
	jsonData,err := json.MarshalIndent(aRoot,"","  ")
	if err != nil { return err }
	file,err := os.Create(filepath.Join(gamepath.Gamedir,"accounts.json"))
	if err != nil { return err }
	defer file.Close()
	_,err = file.Write(jsonData)
	if err != nil { return err }
	return err
}
