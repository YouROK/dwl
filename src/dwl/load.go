package dwl

import (
	"dwl/loader"
)

func (d *DWL) Load() error {
	if d.isLoading {
		return nil
	}
	d.isLoading = true
	var err error
	d.loader = d.getLoader()
	if err != nil {
		return err
	}
	err = d.loader.Connect(d.sets)
	if err != nil {
		return err
	}
	d.loader.Load()
	d.isLoading = false
	return d.loader.GetError()
}

func (d *DWL) Stop() {
	if !d.isLoading {
		return
	}
	d.isLoading = false
	d.loader.Stop()
}

func (d *DWL) getLoader() Loader {
	return loader.NewLoader()
}
