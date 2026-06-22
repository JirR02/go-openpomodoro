package openpomodoro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-logfmt/logfmt"
)

const (
	// TimeFormat is the format we generate and expect to parse timestamps in.
	TimeFormat = time.RFC3339
)

var (
	charNewline = []byte("\n")
	charSpace   = []byte(" ")
	timeFunc    = time.Now
)

// Pomodoro holds a single Pomodoro and related information.
type Pomodoro struct {
	// StartTime is the time that the Pomodoro started.
	StartTime time.Time `json:"start_time"`

	// Description is a description of the Pomodoro.
	Description string `json:"description"`

	// Duration is the length of the Pomodoro.
	Duration time.Duration `json:"-"`
	// JSONDuration is a placeholder for MarshalJSON to convert and store the
	// duration in the format XXmXXs.
	JSONDuration string `json:"duration"`

	// PauseTime is the time the Pomodoro paused.
	PauseTime time.Time `json:"-"`
	// JSONPauseTime is a placeholder for MarshalJSON to convert and store the
	// time as a string.
	JSONPauseTime string `json:"pause_time,omitempty"`

	// PauseDuration is the length of the Pause.
	PauseDuration time.Duration `json:"-"`
	// JSONDuration is a placeholder for MarshalJSON to convert and store the
	// duration in the format XXmXXs.
	JSONPauseDuration string `json:"pause_duration,omitempty"`

	// IsBreak is the status if a Pomodoro is a break timer
	IsBreak bool `json:"is_break,omitempty"`

	// Tags are the list of tags for this Pomodoro.
	Tags []string `json:"tags"`
}

// NewPomodoro returns a Pomodoro with defaults set.
func NewPomodoro() *Pomodoro {
	return &Pomodoro{
		Duration: DefaultSettings.DefaultPomodoroDuration,
	}
}

// EmptyPomodoro returns an empty Pomodoro.
func EmptyPomodoro() *Pomodoro {
	return &Pomodoro{}
}

// String return a string representation of the Pomodoro.
func (p Pomodoro) String() string {
	b, _ := p.MarshalText()
	return string(b)
}

// Matches returns whether or not another Pomodoro has the same StartTime.
func (p Pomodoro) Matches(o *Pomodoro) bool {
	delta := p.StartTime.Sub(o.StartTime)
	return delta >= -time.Second && delta <= time.Second
}

// MarshalJSON implements json.Marshaler.
func (p Pomodoro) MarshalJSON() ([]byte, error) {
	// This is required so that json.Marshal ignores that we also implement
	// encoding.TextMarshaler via MarshalText.
	type alias Pomodoro
	p.JSONDuration = p.Duration.String()

	if p.PauseDuration > 0 {
		p.JSONPauseDuration = p.PauseDuration.String()
	} else {
		p.JSONPauseDuration = ""
	}

	if !p.PauseTime.IsZero() {
		p.JSONPauseTime = p.PauseTime.Format(TimeFormat)
	} else {
		p.JSONPauseTime = ""
	}

	return json.Marshal((alias)(p))
}

// MarshalText marshals the Pomodoro's start time and attributes into a text
// string.
func (p Pomodoro) MarshalText() ([]byte, error) {
	timestamp := []byte(p.StartTime.Format(TimeFormat))

	if !p.PauseTime.IsZero() {
		p.JSONPauseTime = p.PauseTime.Format(TimeFormat)
	} else {
		p.JSONPauseTime = ""
	}

	var buf bytes.Buffer
	enc := logfmt.NewEncoder(&buf)

	if p.Description != "" {
		if err := enc.EncodeKeyval("description", p.Description); err != nil {
			return nil, err
		}
	}

	if err := enc.EncodeKeyval("duration", p.Duration.String()); err != nil {
		return nil, err
	}

	if p.JSONPauseTime != "" {
		if err := enc.EncodeKeyval("pause_time", p.JSONPauseTime); err != nil {
			return nil, err
		}
	}

	if p.PauseDuration > 0 {
		if err := enc.EncodeKeyval("pause_duration", p.PauseDuration.String()); err != nil {
			return nil, err
		}
	}

	if p.IsBreak {
		if err := enc.EncodeKeyval("is_break", strconv.FormatBool(p.IsBreak)); err != nil {
			return nil, err
		}
	}

	if len(p.Tags) > 0 {
		if err := enc.EncodeKeyval("tags", strings.Join(p.Tags, ",")); err != nil {
			return nil, err
		}
	}

	if err := enc.EndRecord(); err != nil {
		return nil, err
	}

	attributes := bytes.TrimRight(buf.Bytes(), "\n")

	return bytes.Join([][]byte{timestamp, attributes}, charSpace), nil
}

// UnmarshalText updates a Pomodoro's timestamp and attributes from a byte
// string.
func (p *Pomodoro) UnmarshalText(b []byte) error {
	b = bytes.TrimSpace(b)
	parts := bytes.SplitN(b, charSpace, 2)

	var timestamp []byte
	var attributes []byte

	switch len(parts) {
	case 0:
		return nil
	case 1:
		timestamp = parts[0]
	case 2:
		if parts[0] == nil {
			return nil
		}
		timestamp = parts[0]
		attributes = parts[1]
	default:
		return nil
	}

	if bytesAllWhitespace(timestamp) {
		return nil
	}

	startTime, err := time.Parse(TimeFormat, string(timestamp))
	if err != nil {
		return err
	}

	p.StartTime = startTime

	if len(attributes) > 0 {
		dec := logfmt.NewDecoder(bytes.NewReader(attributes))
		recordCount := 0
		for dec.ScanRecord() {
			recordCount++
			if recordCount > 1 {
				return fmt.Errorf("unexpected multiple lines in pomodoro attributes")
			}
			for dec.ScanKeyval() {
				key := string(dec.Key())
				val := string(dec.Value())

				switch key {
				case "description":
					p.Description = val
				case "duration":
					dur, err := time.ParseDuration(val)
					if err != nil {
						return err
					}
					p.Duration = dur
				case "pause_time":
					p.JSONPauseTime = val
				case "pause_duration":
					dur, err := time.ParseDuration(val)
					if err != nil {
						return err
					}
					p.PauseDuration = dur
				case "is_break":
					bre, err := strconv.ParseBool(val)
					if err != nil {
						return err
					}
					p.IsBreak = bre
				case "tags":
					if val != "" {
						p.Tags = strings.Split(val, ",")
					}
				}
			}
		}
		if err := dec.Err(); err != nil {
			return err
		}
	}

	if p.JSONPauseTime != "" {
		pauseTime, err := time.Parse(TimeFormat, p.JSONPauseTime)
		if err != nil {
			return err
		}
		p.PauseTime = pauseTime
		p.JSONPauseTime = ""
	}

	return nil
}

// ApplySettings sets the Pomodoro's defaults from settings if they are
// considered to be missing.
func (p *Pomodoro) ApplySettings(s *Settings) {
	if p.IsBreak {
		if p.Duration == 0 {
			p.Duration = s.DefaultBreakDuration
		}
	} else {
		if p.Duration == 0 {
			p.Duration = s.DefaultPomodoroDuration
		}

		if len(p.Tags) == 0 {
			p.Tags = s.DefaultTags
		}
	}
}

// DurationSeconds returns the Pomodoro's duration in seconds.
func (p *Pomodoro) DurationSeconds() int {
	return int(p.Duration.Round(time.Second).Seconds())
}

// EndTime returns the time the Pomodoro would end.
func (p *Pomodoro) EndTime() time.Time {
	totalDuration := p.Duration + p.PauseDuration
	return p.StartTime.Add(totalDuration)
}

// IsActive returns whether or not a Pomodoro is active.
func (p *Pomodoro) IsActive() bool {
	return !p.IsInactive() && !p.IsPaused() && !p.IsDone()
}

// IsDone returns whether or not a Pomodoro was active and is now done.
func (p *Pomodoro) IsDone() bool {
	if p.IsInactive() || p.IsPaused() {
		return false
	}
	return !timeFunc().Before(p.EndTime())
}

// IsInactive returns whether or not a Pomodoro is empty/not set/etc.
func (p *Pomodoro) IsInactive() bool {
	return p.StartTime.IsZero()
}

// IsPaused returns whether or not a Pomodoro is paused.
func (p *Pomodoro) IsPaused() bool {
	return !p.PauseTime.IsZero()
}

// Remaining returns the remaining duration of the Pomodoro.
func (p *Pomodoro) Remaining() time.Duration {
	if p.IsInactive() {
		return time.Duration(0)
	}

	if p.IsPaused() {
		return p.EndTime().Sub(p.PauseTime)
	}

	return p.EndTime().Sub(timeFunc())
}

// RemainingSeconds returns the remaining duration of the Pomodoro in seconds.
// Partial seconds are rounded up and down normally
func (p *Pomodoro) RemainingSeconds() int {
	return int(p.Remaining().Round(time.Second).Seconds())
}

func bytesAllWhitespace(b []byte) bool {
	return len(bytes.TrimSpace(b)) == 0
}
