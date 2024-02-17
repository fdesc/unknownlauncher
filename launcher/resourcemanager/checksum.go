package resourcemanager

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"io"

	"fdesc/unknownlauncher/util/logutil"
)

var DisableValidation bool

func ValidateChecksum(path string,hash string) bool {
	if !DisableValidation {
		info,err := os.Stat(path); if info.IsDir() {
			return true
		}
		if err != nil { logutil.Error("Failed to get file status",err); return false }
		file,err := os.Open(path)
		if err != nil { logutil.Error("Failed to open file",err); return false }
		defer file.Close()
		generatedHash := sha1.New()
		if _,err := io.Copy(generatedHash,file); err != nil { logutil.Error("Failed to copy data",err); return false }
		formattedHash := hex.EncodeToString(generatedHash.Sum(nil))

		if formattedHash == hash {
			logutil.Info("Hash "+formattedHash+" == "+hash)
			return true
		} else {
			logutil.Warn("Hash "+formattedHash+" != "+hash+" ("+path+") ")
			return false
		}
	} else {
		return true
	}
}
