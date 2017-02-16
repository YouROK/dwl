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

	var wg sync.WaitGroup
	var mut sync.Mutex

	for i := 0; i < l.sets.Threads; i++ {
		wg.Add(1)
		go func() {
			for l.isLoading {
				mut.Lock()
				part := l.getPart()
				mut.Unlock()

				if part == nil {
					isEnd := true
					for _, p := range l.parts {
						if !p.Complete() {
							isEnd = false
						}
					}
					if isEnd && len(l.parts) > 0 {
						break
					}
					time.Sleep(time.Millisecond * 100)
					continue
				}
				err := part.LoadPart(l.sets, l.GetProgress)
				if err != nil {
					l.err = err
					l.Stop()
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
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

func (l *Loader) GetProgress() []progress.DownloadProgress {
	if len(l.parts) == 0 {
		return nil
	}
	ret := make([]progress.DownloadProgress, 0)
	for _, p := range l.parts {
		ret = append(ret, p.DownloadProgress)
	}
	return ret
}

func (l *Loader) GetError() error {
	return l.err
}

func (l *Loader) getPart() *Part {
	if len(l.parts) == 0 {
		p := NewPart(l.sets, l.file, 0, 0, 0)
		l.parts = append(l.parts, p)
		return p
	}

	for _, p := range l.parts {
		if p.ConnectTime > l.connTime {
			l.connTime = p.ConnectTime
		}
	}

	for _, p := range l.parts {
		//is not load and not end
		if !p.IsLoading && !p.Complete() {
			return p
		}
		if p.IsLoading && l.isCut(p.DownloadProgress, l.sets.LoadBufferSize*4) {
			pn := l.cut(p)
			return pn
		}

	}
	{
		part := l.findBiggest()
		//is load and isCut
		if part != nil && part.IsLoading && l.isCut(part.DownloadProgress, l.sets.LoadBufferSize*4) {
			pn := l.cut(part)
			return pn
		}
	}
	return nil
}

func (l *Loader) cut(p *Part) *Part {
	sets := *p.settings

	size := p.To - p.Pos
	end := p.To
	p.To -= size / 2

	pn := NewPart(&sets, l.file, p.To, end, p.To)
	l.parts = append(l.parts, pn)
	return pn
}

func (l *Loader) findBiggest() *Part {
	if len(l.parts) == 0 {
		return nil
	}
	max := l.parts[0]
	for _, p := range l.parts {
		if p.To-p.Pos > max.To-max.Pos {
			max = p
		}
	}
	return max
}

func (l *Loader) isCut(dp progress.DownloadProgress, buff int64) bool {
	_, speed := dp.GetSpeed()
	connTime := l.connTime //dp.ConnectTime
	if speed == 0 {
		return false
	}
	return int64(float64(speed)*connTime.Seconds())+buff < dp.To-dp.Pos
}
