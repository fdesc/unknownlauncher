package downloadutil

import (
	"path/filepath"
	"net/http"
	"runtime"
	"sync"
	"time"
	"os"
	"io"

	"egreg10us/unknownlauncher/util/logutil"
)

var client = http.Client{Timeout: 120 * time.Second}
var CurrentFile string

func DownloadSingle(url,path string) error {
	if _,err := os.Stat(path); err == nil { return err }
	err := os.MkdirAll(filepath.Dir(path),os.ModePerm); if err != nil { logutil.Error("Failed to create directory",err); return err }
	out, err := os.Create(path); if err != nil { logutil.Error("Failed to create file",err); return err }
	defer out.Close()
	time.Sleep(150 * time.Millisecond)
	resp, err := client.Get(url)
	if err != nil {
		if os.IsTimeout(err) {
			logutil.Warn("Timeout while downloading "+filepath.Base(path)+" errors and game crashes are expected.")
		}
		return err
	}
	defer resp.Body.Close()
	_,err = io.Copy(out, resp.Body)
	CurrentFile = filepath.Base(path)
	logutil.Info("Downloaded "+CurrentFile)
	return err
}

func DownloadMultiple(urlSlice,pathSlice []string) {
	var wg sync.WaitGroup
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := range urlSlice {
		wg.Add(1)
		go func(url,path string,receivedWg *sync.WaitGroup) error {
			runtime.LockOSThread()
			defer receivedWg.Done()
			if _,err := os.Stat(path); err == nil { return err }
			err := os.MkdirAll(filepath.Dir(path),os.ModePerm); if err != nil { logutil.Error("Failed to create directory",err); return err }
			out, err := os.Create(path); if err != nil { logutil.Error("Failed to create file",err); return err }
			defer out.Close()
			time.Sleep(150 * time.Millisecond)
			resp, err := client.Get(url)
			if err != nil {
				if os.IsTimeout(err) {
					logutil.Warn("Timeout while downloading "+path+" errors and game crashes are expected.")
				}
				return err
			}
			defer resp.Body.Close()
			_,err = io.Copy(out, resp.Body)
			CurrentFile = filepath.Base(path)
			logutil.Info("Downloaded "+CurrentFile)
			runtime.UnlockOSThread()
			return err
		}(urlSlice[i],pathSlice[i],&wg)
	}
	wg.Wait()
}

func GetData(url string) ([]byte,error) {
	resp, err := client.Get(url); if err != nil { logutil.Error("Failed to get response",err); return nil,err }
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body,err
}
