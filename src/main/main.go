package main

import (
	"dwl"
	"dwl/progress"
	"dwl/settings"

	"fmt"
	"sync"
	"time"

	ui "github.com/gizak/termui"
)

func main() {
	sets := settings.NewSettings()
	//	sets.Url = "ftp://prep.ai.mit.edu/README"
	//	sets.Url = "http://www.audiocheck.net/download.php?filename=Audio/audiocheck.net_hdchirp_88k_-3dBFS_lin.wav"
	//	sets.Url = "http://ovh.net/files/1Mio.dat"
	sets.Threads = 5
	sets.FilePath = "test.file"
	sets.Config.Set("Timeout", 0)
	header := settings.NewConfig().Set("Accept", "*/*")
	sets.Config.Set("Header", header)

	sets.Url = "test://localhost:8090/files/text10.txt"
	//	sets.Url = "ftp://prep.ai.mit.edu/README"
	//	sets.Url = "ftp://ftpprd.ncep.noaa.gov/pub/data/nccf/radar/nexrad_level2/KABR/KABR_20170209_025154.bz2"
	//	sets.Url = "ftp://ftp.uconn.edu/48_hour/info.zip"

	d := dwl.NewDWL(sets)

	ui.Init()
	defer ui.Close()
	var wa sync.WaitGroup
	var err error

	wa.Add(1)
	go func() {
		err = d.Load()
		if err != nil {
			lblErr := ui.NewPar(err.Error())
			lblErr.SetX(4)
			lblErr.SetY(0)
			lblErr.Width = 100
			lblErr.Height = 3
			lblErr.BorderLabel = "Error"
			ui.Render(lblErr)
		}
		ui.StopLoop()
	}()

	go func() {
		for {
			dp := d.GetProgress()
			if dp != nil {
				print(dp, sets.Url)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		d.Stop()
		ui.StopLoop()
	})
	go func() {
		ui.Loop()
		wa.Done()
	}()
	wa.Wait()
	print(d.GetProgress(), sets.Url)
	i := 1
	if err != nil {
		i = 3
	}
	time.Sleep(time.Second * time.Duration(i))
}

func print(dp []progress.DownloadProgress, url string) {
	if dp == nil {
		return
	}
	parts := make([]int, 0)
	var speed uint64 = 0
	var mspeed uint64 = 0
	for _, p := range dp {
		parts = append(parts, p.GetPercent())
		s, m := p.GetSpeed()
		speed += s
		mspeed += m
	}
	gg := make([]*ui.Gauge, len(parts))

	width := 100

	for i, _ := range gg {
		gg[i] = ui.NewGauge()
		gg[i].Percent = parts[i]
		gg[i].Width = width / len(parts)
		gg[i].Height = 1
		gg[i].BarColor = ui.ColorRed
		gg[i].PercentColor = ui.ColorBlue
		gg[i].X = 4 + i*gg[i].Width
		gg[i].Y = 4
		gg[i].Border = false
		ui.Render(gg[i])
	}

	lblSpeed := ui.NewPar("")
	lblSpeed.SetX(4)
	lblSpeed.SetY(5)
	lblSpeed.Width = 120
	lblSpeed.Height = 5
	lblSpeed.BorderLabel = "Progress"
	lblSpeed.Text = "Url: " + url
	lblSpeed.Text += fmt.Sprintf("\nAver Speed: %v/sec", progress.ByteSize(mspeed))
	lblSpeed.Text += fmt.Sprintf("\nReal Speed: %v/sec", progress.ByteSize(speed))
	for i, p := range parts {
		_, s := dp[i].GetSpeed()
		ct := dp[i].ConnectTime.Seconds()
		bl := uint64(dp[i].Pos - dp[i].From)
		lblSpeed.Text += fmt.Sprintf("\nPart: %v Procent: %v%% Speed: %v/sec ConnTime: %.2fsec Loaded: %v", i, p, progress.ByteSize(s), ct, progress.ByteSize(bl))
		lblSpeed.Height++
	}
	ui.Render(lblSpeed)
}
