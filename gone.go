// Gone Time Tracker -or- Where has my time gone?
package gone

//go:generate esc -o static.go static/

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

func (w Window) String() string {
	return fmt.Sprintf("%s %s", w.Class, w.Name)
}

type Track struct {
	Seen  time.Time
	Spent time.Duration
	Idle  time.Duration
}

func (t Track) String() string {
	return fmt.Sprintf("%s %s", t.Seen.Format("2006/01/02 15:04:05"), t.Spent)
}

type Tracker struct {
	x       Xorg
	tracks  map[Window]Track
	current Window
	zzz     bool
	file    string
}

func (t Tracker) Snooze(idle time.Duration) {
	if !t.zzz {
		if c, ok := t.tracks[t.current]; ok {
			c.Idle += idle
			t.tracks[t.current] = c
		}
		t.zzz = true
	}
}

func (t Tracker) Wakeup() {
	if t.zzz {
		if c, ok := t.tracks[t.current]; ok {
			c.Seen = time.Now()
			t.tracks[t.current] = c
		}
		t.zzz = false
	}
}

func (t Tracker) Update(w Window) {
	if !t.zzz {
		if c, ok := t.tracks[t.current]; ok {
			c.Spent += time.Since(c.Seen)
			t.tracks[t.current] = c
		}
	}

	if _, ok := t.tracks[w]; !ok {
		t.tracks[w] = Track{}
	}

	s := t.tracks[w]
	s.Seen = time.Now()
	t.tracks[w] = s

	t.current = w
}

func (t Tracker) removeSince(d time.Duration) {
	for k, v := range t.tracks {
		if time.Since(v.Seen) > d || v.Idle > d {
			delete(t.tracks, k)
		}
	}
}

func (t Tracker) load() error {
	dump, err := os.Open(t.file)
	if err != nil {
		return err
	}
	defer dump.Close()
	dec := gob.NewDecoder(dump)
	err = dec.Decode(&t.tracks)
	if err != nil {
		return nil
	}

	return nil
}

func (t Tracker) store() error {
	tmp := t.file + ".tmp"
	dump, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer dump.Close()
	enc := gob.NewEncoder(dump)
	err = enc.Encode(t.tracks)
	if err != nil {
		os.Remove(tmp)
		return err
	}
	os.Rename(tmp, t.file)
	return nil
}

func (t Tracker) Cleanup(every, since time.Duration) {
	tick := time.NewTicker(every)
	defer tick.Stop()
	for range tick.C {
		t.removeSince(since)
		t.store()
	}
}

func (t Tracker) Start() {
	defer t.store()
	defer t.x.Close()

	go t.Cleanup(time.Minute, time.Hour*24*7)
	t.x.Collect(t, time.Minute*5)
}

func NewTracker(display, trackingFile string) (*Tracker, error) {
	X := Connect(display)
	var w Window

	tracker := &Tracker{
		x:       X,
		tracks:  make(map[Window]Track),
		current: w,
		zzz:     false,
		file:    trackingFile,
	}
	err := tracker.load()
	if err != nil {
		return nil, err
	}
	return tracker, nil
}

