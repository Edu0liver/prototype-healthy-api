package service

import (
	"encoding/json"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
)

// businessHours is the parsed form of automations.business_hours.
//
// JSON shape:
//
//	{
//	  "timezone": "America/Sao_Paulo",
//	  "windows": {
//	    "mon": [{"start": "09:00", "end": "18:00"}],
//	    "sat": [{"start": "09:00", "end": "12:00"}, {"start": "13:00", "end": "18:00"}]
//	  }
//	}
//
// Days absent from "windows" (or with an empty array) are closed. An
// empty/absent config means always open (24/7), preserving the prior
// behaviour unless a tenant explicitly configures hours.
type businessHours struct {
	Timezone string                  `json:"timezone"`
	Windows  map[string][]timeWindow `json:"windows"`
}

type timeWindow struct {
	Start string `json:"start"` // "HH:MM"
	End   string `json:"end"`   // "HH:MM" (exclusive)
}

// weekdayKeys maps time.Weekday (Sunday=0) to the lowercase keys used in config.
var weekdayKeys = [...]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

// withinBusinessHours reports whether now falls inside the agent's configured
// business hours. An unset/empty config is treated as always open.
func withinBusinessHours(raw database.JSONMap, now time.Time) bool {
	if len(raw) == 0 {
		return true
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return true
	}
	var bh businessHours
	if err := json.Unmarshal(b, &bh); err != nil {
		return true
	}
	if len(bh.Windows) == 0 {
		return true // configured but no windows defined → no restriction
	}

	loc := time.UTC
	if bh.Timezone != "" {
		if l, err := time.LoadLocation(bh.Timezone); err == nil {
			loc = l
		}
	}
	now = now.In(loc)

	wins := bh.Windows[weekdayKeys[int(now.Weekday())]]
	if len(wins) == 0 {
		return false // day not listed → closed
	}
	cur := now.Hour()*60 + now.Minute()
	for _, w := range wins {
		start, ok1 := parseHHMM(w.Start)
		end, ok2 := parseHHMM(w.End)
		if !ok1 || !ok2 {
			continue
		}
		if start <= cur && cur < end {
			return true
		}
	}
	return false
}

func parseHHMM(s string) (int, bool) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}
