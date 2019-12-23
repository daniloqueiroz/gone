// Gone Time Tracker -or- Where has my time gone?
package gone

import (
	"fmt"
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

type Recorder struct {
	x      Xorg
	tracks *TrackStore
	active *Window
	zzz    bool
}

func (r Recorder) Snooze(idle time.Duration) {
	if r.active != nil && !r.zzz && r.tracks.Has(r.active) {
		track, err := r.tracks.Get(r.active)
		if err == nil {
			track.Idle += idle
			r.tracks.Put(r.active, track)
		}
		r.zzz = true
	}
}

func (r Recorder) Wakeup() {
	if r.active != nil && r.zzz && r.tracks.Has(r.active) {
		track, err := r.tracks.Get(r.active)
		if err == nil {
			track.Seen = time.Now()
			r.tracks.Put(r.active, track)
		}
		r.zzz = false
	}
}

func (r Recorder) Update(w Window) {
	if !r.zzz {
		if r.active != nil && r.tracks.Has(&w) {
			track, _ := r.tracks.Get(&w)
			track.Spent += time.Since(track.Seen)
		}

		r.active = &w
		var newTrack *Track
		if !r.tracks.Has(&w) {
			newTrack, _ = r.tracks.Get(r.active)
		} else {
			newTrack = &Track{}
		}
		newTrack.Seen = time.Now()
	}
}

func (r Recorder) Start() {
	defer r.x.Close()
	r.x.Collect(r, time.Minute*5)
}

func NewTracker(display, trackingDir string) (*Recorder, error) {
	X := Connect(display)
	store, err := NewTrackStore(trackingDir)
	if err != nil {
		return nil, err
	}

	tracker := &Recorder{
		x:      X,
		tracks: store,
		active: nil,
		zzz:    false,
	}
	return tracker, nil
}
