package loader

import (
	"dwl/file"
	"dwl/progress"
	"dwl/settings"
	"io"
	"time"
)

type Part struct {
	settings *settings.Settings
	file     *file.File
	progress.DownloadProgress
}

func NewPart(sets *settings.Settings, file *file.File, From, To, Pos int64) *Part {
	p := new(Part)
	p.settings = sets
	p.file = file
	p.From = From
	p.To = To
	p.Pos = Pos
	return p
}

func (p *Part) LoadPart(sets *settings.Settings, getProgress func() []progress.DownloadProgress) error {
	startTime := time.Now()
	if p.IsLoading {
		return nil
	}
	p.IsLoading = true
	defer func() { p.IsLoading = false }()

	client, err := GetClient(sets, p.Pos)
	if err != nil {
		return err
	}

	err = client.Connect()
	if err != nil {
		return err
	}

	defer func() {
		client.Close()
		client = nil
	}()

	if p.To == 0 {
		p.To = client.GetAllSize()
	}
	bufsize := p.settings.LoadBufferSize
	if bufsize == 0 {
		bufsize = 65560
	}
	buffer := make([]byte, bufsize)
	n := 0
	p.ConnectTime = time.Since(startTime)
	p.StartSpeed()
	for err == nil && p.IsLoading && !p.Complete() {
		if int64(len(buffer)) > p.To-p.Pos {
			buffer = buffer[:p.To-p.Pos]
		}
		p.Pos = client.Pos()
		n, err = client.Read(buffer)
		p.MessureSpeed(n)
		if p.file == nil {
			break
		}
		if n > 0 {
			p.file.SaveState(p.settings, getProgress())
			n, err = p.file.WriteAt(buffer, p.Pos)
			if err != nil {
				break
			}
			p.Pos = client.Pos()
		}
	}

	if err == io.EOF || err == nil {
		p.file.SaveState(p.settings, getProgress())
		err = nil
	}
	return err
}
