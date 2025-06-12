package kctoeth

import (
	"context"
	"testing"

	ethKc "github.com/klever-io/klv-bridge-eth-go/bridges/ethKc"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

var actionID = uint64(662528)

func TestExecute_SignProposedSetStatus(t *testing.T) {
	t.Parallel()
	t.Run("nil batch on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return nil
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("error on GetAndStoreActionIDForProposeSetStatusFromKc", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.GetAndStoreActionIDForProposeSetStatusFromKcCalled = func(ctx context.Context) (uint64, error) {
			return ethKc.InvalidActionID, expectedError
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("invalid actionID on GetAndStoreActionIDForProposeSetStatusFromKc", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.GetAndStoreActionIDForProposeSetStatusFromKcCalled = func(ctx context.Context) (uint64, error) {
			return ethKc.InvalidActionID, nil
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("error on WasActionSignedOnKc", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.WasActionSignedOnKcCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("error on SignActionOnKc", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.SignActionOnKcCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if proposed set status was signed, go to WaitingForQuorumOnSetStatus", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorSignProposedSetStatus()
			bridgeStub.WasActionSignedOnKcCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			wasCalled := false
			bridgeStub.SignActionOnKcCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}

			step := signProposedSetStatusStep{
				bridge: bridgeStub,
			}

			expectedStep := bridgeCore.StepIdentifier(WaitingForQuorumOnSetStatus)
			stepIdentifier := step.Execute(context.Background())
			assert.False(t, wasCalled)
			assert.Equal(t, expectedStep, stepIdentifier)
		})
		t.Run("if proposed set status was not signed, sign and go to WaitingForQuorumOnSetStatus", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorSignProposedSetStatus()
			wasCalled := false
			bridgeStub.SignActionOnKcCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}

			step := signProposedSetStatusStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())
			expectedStep := bridgeCore.StepIdentifier(WaitingForQuorumOnSetStatus)
			stepIdentifier := step.Execute(context.Background())
			assert.True(t, wasCalled)
			assert.NotEqual(t, step.Identifier(), stepIdentifier)
			assert.Equal(t, expectedStep, stepIdentifier)
		})
	})

}

func createStubExecutorSignProposedSetStatus() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
		return testBatch
	}
	stub.GetAndStoreActionIDForProposeSetStatusFromKcCalled = func(ctx context.Context) (uint64, error) {
		return actionID, nil
	}
	stub.WasActionSignedOnKcCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	stub.SignActionOnKcCalled = func(ctx context.Context) error {
		return nil
	}
	return stub
}
