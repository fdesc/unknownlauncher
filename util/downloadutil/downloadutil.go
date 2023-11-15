package downloadutil

import (
	"path/filepath"
	"net/http"
	"runtime"
	"sync"
	"time"
	"os"
	"io"

	"egreg10us/faultylauncher/util/logutil"
)

var client = http.Client{Timeout: 120 * time.Second}

func DownloadSingle(url,path string) error {
	if _,err := os.Stat(path); err == nil { return err }
	err := os.MkdirAll(filepath.Dir(path),os.ModePerm); if err != nil { logutil.Critical(err.Error()); return err }
	out, err := os.Create(path); if err != nil { logutil.Critical(err.Error()); return err }
	defer out.Close()
	time.Sleep(150 * time.Millisecond)
	resp, err := client.Get(url)
	if err != nil {
		if os.IsTimeout(err) {
			logutil.Error("Timeout while downloading "+filepath.Base(path)+" errors and game crashes are expected.")
		}
		return err
	}
	defer resp.Body.Close()
	_,err = io.Copy(out, resp.Body)
	logutil.Info("Downloaded "+filepath.Base(path))
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
			err := os.MkdirAll(filepath.Dir(path),os.ModePerm); if err != nil { logutil.Critical(err.Error()); return err }
			out, err := os.Create(path); if err != nil { logutil.Critical(err.Error()); return err }
			defer out.Close()
			time.Sleep(150 * time.Millisecond)
			resp, err := client.Get(url)
			if err != nil {
				if os.IsTimeout(err) {
					logutil.Error("Timeout while downloading "+filepath.Base(path)+" errors and game crashes are expected.")
				}
				return err
			}
			defer resp.Body.Close()
			_,err = io.Copy(out, resp.Body)
			logutil.Info("Downloaded "+filepath.Base(path))
			runtime.UnlockOSThread()
			return err
		}(urlSlice[i],pathSlice[i],&wg)
	}
	wg.Wait()
}

func GetData(url string) ([]byte,error) {
	resp, err := client.Get(url); if err != nil { logutil.Error(err.Error()); return nil,err }
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body,nil
}
