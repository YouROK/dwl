package file

import (
	"dwl/progress"
	"dwl/settings"

	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type File struct {
	*os.File
	lock sync.Mutex
}

func Open(filename string) (*File, error) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	ff := new(File)
	ff.File = f
	return ff, nil
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	f.lock.Lock()
	defer func() {
		f.File.Sync()
		f.lock.Unlock()
	}()
	return f.File.WriteAt(b, off)
}

type jsSave struct {
	Sets *settings.Settings          `json:"Settings"`
	Dp   []progress.DownloadProgress `json:"Progress"`
}

func (f *File) SaveState(sets *settings.Settings, dp []progress.DownloadProgress) error {
	if dp == nil || sets == nil {
		return nil
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	fn := f.Name() + ".dwl"

	isEnd := true
	for _, p := range dp {
		if p.To != p.Pos {
			isEnd = false
			break
		}
	}
	if isEnd {
		os.Remove(fn)
		return nil
	}

	js := new(jsSave)
	js.Sets = sets
	js.Dp = dp
	buf, err := json.MarshalIndent(js, "", " ")
	if err == nil {
		err = ioutil.WriteFile(fn, buf, 0666)
	}
	return err
}

func (f *File) LoadState() (*settings.Settings, []progress.DownloadProgress, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	fn := f.Name() + ".dwl"
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, nil, err
	}

	js := new(jsSave)
	err = json.Unmarshal(buf, js)
	if err != nil {
		return nil, nil, err
	}

	return js.Sets, js.Dp, nil
}

func (f *File) Close() error {
	defer f.Sync()
	return f.File.Close()
}
