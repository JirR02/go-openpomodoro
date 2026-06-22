package openpomodoro

import (
	"encoding"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PomodoroInterfaces(t *testing.T) {
	var _ encoding.TextMarshaler = Pomodoro{}
	var _ encoding.TextUnmarshaler = &Pomodoro{}
	var _ json.Marshaler = Pomodoro{}
}

func TestPomodoro_MarshalJSON(t *testing.T) {
	p := &Pomodoro{
		StartTime:   time.Date(2016, 06, 14, 12, 0, 0, 0, time.UTC),
		Duration:    25 * time.Minute,
		Tags:        []string{"a", "b"},
		Description: "A description",
	}
	b, err := p.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t,
		`{"start_time":"2016-06-14T12:00:00Z","description":"A description","duration":"25m0s","tags":["a","b"]}`,
		string(b))
}

func TestPomodoro_MarshalJSON_WithPause(t *testing.T) {
	p := &Pomodoro{
		StartTime:     time.Date(2016, 06, 14, 12, 0, 0, 0, time.UTC),
		Duration:      25 * time.Minute,
		PauseTime:     time.Date(2016, 06, 14, 12, 10, 0, 0, time.UTC),
		PauseDuration: 5 * time.Minute,
		Tags:          []string{"a", "b"},
		Description:   "A description",
	}
	b, err := p.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t,
		`{"start_time":"2016-06-14T12:00:00Z","description":"A description","duration":"25m0s","pause_time":"2016-06-14T12:10:00Z","pause_duration":"5m0s","tags":["a","b"]}`,
		string(b))
}

func TestPomodoro_MarshalText(t *testing.T) {
	p := &Pomodoro{
		StartTime:   time.Date(2016, 06, 14, 12, 0, 0, 0, time.UTC),
		Duration:    25 * time.Minute,
		Tags:        []string{"a", "b"},
		Description: "A description",
	}
	b, err := p.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t,
		"2016-06-14T12:00:00Z description=\"A description\" duration=25m0s tags=a,b",
		string(b))
}

func TestPomodoro_MarshalText_WithPause(t *testing.T) {
	p := &Pomodoro{
		StartTime:     time.Date(2016, 06, 14, 12, 0, 0, 0, time.UTC),
		Duration:      25 * time.Minute,
		PauseTime:     time.Date(2016, 06, 14, 12, 10, 0, 0, time.UTC),
		PauseDuration: 5 * time.Minute,
		Tags:          []string{"a", "b"},
		Description:   "A description",
	}
	b, err := p.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t,
		"2016-06-14T12:00:00Z description=\"A description\" duration=25m0s pause_time=2016-06-14T12:10:00Z pause_duration=5m0s tags=a,b",
		string(b))
}

func Test_Matches(t *testing.T) {
	timestamp, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)

	a := &Pomodoro{StartTime: timestamp}

	b := &Pomodoro{StartTime: timestamp}
	assert.True(t, a.Matches(b))

	b = &Pomodoro{StartTime: timestamp.Add(time.Minute)}
	assert.False(t, a.Matches(b))

	b = &Pomodoro{StartTime: timestamp.Add(500 * time.Millisecond)}
	assert.True(t, a.Matches(b))

	b = &Pomodoro{StartTime: timestamp.Add(-500 * time.Millisecond)}
	assert.True(t, a.Matches(b))
}

func Test_MarshalText(t *testing.T) {
	timestamp, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)

	var p *Pomodoro
	var actual []byte
	var expected string

	p = &Pomodoro{
		StartTime: timestamp,
		Duration:  25 * time.Minute,
	}
	expected = `2026-06-14T12:34:56-04:00 duration=25m0s`
	actual, err = p.MarshalText()
	require.Nil(t, err)
	assert.Equal(t, expected, string(actual))

	p = &Pomodoro{
		StartTime:   timestamp,
		Duration:    25 * time.Minute,
		Description: "working on stuff",
		Tags:        []string{"work", "stuff"},
	}
	expected = `2026-06-14T12:34:56-04:00 description="working on stuff" duration=25m0s tags=work,stuff`
	actual, err = p.MarshalText()
	require.Nil(t, err)
	assert.Equal(t, expected, string(actual))

	pauseTimestamp, err := time.Parse(TimeFormat, "2026-06-14T12:44:56-04:00")
	require.Nil(t, err)

	p = &Pomodoro{
		StartTime:     timestamp,
		Duration:      25 * time.Minute,
		PauseTime:     pauseTimestamp,
		PauseDuration: 5 * time.Minute,
		Description:   "working on stuff",
		Tags:          []string{"work", "stuff"},
	}
	expected = `2026-06-14T12:34:56-04:00 description="working on stuff" duration=25m0s pause_time=2026-06-14T12:44:56-04:00 pause_duration=5m0s tags=work,stuff`
	actual, err = p.MarshalText()
	require.Nil(t, err)
	assert.Equal(t, expected, string(actual))

	p = &Pomodoro{
		StartTime: timestamp,
		Duration:  10 * time.Minute,
		IsBreak:   true,
	}
	expected = `2026-06-14T12:34:56-04:00 duration=10m0s is_break=true`
	actual, err = p.MarshalText()
	require.Nil(t, err)
	assert.Equal(t, expected, string(actual))
}

func Test_UnmarshalText_timeOnly(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(`2026-06-14T12:34:56-04:00`))
	require.Nil(t, err)

	startTime, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)
	expected := &Pomodoro{StartTime: startTime}

	assert.Equal(t, expected, p)
}

func Test_UnmarshalText_timeOnlyWithNewline(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(`2026-06-14T12:34:56-04:00
`))
	require.Nil(t, err)

	startTime, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)
	expected := &Pomodoro{StartTime: startTime}

	assert.Equal(t, expected, p)
}

func Test_UnmarshalText_singleLineAttributes(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(`2026-06-14T12:34:56-04:00 description="working on stuff" duration=25m0s tags=work,stuff`))
	require.Nil(t, err)

	startTime, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)
	expected := &Pomodoro{
		StartTime:   startTime,
		Description: "working on stuff",
		Duration:    25 * time.Minute,
		Tags:        []string{"work", "stuff"},
	}

	assert.Equal(t, expected, p)
}

func Test_UnmarshalText_PauseAttributes(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(`2026-06-14T12:34:56-04:00 description="working on stuff" duration=25m0s pause_time=2026-06-14T12:44:56-04:00 pause_duration=5m0s tags=work,stuff`))
	require.Nil(t, err)

	startTime, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)

	pauseTime, err := time.Parse(TimeFormat, "2026-06-14T12:44:56-04:00")
	require.Nil(t, err)

	expected := &Pomodoro{
		StartTime:     startTime,
		Description:   "working on stuff",
		Duration:      25 * time.Minute,
		PauseTime:     pauseTime,
		PauseDuration: 5 * time.Minute,
		Tags:          []string{"work", "stuff"},
	}

	assert.Equal(t, expected, p)
}

func Test_UnmarshalText_empty(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(``))
	require.Nil(t, err)
	assert.True(t, p.IsInactive())
}

func Test_UnmarshalText_whitespace(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(" \n "))
	require.Nil(t, err)
	assert.True(t, p.IsInactive())
}

func Test_UnmarshalText_multipleEntries(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(`2026-06-14T12:34:56-04:00 description="working on stuff" duration=25m0s tags=work,stuff
2026-06-14T12:34:56-04:00 description="working on stuff" duration=25m0s tags=work,stuff`))
	assert.Error(t, err)
}

func Test_ApplySettings_empty(t *testing.T) {
	p := &Pomodoro{}

	s := &Settings{
		DefaultPomodoroDuration: 25 * time.Minute,
		DefaultTags:             []string{"work"},
	}

	p.ApplySettings(s)

	assert.Equal(t, p.Duration, 25*time.Minute)
	assert.Equal(t, p.Tags, []string{"work"})
}

func Test_ApplySettings_existing(t *testing.T) {
	p := &Pomodoro{
		Duration: 30 * time.Minute,
		Tags:     []string{"play"},
	}

	s := &Settings{
		DefaultPomodoroDuration: 25 * time.Minute,
		DefaultTags:             []string{"work"},
	}

	p.ApplySettings(s)

	assert.Equal(t, p.Duration, 30*time.Minute)
	assert.Equal(t, p.Tags, []string{"play"})
}

func TestPomodoro_MarshalJSON_Break(t *testing.T) {
	p := &Pomodoro{
		StartTime:   time.Date(2016, 06, 14, 12, 0, 0, 0, time.UTC),
		Duration:    10 * time.Minute,
		IsBreak:     true,
		Tags:        []string{"rest"},
		Description: "Coffee break",
	}
	b, err := p.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t,
		`{"start_time":"2016-06-14T12:00:00Z","description":"Coffee break","duration":"10m0s","is_break":true,"tags":["rest"]}`,
		string(b))
}

func TestPomodoro_MarshalText_Break(t *testing.T) {
	p := &Pomodoro{
		StartTime:   time.Date(2016, 06, 14, 12, 0, 0, 0, time.UTC),
		Duration:    10 * time.Minute,
		IsBreak:     true,
		Tags:        []string{"rest"},
		Description: "Coffee break",
	}
	b, err := p.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t,
		"2016-06-14T12:00:00Z description=\"Coffee break\" duration=10m0s is_break=true tags=rest",
		string(b))
}

func Test_UnmarshalText_Break(t *testing.T) {
	p := &Pomodoro{}
	err := p.UnmarshalText([]byte(`2026-06-14T12:34:56-04:00 description="Coffee break" duration=10m0s is_break=true tags=rest`))
	require.Nil(t, err)

	startTime, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)

	expected := &Pomodoro{
		StartTime:   startTime,
		Description: "Coffee break",
		Duration:    10 * time.Minute,
		IsBreak:     true,
		Tags:        []string{"rest"},
	}

	assert.Equal(t, expected, p)
}

func Test_DurationSeconds(t *testing.T) {
	p := Pomodoro{}

	p.Duration = 30 * time.Minute
	assert.Equal(t, 1800, p.DurationSeconds())

	p.Duration = 29*time.Minute + 30*time.Second
	assert.Equal(t, 1770, p.DurationSeconds())

	p.Duration = 29*time.Minute + 29*time.Second
	assert.Equal(t, 1769, p.DurationSeconds())
}

func Test_EndTime(t *testing.T) {
	start, err := time.Parse(TimeFormat, "2026-06-14T12:34:56-04:00")
	require.Nil(t, err)
	expected, err := time.Parse(TimeFormat, "2026-06-14T12:59:56-04:00")
	require.Nil(t, err)

	p := Pomodoro{StartTime: start, Duration: 25 * time.Minute}
	assert.Equal(t, expected, p.EndTime())

	expectedWithPause, err := time.Parse(TimeFormat, "2026-06-14T13:04:56-04:00")
	require.Nil(t, err)

	pWithPause := Pomodoro{
		StartTime:     start,
		Duration:      25 * time.Minute,
		PauseDuration: 5 * time.Minute,
	}
	assert.Equal(t, expectedWithPause, pWithPause.EndTime())
}

func Test_IsActive(t *testing.T) {
	timeFunc = time.Now

	p := NewPomodoro()
	p.Duration = 25 * time.Minute

	cases := map[time.Duration]bool{
		24 * time.Minute: true,
		25 * time.Minute: false,
		26 * time.Minute: false,
		time.Hour:        false,
		-time.Hour:       true,
		0 * time.Second:  true,
	}

	for duration, expected := range cases {
		p.StartTime = timeFunc().Add(-duration)
		assert.Equal(t, expected, p.IsActive(), duration.String())
	}
	p.StartTime = timeFunc().Add(-10 * time.Minute)
	p.PauseTime = timeFunc()
	assert.False(t, p.IsActive())

	p.PauseTime = time.Time{}
}

func Test_IsDone(t *testing.T) {
	timeFunc = time.Now

	p := NewPomodoro()
	p.Duration = 25 * time.Minute

	cases := map[time.Duration]bool{
		24 * time.Minute: false,
		25 * time.Minute: true,
		26 * time.Minute: true,
		time.Hour:        true,
		-time.Hour:       false,
		0 * time.Second:  false,
	}

	for duration, expected := range cases {
		p.StartTime = timeFunc().Add(-duration)
		assert.Equal(t, expected, p.IsDone(), duration.String())
	}
	p.StartTime = timeFunc().Add(-30 * time.Minute)
	p.PauseTime = timeFunc()
	assert.False(t, p.IsDone())

	p.PauseTime = time.Time{}
}

func Test_isPaused(t *testing.T) {
	p := NewPomodoro()

	assert.False(t, p.IsPaused())

	p.PauseTime = time.Now()
	assert.True(t, p.IsPaused())

	p.PauseTime = time.Time{}
	assert.False(t, p.IsPaused())
}

func Test_IsInactive_true(t *testing.T) {
	assert.True(t, EmptyPomodoro().IsInactive())
}

func Test_IsInactive_false(t *testing.T) {
	timestamp, err := time.Parse(
		TimeFormat,
		"2026-06-14T12:34:56-04:00",
	)
	require.Nil(t, err)
	p := Pomodoro{StartTime: timestamp}

	assert.False(t, p.IsInactive())
}

func Test_Remaining(t *testing.T) {
	timeFunc = time.Now

	p := NewPomodoro()
	p.Duration = 25 * time.Minute

	assert.Equal(t, float64(0), p.Remaining().Seconds())

	cases := map[time.Duration]time.Duration{
		0 * time.Minute:  25 * time.Minute,
		1 * time.Minute:  24 * time.Minute,
		24 * time.Minute: 1 * time.Minute,
		25 * time.Minute: 0 * time.Minute,
		26 * time.Minute: -1 * time.Minute,
	}

	for duration, expected := range cases {
		p.StartTime = timeFunc().Add(-duration)
		assert.InDelta(t, expected.Seconds(), p.Remaining().Seconds(), 1)
	}

	p.StartTime = timeFunc().Add(-10 * time.Minute)
	p.PauseTime = timeFunc()
	assert.InDelta(t, (15 * time.Minute).Seconds(), p.Remaining().Seconds(), 1)

	oldTimeFunc := timeFunc
	timeFunc = func() time.Time { return oldTimeFunc().Add(5 * time.Minute) }

	assert.InDelta(t, (15 * time.Minute).Seconds(), p.Remaining().Seconds(), 1)

	timeFunc = oldTimeFunc
	p.PauseTime = time.Time{}
}

func Test_RemainingSeconds(t *testing.T) {
	p := NewPomodoro()
	p.Duration = 25 * time.Minute

	assert.Equal(t, 0, p.RemainingSeconds())

	cases := map[time.Duration]int{
		0 * time.Minute:  1500,
		1 * time.Minute:  1440,
		24 * time.Minute: 60,
		25 * time.Minute: 0,
		26 * time.Minute: -60,

		29 * time.Second:                1471,
		30 * time.Second:                1470,
		24*time.Minute + 29*time.Second: 31,
		24*time.Minute + 30*time.Second: 30,
	}

	for duration, expected := range cases {
		p.StartTime = timeFunc().Add(-duration)
		assert.Equal(t, expected, p.RemainingSeconds())
	}

	p.StartTime = timeFunc().Add(-(10*time.Minute + 29*time.Second))
	p.PauseTime = timeFunc()

	assert.Equal(t, 871, p.RemainingSeconds())

	p.PauseTime = time.Time{}
}
