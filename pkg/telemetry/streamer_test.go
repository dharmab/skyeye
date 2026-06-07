package telemetry

import (
	"bufio"
	"strings"
	"testing"
	"time"

	"github.com/dharmab/goacmi/v2/parsing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleLineAcceptsZeroReferenceTime verifies that a zero-valued
// ReferenceTime (0001-01-01T00:00:00Z) is accepted and does not cause
// subsequent time frames to be rejected. This is regression test for
// a case that sometimes happens on Lima Kilo.
func TestHandleLineAcceptsZeroReferenceTime(t *testing.T) {
	t.Parallel()

	const timeFrame = "#62754967308.46" // Observed value from eu.limakilo.net sometimes
	lines := []string{
		"FileType=text/acmi/tacview",
		"FileVersion=2.2",
		"0,ReferenceTime=0001-01-01T00:00:00Z",
		"0,ReferenceLongitude=30",
		"0,ReferenceLatitude=30",
		timeFrame,
	}

	client := newStreamingClient(time.Second)
	reader := bufio.NewReader(strings.NewReader(strings.Join(lines, "\n") + "\n"))
	for range lines {
		line, err := readACMILine(reader)
		require.NoError(t, err)
		require.NoError(t, client.handleLine(line))
	}

	offset, err := parsing.ParseTimeFrame(timeFrame)
	require.NoError(t, err)
	assert.Equal(t, time.Time{}.Add(offset), client.Time())
}

func TestReadACMILineContinuationPreservesNextRecord(t *testing.T) {
	t.Parallel()

	reader := bufio.NewReader(strings.NewReader(strings.Join([]string{
		"0,Comments=DCS Retribution Turn 6\\",
		"====================\\",
		"Most briefing information can be found on your kneeboard.",
		"0,ReferenceTime=1989-09-13T10:08:31Z",
	}, "\n") + "\n"))

	first, err := readACMILine(reader)
	require.NoError(t, err)
	assert.Equal(
		t,
		"0,Comments=DCS Retribution Turn 6====================\\\nMost briefing information can be found on your kneeboard.\n",
		first,
	)

	second, err := readACMILine(reader)
	require.NoError(t, err)
	assert.Equal(t, "0,ReferenceTime=1989-09-13T10:08:31Z\n", second)
}

func TestHandleLineAcceptsContinuationRecordBeforeNextObject(t *testing.T) {
	t.Parallel()

	client := newStreamingClient(time.Second)
	reader := bufio.NewReader(strings.NewReader(strings.Join([]string{
		"0,Comments=DCS Retribution Turn 6\\",
		"====================\\",
		"Most briefing information can be found on your kneeboard.",
		"0,ReferenceTime=1989-09-13T10:08:31Z",
	}, "\n") + "\n"))

	line, err := readACMILine(reader)
	require.NoError(t, err)
	require.NoError(t, client.handleLine(line))

	line, err = readACMILine(reader)
	require.NoError(t, err)
	require.NoError(t, client.handleLine(line))

	assert.Equal(t, time.Date(1989, 9, 13, 10, 8, 31, 0, time.UTC), client.Time())
}
