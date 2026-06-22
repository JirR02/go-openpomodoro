package openpomodoro

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/go-logfmt/logfmt"
)

// Settings is a collection of user settings, which can come from a file, env
// var, or set from the client program.
type Settings struct {
	DailyGoal               int
	DefaultBreakDuration    time.Duration
	DefaultPomodoroDuration time.Duration
	DefaultTags             []string
}

// DefaultSettings are used as a starting point before settings are overridden
// by the user.
var DefaultSettings = Settings{
	DailyGoal:               0,
	DefaultBreakDuration:    5 * time.Minute,
	DefaultPomodoroDuration: 25 * time.Minute,
	DefaultTags:             []string{},
}

// SetDefaults fills in settings values from another setting struct if the
// existing values are considered to not be set yet.
func (s *Settings) SetDefaults(d *Settings) {
	if s.DailyGoal == 0 {
		s.DailyGoal = d.DailyGoal
	}

	if s.DefaultBreakDuration == 0 {
		s.DefaultBreakDuration = d.DefaultBreakDuration
	}

	if s.DefaultPomodoroDuration == 0 {
		s.DefaultPomodoroDuration = d.DefaultPomodoroDuration
	}

	if len(s.DefaultTags) == 0 {
		s.DefaultTags = d.DefaultTags
	}
}

// UnmarshalText updates settings by parsing each key/value pair in logfmt.
func (s *Settings) UnmarshalText(b []byte) error {
	b = bytes.ReplaceAll(b, charNewline, charSpace)

	dec := logfmt.NewDecoder(bytes.NewReader(b))
	for dec.ScanRecord() {
		for dec.ScanKeyval() {
			key := string(dec.Key())
			val := string(dec.Value())

			switch key {
			case "daily_goal":
				goal, err := strconv.Atoi(val)
				if err != nil {
					return err
				}
				s.DailyGoal = goal
			case "default_break_duration":
				dur, err := strconv.Atoi(val)
				if err != nil {
					return err
				}
				s.DefaultBreakDuration = time.Duration(dur) * time.Minute
			case "default_pomodoro_duration":
				dur, err := strconv.Atoi(val)
				if err != nil {
					return err
				}
				s.DefaultPomodoroDuration = time.Duration(dur) * time.Minute
			case "default_tags":
				if val != "" {
					s.DefaultTags = strings.Split(val, ",")
				}
			}
		}
	}
	if err := dec.Err(); err != nil {
		return err
	}

	return nil
}
