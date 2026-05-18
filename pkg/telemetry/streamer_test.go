package telemetry

import (
	"bufio"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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