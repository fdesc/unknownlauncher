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
var wg sync.WaitGroup

func DownloadSingle(url,path string,threaded bool) error {
	if _,err := os.Stat(path); err == nil { return err }
	if threaded {
		wg.Add(2)
		runtime.LockOSThread()
		defer wg.Done()
		defer wg.Done()
	}
	err := os.MkdirAll(filepath.Dir(path),os.ModePerm); if err != nil { logutil.Critical(err.Error()); return err }
	out, err := os.Create(path); if err != nil { logutil.Critical(err.Error()); return err }
	defer out.Close()
	resp, err := client.Get(url)
	if err != nil {
		if os.IsTimeout(err) {
			logutil.Error("Timeout while downloading "+filepath.Base(path)+" errors and game crashes are expected.")
			return err
		}
		return err
	}
	defer resp.Body.Close()
	_,err = io.Copy(out, resp.Body)
	logutil.Info("Downloaded "+filepath.Base(path))
	if threaded {
		runtime.UnlockOSThread()
	}
	return err
}

func DownloadMultiple(url,path []string) error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := range url {
		go DownloadSingle(url[i],path[i],true)
	}
	defer wg.Wait()
	return nil
}

func GetJSON(url string) ([]byte,error) {
	resp, err := client.Get(url); if err != nil { logutil.Error(err.Error()); return nil,err }
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body,nil
}
