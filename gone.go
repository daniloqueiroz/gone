// Gone Time Tracker -or- Where has my time gone?
package gone

//go:generate esc -o static.go static/

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Tracks map[Window]Track

type Track struct {
	Seen  time.Time
	Spent time.Duration
	Idle  time.Duration
}

var (
	tracks  Tracks
	current Window
	logger  *log.Logger
	zzz     bool
)

func (t Track) String() string {
	return fmt.Sprintf("%s %s", t.Seen.Format("2006/01/02 15:04:05"), t.Spent)
}

func (w Window) String() string {
	return fmt.Sprintf("%s %s", w.Class, w.Name)
}

func (t Tracks) Snooze(idle time.Duration) {
	if !zzz {
		logger.Println("away from keyboard, idle for", idle)
		if c, ok := t[current]; ok {
			c.Idle += idle
			t[current] = c
		}
		zzz = true
	}
}

func (t Tracks) Wakeup() {
	if zzz {
		logger.Println("back to keyboard")
		if c, ok := t[current]; ok {
			c.Seen = time.Now()
			t[current] = c
		}
		zzz = false
	}
}

func (t Tracks) Update(w Window) {
	if !zzz {
		if c, ok := t[current]; ok {
			c.Spent += time.Since(c.Seen)
			t[current] = c
		}
	}

	if _, ok := t[w]; !ok {
		t[w] = Track{}
	}

	s := t[w]
	s.Seen = time.Now()
	t[w] = s

	current = w
}

func (t Tracks) RemoveSince(d time.Duration) {
	for k, v := range t {
		if time.Since(v.Seen) > d || v.Idle > d {
			logger.Println(v, k)
			delete(t, k)
		}
	}
}

func Load(fname string) Tracks {
	t := make(Tracks)
	dump, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return t
	}
	defer dump.Close()
	dec := gob.NewDecoder(dump)
	err = dec.Decode(&t)
	if err != nil {
		log.Println(err)
	}
	return t
}

func (t Tracks) Store(fname string) {
	tmp := fname + ".tmp"
	dump, err := os.Create(tmp)
	if err != nil {
		log.Println(err)
		return
	}
	defer dump.Close()
	enc := gob.NewEncoder(dump)
	err = enc.Encode(t)
	if err != nil {
		log.Println(err)
		os.Remove(tmp)
		return
	}
	os.Rename(tmp, fname)
}

func (t Tracks) Cleanup(file string, every, since time.Duration) {
	tick := time.NewTicker(every)
	defer tick.Stop()
	for range tick.C {
		t.RemoveSince(since)
		t.Store(file)
	}
}

func StartTracker(display, trackingFile string) {
	X := Connect(display)
	defer X.Close()

	logger = log.New(ioutil.Discard, "", log.LstdFlags)

	tracks = Load(trackingFile)
	defer tracks.Store(trackingFile)

	go X.Collect(tracks, time.Minute*5)
	go tracks.Cleanup(trackingFile, time.Minute, time.Hour*8)

	select {}
}

func main() {
	var (
		display = flag.String("display", os.Getenv("DISPLAY"), "X11 display")
		file = flag.String("file", "/tmp/track.bin", "tracking file")
	)
	flag.Parse()
	StartTracker(*display, *file)
}
