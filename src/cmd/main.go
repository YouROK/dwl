package main

import (
	"dwl"
	"dwl/progress"
	"dwl/settings"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
)

var (
	sets   *settings.Settings
	dwload *dwl.DWL
	err    error
)

func init() {
	sets = settings.NewSettings()

	flag.IntVar(&sets.Threads, "t", -1, "Loading threads")
	flag.Int64Var(&sets.LoadBufferSize, "b", 65536, "Thread buffer size")
	flag.StringVar(&sets.FilePath, "o", "", "Output file")
	flag.Parse()
	sets.Url = flag.Arg(0)
	if sets.Url == "" {
		flag.Usage()
		os.Exit(1)
	}

	if sets.FilePath == "" {
		u, _ := url.Parse(sets.Url)
		sets.FilePath = filepath.Base(u.Path)
		if sets.FilePath == "" {
			fmt.Println("Error parse filename, set manualy")
			os.Exit(1)
		}
	}
}

func main() {
	sets.Config.Set("Timeout", 0)
	dwload = dwl.NewDWL(sets)
	var wa sync.WaitGroup
	wa.Add(1)
	go func() {
		err = dwload.Load()
		if err == nil {
			err = errors.New("no errors, all ok")
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT, syscall.SIGABRT)
		<-c
		dwload.Stop()
	}()

	update()
}

func update() {
	var prog progress.Progress
	clear()
	for {
		println(5, 1, "Url:", sets.Url)
		prog = dwload.GetProgress()

		threads := prog.GetThreads()
		percent := prog.GetPercent()
		rspeed, aspeed := prog.GetSpeed()

		println(5, 3, "Threads:", threads)
		println(5, 4, "Percent:", percent)
		println(5, 5, "Aver Speed:", progress.ByteSize(aspeed)+"/sec")
		println(5, 6, "Real Speed:", progress.ByteSize(rspeed)+"/sec")

		sort.Slice(prog, func(i, j int) bool {
			return prog[i].From < prog[j].From
		})

		for i, p := range prog {
			_, s := p.GetSpeed()
			ct := p.ConnectTime.Seconds()
			bl := uint64(p.Pos - p.From)
			printf(6, 7+i, "Part: %v Procent: %v%% Speed: %v/sec ConnTime: %.2fsec Loaded: %v", i, p.GetPercent(), progress.ByteSize(s), ct, progress.ByteSize(bl))
		}

		if err != nil {
			println(5, 2, "Error:", err)
			break
		}
		if prog.Complete() {
			break
		}
	}
	setpos(0, len(prog)+8)
}

func clear() {
	fmt.Printf("\033[2J")
}

func clearline() {
	fmt.Printf("\033[K")
}

func setpos(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

func println(x, y int, v ...interface{}) {
	setpos(x, y)
	fmt.Print(v...)
	clearline()
}

func printf(x, y int, f string, v ...interface{}) {
	setpos(x, y)
	fmt.Printf(f, v...)
	clearline()
}
