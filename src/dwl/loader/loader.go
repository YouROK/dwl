package loader

import (
	"dwl/file"
	"dwl/progress"
	"dwl/settings"
	"sync"
	"time"
)

type Loader struct {
	sets      *settings.Settings
	parts     []*Part
	file      *file.File
	isLoading bool
	connTime  time.Duration
	err       error

	partMut sync.Mutex
}

func NewLoader() *Loader {
	h := new(Loader)
	return h
}

func (l *Loader) Connect(sets *settings.Settings) error {
	l.Stop()
	l.sets = sets
	l.err = nil
	l.isLoading = false
	l.parts = nil

	l.file, l.err = file.Open(l.sets.FilePath)
	if l.err == nil {
		lsets, ldp, err := l.file.LoadState()
		if err == nil {
			if l.sets.Url == "" {
				l.sets.Url = lsets.Url
			}
			if l.sets.Config == nil {
				l.sets.Config = lsets.Config
			}
			if l.sets.FilePath == "" {
				l.sets.FilePath = lsets.FilePath
			}
			if l.sets.Threads == 0 {
				l.sets.Threads = lsets.Threads
			}
			if l.sets.LoadBufferSize == 0 {
				l.sets.LoadBufferSize = lsets.LoadBufferSize
			}

			for _, p := range ldp {
				l.parts = append(l.parts, NewPart(l.sets, l.file, p.From, p.To, p.Pos))
			}
		}
	}

	return l.err
}

func (l *Loader) Load() {
	if l.isLoading {
		return
	}

	l.isLoading = true
	defer func() {
		l.isLoading = false
		l.file.Close()
	}()

	for l.isLoading {
		for l.isRunLoad() && l.isLoading {
			go l.loadPart()
			if l.isEndLoad() {
				l.isLoading = false
			}
			time.Sleep(time.Millisecond * 50)
		}

		if !l.isLoading {
			break
		}

		l.WaitForNextLoading()
	}
}

func (l *Loader) Stop() {
	l.isLoading = false
	for _, p := range l.parts {
		p.IsLoading = false
	}
}

func (l *Loader) Complete() bool {
	complete := true
	for _, p := range l.parts {
		if !p.Complete() {
			complete = false
			break
		}
	}
	return complete
}

func (l *Loader) GetProgress() progress.Progress {
	if len(l.parts) == 0 {
		return nil
	}
	var ret progress.Progress
	for _, p := range l.parts {
		ret = append(ret, p.DownloadProgress)
	}
	return ret
}

func (l *Loader) GetError() error {
	return l.err
}
