package progress

import (
	"fmt"
	"time"
)

type DownloadProgress struct {
	From int64 `json:"From"`
	To   int64 `json:"To"`
	Pos  int64 `json:"Pos"`

	BytesLoaded uint64    `json:"-"`
	Speed       uint64    `json:"-"`
	MiddleSpeed uint64    `json:"-"`
	LastTime    time.Time `json:"-"`
	StartTime   time.Time `json:"-"`

	ConnectTime time.Duration `json:"-"`
	IsLoading   bool          `json:"-"`
}

func (s *DownloadProgress) StartSpeed() {
	s.Speed = 0
	s.MiddleSpeed = 0
	s.BytesLoaded = 0
	s.StartTime = time.Now()
	s.LastTime = time.Now()
}

func (s *DownloadProgress) EndSpeed() {
	s.Speed = 0
	s.MiddleSpeed = 0
	s.BytesLoaded = 0
}

func (s *DownloadProgress) GetSpeed() (uint64, uint64) {
	return s.Speed, s.MiddleSpeed
}

func (s *DownloadProgress) MessureSpeed(realc int) {
	s.BytesLoaded += uint64(realc)

	delta := time.Since(s.StartTime).Seconds()
	if time.Since(s.LastTime).Seconds() > 0.1 {
		s.LastTime = time.Now()
		lstSpeed := s.Speed
		s.Speed = uint64(float64(s.BytesLoaded) / delta)
		s.MiddleSpeed = (s.MiddleSpeed + (s.Speed+lstSpeed)/2) / 2
	}
	if time.Since(s.StartTime).Seconds() > 5 {
		s.StartTime = time.Now()
		s.LastTime = time.Now()
		s.BytesLoaded = 0
	}
}

func (s *DownloadProgress) Complete() bool {
	return s.Pos >= s.To && s.To > 0
}

func (s *DownloadProgress) GetPercent() int {
	if s.To < 1 {
		return -1
	}

	loaded := s.Pos - s.From
	all := s.To - s.From

	return int(loaded * 100 / all)
}

func (s DownloadProgress) String() string {
	_, sm := s.GetSpeed()
	return fmt.Sprintf("P: %v %v %v %v%% %v/s", s.From, s.Pos, s.To, s.GetPercent(), ByteSize(uint64(sm)))
}
