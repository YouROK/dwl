package dwl

import (
	"dwl/progress"
	"dwl/settings"
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

func (d *DWL) GetProgress() []progress.DownloadProgress {
	if d.loader != nil {
		return d.loader.GetProgress()
	}
	return nil
}
