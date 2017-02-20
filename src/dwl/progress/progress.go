package progress

import (
	"time"
)

type Progress []DownloadProgress

func (p Progress) GetSpeed() (uint64, uint64) {
	var speed uint64 = 0
	var mspeed uint64 = 0

	for _, dp := range p {
		s, ms := dp.GetSpeed()
		speed += s
		mspeed += ms
	}

	return speed, mspeed
}

func (p Progress) GetPercent() int {
	prc := 0
	for _, dp := range p {
		prc += dp.GetPercent()
	}
	return prc / len(p)
}

func (p Progress) GetAverageConnTime() time.Duration {
	var ret time.Duration
	count := 0
	for _, dp := range p {
		if dp.ConnectTime > 0 {
			count++
			ret += dp.ConnectTime
		}
	}
	return ret / time.Duration(count)
}

func (p Progress) Complete() bool {
	for _, dp := range p {
		if !dp.Complete() {
			return false
		}
	}
	return true
}

func (p Progress) IsLoading() bool {
	for _, dp := range p {
		if dp.IsLoading {
			return true
		}
	}
	return false
}
