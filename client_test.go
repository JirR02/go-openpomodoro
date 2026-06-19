package openpomodoro

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/crufter/copyrecur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Case struct {
	Name    string
	Fixture string
	Steps   []Step
}

type Step func(*testing.T, *Client, string)

func Test_Client(t *testing.T) {
	cases := []Case{
		Case{
			Name: "typical pomodoro",
			Steps: []Step{
				assertActive(false), assertDone(false), assertInactive(true), assertIsPaused(false),

				start(&Pomodoro{}),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertMinutesRemaining(25),
				assertHistoryLength(1),

				timeTravel(1 * time.Minute),
				assertMinutesRemaining(24),

				pause(),
				assertMinutesRemaining(24),
				assertIsPaused(true),

				timeTravel(5 * time.Minute),
				assertMinutesRemaining(24),
				assertIsPaused(true),

				resume(),
				assertPauseDuration(5),
				assertMinutesRemaining(24),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),

				timeTravel(4 * time.Minute),
				pause(),
				assertMinutesRemaining(20),
				assertIsPaused(true),

				timeTravel(10 * time.Minute),
				assertMinutesRemaining(20),
				assertIsPaused(true),

				resume(),
				assertPauseDuration(15),
				assertMinutesRemaining(20),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),

				timeTravel(19 * time.Minute),
				assertMinutesRemaining(1),

				timeTravel(1 * time.Minute),
				assertMinutesRemaining(0),
				assertActive(false), assertDone(true), assertInactive(false), assertIsPaused(false),

				timeTravel(1 * time.Minute),
				assertMinutesRemaining(-1),
				assertActive(false), assertDone(true), assertInactive(false), assertIsPaused(false),

				clear(),
				assertActive(false), assertDone(false), assertInactive(true), assertIsPaused(false),
				assertMinutesRemaining(0),
			},
		},

		Case{
			Name: "finish early",
			Steps: []Step{
				start(&Pomodoro{}),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertMinutesRemaining(25),
				assertHistoryLength(1),

				timeTravel(15 * time.Minute),
				pause(),
				assertIsPaused(true),
				assertMinutesRemaining(10),

				timeTravel(5 * time.Minute),
				assertIsPaused(true),
				assertMinutesRemaining(10),

				end(),
				assertActive(false), assertDone(false), assertInactive(true), assertIsPaused(false),
				assertHistoryLength(1),
			},
		},

		Case{
			Name: "pause while paused",
			Steps: []Step{
				start(&Pomodoro{}),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertMinutesRemaining(25),
				assertHistoryLength(1),

				timeTravel(15 * time.Minute),
				pause(),
				assertIsPaused(true),
				assertMinutesRemaining(10),

				timeTravel(5 * time.Minute),
				pause(),
				assertIsPaused(true),
				assertMinutesRemaining(10),

				end(),
				assertActive(false), assertDone(false), assertInactive(true), assertIsPaused(false),
				assertHistoryLength(1),
			},
		},

		Case{
			Name: "resume while not paused",
			Steps: []Step{
				start(&Pomodoro{}),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertMinutesRemaining(25),
				assertHistoryLength(1),

				timeTravel(15 * time.Minute),
				resume(),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertMinutesRemaining(10),

				end(),
				assertActive(false), assertDone(false), assertInactive(true), assertIsPaused(false),
				assertHistoryLength(1),
			},
		},

		Case{
			Name: "restart during",
			Steps: []Step{
				start(&Pomodoro{}),
				timeTravel(15 * time.Minute),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertMinutesRemaining(10),
				assertHistoryLength(1),

				start(&Pomodoro{}),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertMinutesRemaining(25),
				assertHistoryLength(1),
			},
		},

		Case{
			Name: "restart after",
			Steps: []Step{
				start(&Pomodoro{}),
				timeTravel(30 * time.Minute),
				assertActive(false), assertDone(true), assertInactive(false), assertIsPaused(false),
				assertHistoryLength(1),

				start(&Pomodoro{}),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertHistoryLength(2),
			},
		},

		Case{
			Name: "pause resume active/done",
			Steps: []Step{
				pause(),
				assertActive(false), assertDone(false), assertInactive(true), assertIsPaused(false),

				resume(),
				assertActive(false), assertDone(false), assertInactive(true), assertIsPaused(false),

				start(&Pomodoro{}),
				timeTravel(30 * time.Minute),
				assertActive(false), assertDone(true), assertInactive(false), assertIsPaused(false),
				assertHistoryLength(1),

				pause(),
				assertActive(false), assertDone(true), assertInactive(false), assertIsPaused(false),

				resume(),
				assertActive(false), assertDone(true), assertInactive(false), assertIsPaused(false),

				start(&Pomodoro{}),
				assertActive(true), assertDone(false), assertInactive(false), assertIsPaused(false),
				assertHistoryLength(2),
			},
		},
	}

	for _, tc := range cases {
		timeFunc = time.Now

		c, err := NewClient(fixture(tc.Fixture))
		require.Nil(t, err)

		for _, s := range tc.Steps {
			s(t, c, tc.Name)
		}
	}
}

func start(p *Pomodoro) Step {
	return func(t *testing.T, c *Client, n string) {
		err := c.Start(p)
		require.Nil(t, err, n)
	}
}

func pause() Step {
	return func(t *testing.T, c *Client, n string) {
		err := c.Pause()
		require.Nil(t, err, n)
	}
}

func resume() Step {
	return func(t *testing.T, c *Client, n string) {
		err := c.Resume()
		require.Nil(t, err, n)
	}
}

func clear() Step {
	return func(t *testing.T, c *Client, n string) {
		require.Nil(t, c.Clear(), n)
	}
}

func end() Step {
	return func(t *testing.T, c *Client, n string) {
		require.Nil(t, c.Finish(), n)
	}
}

func assertActive(b bool) Step {
	return func(t *testing.T, c *Client, n string) {
		p, err := c.Pomodoro()
		require.Nil(t, err, n)
		assert.Equal(t, b, p.IsActive(), n)
	}
}

func assertDone(b bool) Step {
	return func(t *testing.T, c *Client, n string) {
		p, err := c.Pomodoro()
		require.Nil(t, err, n)
		assert.Equal(t, b, p.IsDone(), n)
	}
}

func assertDuration(d int) Step {
	return func(t *testing.T, c *Client, n string) {
		p, err := c.Pomodoro()
		require.Nil(t, err)
		assert.Equal(t, d, p.DurationMinutes(), n)
	}
}

func assertPauseDuration(d int) Step {
	return func(t *testing.T, c *Client, n string) {
		p, err := c.Pomodoro()
		require.Nil(t, err)
		assert.Equal(t, d, int(p.PauseDuration.Minutes()), n)
	}
}

func assertInactive(b bool) Step {
	return func(t *testing.T, c *Client, n string) {
		p, err := c.Pomodoro()
		require.Nil(t, err)
		assert.Equal(t, b, p.IsInactive(), n)
	}
}

func assertIsPaused(b bool) Step {
	return func(t *testing.T, c *Client, n string) {
		p, err := c.Pomodoro()
		require.Nil(t, err)
		assert.Equal(t, b, p.IsPaused(), n)
	}
}

func assertHistoryLength(i int) Step {
	return func(t *testing.T, c *Client, n string) {
		h, err := c.History()
		require.Nil(t, err)
		assert.Equal(t, i, len(h.Pomodoros), n)
	}
}
func assertMinutesRemaining(m int) Step {
	return func(t *testing.T, c *Client, n string) {
		p, err := c.Pomodoro()
		require.Nil(t, err)
		assert.Equal(t, m, round(p.Remaining().Minutes()), n)
	}
}

func timeTravel(d time.Duration) Step {
	return func(t *testing.T, c *Client, n string) {
		current := timeFunc()
		new := current.Add(d)
		timeFunc = func() time.Time { return new }
	}
}

func Test_Pomodoro_simple(t *testing.T) {
	c, err := NewClient(fixture("simple"))
	require.Nil(t, err)

	actual, err := c.Pomodoro()
	require.Nil(t, err)

	expectedStartTime, err := time.Parse(
		TimeFormat,
		"2026-06-14T12:34:56-04:00",
	)
	require.Nil(t, err)

	expected := NewPomodoro()
	expected.StartTime = expectedStartTime
	assert.Equal(t, expected, actual)
}

func Test_Pomodoro_emptyFiles(t *testing.T) {
	c, err := NewClient(fixture("empty"))
	require.Nil(t, err)

	actual, err := c.Pomodoro()
	require.Nil(t, err)

	assert.Equal(t, EmptyPomodoro(), actual)
}

func Test_Pomodoro_noFiles(t *testing.T) {
	c, err := NewClient(fixture("none"))
	require.Nil(t, err)

	actual, err := c.Pomodoro()
	require.Nil(t, err)

	assert.Equal(t, &Pomodoro{}, actual)
}

func Test_Pomodoro_fileInsteadOfDir(t *testing.T) {
	c, err := NewClient(filepath.Join(fixture("file"), "file"))
	require.Nil(t, err)

	_, err = c.Pomodoro()
	assert.NotNil(t, err)
}

func Test_Settings_defaults(t *testing.T) {
	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	s, err := c.Settings()
	require.Nil(t, err)

	assert.Equal(t, &DefaultSettings, s)
}

func Test_Settings_exists(t *testing.T) {
	c, err := NewClient(fixture("settings"))
	require.Nil(t, err)

	s, err := c.Settings()
	require.Nil(t, err)

	assert.Equal(t, 8, s.DailyGoal)
	assert.Equal(t, 10*time.Minute, s.DefaultBreakDuration)
	assert.Equal(t, 20*time.Minute, s.DefaultPomodoroDuration)
	assert.Equal(t, []string{"billable", "work"}, s.DefaultTags)
}

func Test_Start(t *testing.T) {
	timeFunc = fakeTime

	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	p := &Pomodoro{}
	err = c.Start(p)
	require.Nil(t, err)

	assert.Equal(t, p.StartTime, fakeTime())

	current, err := c.Pomodoro()
	require.Nil(t, err)

	assert.Equal(t, current.Duration, p.Duration)
	assert.Equal(t, current.StartTime.Second(), p.StartTime.Second())
}

func Test_Start_withOptions(t *testing.T) {
	timeFunc = fakeTime
	startTime := timeFunc().Add(-10 * time.Minute)

	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	p := &Pomodoro{
		Description: "description",
		Duration:    30 * time.Minute,
		Tags:        []string{"tag1", "tag2"},
		StartTime:   startTime,
	}

	require.Nil(t, c.Start(p))

	assert.Equal(t, p.StartTime, startTime)

	current, err := c.Pomodoro()
	require.Nil(t, err)

	assert.Equal(t, current.Description, "description")
	assert.Equal(t, current.Duration, 30*time.Minute)
	assert.Equal(t, current.Tags, []string{"tag1", "tag2"})
}

func Test_Finish_active(t *testing.T) {
	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	p := &Pomodoro{}

	err = c.Start(p)
	require.Nil(t, err)

	err = c.Finish()
	require.Nil(t, err)

	current, err := c.Pomodoro()
	require.Nil(t, err)
	assert.True(t, current.IsInactive())
}

func Test_Finish_inactive(t *testing.T) {
	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	err = c.Finish()
	require.Nil(t, err)

	current, err := c.Pomodoro()
	require.Nil(t, err)
	assert.True(t, current.IsInactive())
}

func Test_Cancel_active(t *testing.T) {
	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	p := &Pomodoro{}

	err = c.Start(p)
	require.Nil(t, err)

	err = c.Cancel()
	require.Nil(t, err)

	current, err := c.Pomodoro()
	require.Nil(t, err)
	assert.True(t, current.IsInactive())

	history, err := c.History()
	require.Nil(t, err)
	assert.Empty(t, history.Pomodoros)
}

func Test_Cancel_finished(t *testing.T) {
	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	p := &Pomodoro{}

	err = c.Start(p)
	require.Nil(t, err)

	err = c.Finish()
	require.Nil(t, err)

	err = c.Cancel()
	require.Nil(t, err)

	current, err := c.Pomodoro()
	require.Nil(t, err)
	assert.True(t, current.IsInactive())

	history, err := c.History()
	require.Nil(t, err)
	assert.NotEmpty(t, history.Pomodoros)
}

func Test_Cancel_inactive(t *testing.T) {
	c, err := NewClient(fixture(""))
	require.Nil(t, err)

	err = c.Cancel()
	require.Nil(t, err)
}

func fixture(f string) string {
	tmpDir, err := ioutil.TempDir("", f)
	if err != nil {
		log.Fatal(err)
	}
	tmpDir = filepath.Join(tmpDir, f)

	if f != "" {
		err := copyrecur.CopyDir(filepath.Join("fixtures", f), tmpDir)
		if err != nil {
			log.Fatal(err)
		}
	}

	return tmpDir
}

func fakeTime() time.Time {
	t, err := time.Parse(TimeFormat, "2016-06-14T12:34:56-04:00")
	if err != nil {
		panic(err)
	}
	return t
}
