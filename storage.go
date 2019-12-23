package gone

import (
	"encoding/json"
	"fmt"
	"github.com/google/logger"
	"github.com/prologic/bitcask"
	"strings"
	"time"
)

const (
	separator = "=::="
)

type TrackStore struct {
	db *bitcask.Bitcask
}

func NewTrackStore(storageDir string) (*TrackStore, error) {
	opts := []bitcask.Option{
		bitcask.WithMaxDatafileSize(10 << 20),
		bitcask.WithSync(true),
	}
	bc, err := bitcask.Open(storageDir, opts...)
	if err != nil {
		logger.Fatalf("Unable to open bitcask: %v", err)
		return nil, err
	}
	store := &TrackStore{db: bc}
	go store.compact()
	return store, nil
}

func (t *TrackStore) compact() {
	tick := time.NewTicker(1 * time.Hour)
	defer tick.Stop()
	for range tick.C {
		// TODO remove old entries before merge
		t.db.Merge()
	}
}

func (t *TrackStore) getKey(win *Window) []byte {
	return []byte(fmt.Sprint(win.Class, separator, win.Name))
}

func (t *TrackStore) Put(win *Window, rec *Track) error {
	data, err := json.Marshal(rec)
	if err != nil {
		logger.Errorf("Unable to serialize track: %v", err)
		return err
	}
	return t.db.Put(t.getKey(win), data)
}

func (t *TrackStore) Get(win *Window) (*Track, error) {
	data, err := t.db.Get(t.getKey(win))
	if err != nil {
		logger.Errorf("Unable to retrieve track %s: %v", t.getKey(win), err)
		return nil, err
	}
	var track *Track
	err = json.Unmarshal(data, track)
	if err != nil {
		logger.Errorf("Unable to deserialize track: %v", err)
		return nil, err
	}
	return track, nil
}

func (t *TrackStore) Has(win *Window) bool {
	return t.db.Has(t.getKey(win))
}

func (t *TrackStore) Keys() chan *Window {
	keys := make(chan *Window)
	go func() {
		defer close(keys)
		for key := range t.db.Keys() {
			str := strings.Split(string(key), separator)
			keys <- &Window{
				Class: str[0],
				Name:  str[1],
			}
		}
	}()
	return keys
}
