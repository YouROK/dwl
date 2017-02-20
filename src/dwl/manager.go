package dwl

import (
	"dwl/progress"
	"dwl/settings"
	"runtime"
)

type DWL struct {
	sets      *settings.Settings
	loader    Loader
	size      int64
	isLoading bool
}

type OnChangeFunc func(int, []progress.DownloadProgress)

func NewDWL(sets *settings.Settings) *DWL {
	d := new(DWL)
	d.sets = sets
	if d.sets.Threads == -1 {
		d.sets.Threads = runtime.NumCPU() * 4
	}
	return d
}

func (d *DWL) Complete() bool {
	if d.loader == nil {
		return false
	}
	dp := d.loader.GetProgress()
	if dp == nil {
		return false
	}
	for _, p := range dp {
		if !p.Complete() {
			return false
		}
	}
	return true
}

func (d *DWL) GetProgress() progress.Progress {
	if d.loader != nil {
		return d.loader.GetProgress()
	}
	return nil
}
