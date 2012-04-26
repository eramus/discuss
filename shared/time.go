package shared

import (
	"fmt"
	"time"
)

// yah i dunno... late night coding
func FormatTime(t int64) (f string) {
	ts := time.Unix(t, 0)
	now := time.Now()
	diff := now.Unix() - ts.Unix()
	var unit string
	var m int64
	if diff < 60 {
		unit = "second"
		m = 1
	} else if diff < 3600 {
		unit = "minute"
		m = 60
	} else if diff < 86400 {
		unit = "hour"
		m = 3600
	} else if diff < 604800 {
		unit = "day"
		m = 86400
	} else if diff < 31556926 {
		unit = "month"
		m = 604800
	} else {
		unit = "year"
		m = 31556926
	}
	m = diff / m
	if m == 1 {
		if unit == "hour" {
			f = "an hour ago"
		} else {
			f = "a " + unit + " ago"
		}
	} else {
		f = fmt.Sprintf("%d %ss ago", m, unit)
	}
	return
}