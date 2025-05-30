package kleverchaintoeth

import (
	"context"
	"testing"

	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

func TestExecute_PerformSetStatus(t *testing.T) {
	t.Parallel()

	t.Run("error on WasActionPerformedOnKleverchainCalled", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformSetStatus()
		bridgeStub.WasActionPerformedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := performSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on PerformActionOnKleverchainCalled", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformSetStatus()
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.PerformActionOnKleverchainCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := performSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if transfer was performed we should go to initial step", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			bridgeStub.WasActionPerformedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())
			stepIdentifier := step.Execute(context.Background())
			assert.Equal(t, initialStep, stepIdentifier)
		})
		t.Run("if not leader, wait in this step", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			wasCalled := false
			bridgeStub.PerformActionOnKleverchainCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}

			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			stepIdentifier := step.Execute(context.Background())
			assert.False(t, wasCalled)
			assert.Equal(t, step.Identifier(), stepIdentifier)
		})
		t.Run("if leader, first perform Set Status and then check again WasSetStatusPerformedOnKleverchain", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			bridgeStub.MyTurnAsLeaderCalled = func() bool {
				return true
			}
			wasCalled := false
			bridgeStub.PerformActionOnKleverchainCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}
			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			stepIdentifier := step.Execute(context.Background())
			assert.True(t, wasCalled)
			assert.Equal(t, step.Identifier(), stepIdentifier)
		})
	})
}

func createStubExecutorPerformSetStatus() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.WasActionPerformedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	stub.MyTurnAsLeaderCalled = func() bool {
		return false
	}
	return stub
}
