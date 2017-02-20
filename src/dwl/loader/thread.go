package loader

import (
	"time"
)

func (l *Loader) isRunLoad() bool {
	if l.sets.Threads != -1 && l.getCurrThreads() >= l.sets.Threads && l.isLoading {
		return false
	}
	return true
}

func (l *Loader) isEndLoad() bool {
	if len(l.parts) == 0 {
		return false
	}
	for _, p := range l.parts {
		if !p.Complete() {
			return false
		}
	}
	return true
}

func (l *Loader) loadPart() {
	part := l.getPart()
	if part != nil {
		err := part.LoadPart(l.sets, l.GetProgress)
		if err != nil {
			l.err = err
			l.Stop()
		}
	}
}

func (l *Loader) getPart() *Part {
	l.partMut.Lock()
	defer l.partMut.Unlock()

	//if no parts create new
	if len(l.parts) == 0 {
		part := NewPart(l.sets, l.file, 0, 0, 0)
		l.parts = append(l.parts, part)
		return part
	}

	//find first not loaded or is cut
	for _, p := range l.parts {
		//is not load and not end
		if !p.IsLoading && !p.Complete() {
			return p
		}
		//is cut
		if p.IsLoading {
			if l.isCut(p) {
				part := l.cut(p)
				l.parts = append(l.parts, part)
				return part
			}
		}
	}

	//find biggest part and cutit
	part := l.findBiggest()
	if part != nil && part.IsLoading && l.isCut(part) {
		part := l.cut(part)
		l.parts = append(l.parts, part)
		return part
	}
	return nil
}

func (l *Loader) getCurrThreads() int {
	if len(l.parts) == 0 {
		return 0
	}
	comp := 0
	work := 0
	for _, p := range l.parts {
		if !p.DownloadProgress.Complete() {
			comp++
		}
		if p.DownloadProgress.IsLoading {
			work++
		}
	}

	if work < comp {
		return work
	}
	return comp
}

func (l *Loader) cut(p *Part) *Part {
	sets := *p.settings

	size := p.To - p.Pos
	end := p.To
	p.To -= size / 2

	pn := NewPart(&sets, l.file, p.To, end, p.To)
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

func (l *Loader) getConnBuf(speed uint64) int64 {
	buff := l.sets.LoadBufferSize * 2

	var connTime time.Duration
	count := 0
	for _, p := range l.parts {
		if p.ConnectTime > 0 {
			count++
			connTime += p.ConnectTime
		}
	}
	connTime = connTime / time.Duration(count)
	return int64(float64(speed)*connTime.Seconds()) + buff
}

func (l *Loader) isCut(dp *Part) bool {
	_, speed := dp.GetSpeed()

	if speed == 0 {
		return false
	}
	buff := l.getConnBuf(speed)
	return buff < dp.To-dp.Pos
}

func (l *Loader) WaitForNextLoading() {
	//calc wait time for next loading
	wait := time.Second * 5
	for _, p := range l.parts {
		_, sp := p.GetSpeed()
		fe := p.To - p.Pos
		if sp == 0 || fe == 0 {
			continue
		}
		pp := float64(fe*1000) / float64(sp)
		if wait.Seconds()*1000 > float64(pp) {
			wait = time.Millisecond * time.Duration(pp)
		}
	}
	if wait.Seconds() < 0.100 {
		wait = time.Millisecond * 100
	}
	//wait
	time.Sleep(wait)
}
