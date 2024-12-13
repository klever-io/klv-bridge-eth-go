package timer

import (
	"testing"
	"time"

	"github.com/klever-io/klv-bridge-eth-go/core/timer/mock"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNTPTimer(t *testing.T) {
	t.Parallel()

	timer := NewNTPTimer()

	require.False(t, check.IfNil(timer))

	_ = timer.Close()
}

func TestNtpTimer_Close(t *testing.T) {
	t.Parallel()

	wasCalled := false
	ntpSyncer := &mock.SyncTimerStub{
		CloseCalled: func() error {
			wasCalled = true
			return nil
		},
	}

	timer := newNTPTimerWithInnerSyncTimer(ntpSyncer)

	err := timer.Close()
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestNtpTimer_Start(t *testing.T) {
	t.Parallel()

	wasCalled := false
	ntpSyncer := &mock.SyncTimerStub{
		StartSyncingTimeCalled: func() {
			wasCalled = true
		},
	}

	timer := newNTPTimerWithInnerSyncTimer(ntpSyncer)

	timer.Start()
	assert.True(t, wasCalled)
}

func TestNtpTimer_NowUnix(t *testing.T) {
	t.Parallel()

	timeValue := time.Unix(16438253, 0)
	ntpSyncer := &mock.SyncTimerStub{
		CurrentTimeCalled: func() time.Time {
			return timeValue
		},
	}

	timer := newNTPTimerWithInnerSyncTimer(ntpSyncer)

	unixTime := timer.NowUnix()
	assert.Equal(t, timeValue.Unix(), unixTime)
}
