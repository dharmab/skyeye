package simpleradio

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

// The bot waits for an incoming transmission to finish before talking, so it does not step on a
// player mid-sentence.
func TestWaitForClearChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns immediately when nobody is talking", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			c := newTestClient(testGracePeriod, testRadio, testOtherRadio)

			start := time.Now()
			c.waitForClearChannel()

			assert.Equal(t, start, time.Now(), "no reason to wait")
		})
	})

	t.Run("waits past the deadline of an in-progress transmission", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			c := newTestClient(testGracePeriod, testRadio)
			c.receivers[testRadio].receive(testPacket(testGUID("alpha"), 1))

			start := time.Now()
			c.waitForClearChannel()

			// The grace period must elapse before the channel is clear, plus the courtesy pause.
			assert.GreaterOrEqual(t, time.Since(start), testGracePeriod)
		})
	})

	t.Run("waits for the busiest channel", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			c := newTestClient(testGracePeriod, testRadio, testOtherRadio)

			c.receivers[testRadio].receive(testPacket(testGUID("alpha"), 1))
			time.Sleep(200 * time.Millisecond)
			c.receivers[testOtherRadio].receive(
				testPacket(testGUID("bravo"), 1, testFrequency(testOtherRadio)),
			)
			latest := c.receivers[testOtherRadio].deadlineAt()

			c.waitForClearChannel()

			assert.False(t, time.Now().Before(latest), "must outlast every channel's deadline")
			for _, receiver := range c.receivers {
				assert.False(t, receiver.isReceivingTransmission())
			}
		})
	})
}

func TestReceiverDeadlineAt(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		r := newReceiver(testGracePeriod)
		assert.True(t, r.deadlineAt().IsZero(), "no transmission, no deadline")

		r.receive(testPacket(testGUID("alpha"), 1))
		assert.Equal(t, time.Now().Add(testGracePeriod), r.deadlineAt())

		time.Sleep(testGracePeriod + time.Nanosecond)
		packets, _, _ := r.completedTransmission()
		assert.Len(t, packets, 1)
		assert.True(t, r.deadlineAt().IsZero(), "taking the transmission clears the deadline")
	})
}
