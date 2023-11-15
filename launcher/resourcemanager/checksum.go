package resourcemanager

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"io"

	"egreg10us/faultylauncher/util/logutil"
)

func ValidateChecksum(path string,hash string) bool {
	info,_ := os.Stat(path); if info.IsDir() {
		return true
	}
	file,err := os.Open(path); if err != nil { logutil.Error(err.Error()) }
	defer file.Close()
	generatedHash := sha1.New()
	if _,err := io.Copy(generatedHash,file); err != nil { logutil.Error(err.Error()) }
	formattedHash := hex.EncodeToString(generatedHash.Sum(nil))

	if formattedHash == hash {
		logutil.Info("Hash "+formattedHash+" == "+hash)
		return true
	} else {
		logutil.Warn("Hash "+formattedHash+" != "+hash)
		return false
	}
}
