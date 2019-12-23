package gone

import (
	"sort"
	"time"
)

type Report struct {
	Records Records
	Classes Classes
	Total   time.Duration
	Idle    time.Duration
}

type Record struct {
	Class string
	Name  string
	Spent time.Duration
	Idle  time.Duration
	Seen  time.Time
}

type Class struct {
	Class   string
	Spent   time.Duration
	Percent float64
}

type Records []Record

type Classes []Class

func (r Records) Len() int           { return len(r) }
func (r Records) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Records) Less(i, j int) bool { return r[i].Spent < r[j].Spent }

func (c Classes) Len() int           { return len(c) }
func (c Classes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Classes) Less(i, j int) bool { return c[i].Spent < c[j].Spent }

func GenerateReport(recorder *Recorder) Report {
	var idx Report
	classes := make(map[string]time.Duration)

	for window := range recorder.tracks.Keys() {
		rec, _ := recorder.tracks.Get(window)
		classes[window.Class] += rec.Spent
		idx.Total += rec.Spent
		idx.Idle += rec.Idle

		idx.Records = append(idx.Records, Record{
			Class: window.Class,
			Name:  window.Name,
			Spent: rec.Spent,
			Idle:  rec.Idle})
	}
	for k, v := range classes {
		idx.Classes = append(idx.Classes, Class{
			Class:   k,
			Spent:   v,
			Percent: 100.0 * float64(v) / float64(idx.Total)})
	}
	sort.Sort(sort.Reverse(idx.Classes))
	sort.Sort(sort.Reverse(idx.Records))
	return idx
}
