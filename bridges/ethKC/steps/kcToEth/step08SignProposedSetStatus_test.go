package kctoeth

import (
	"context"
	"testing"

	ethKC "github.com/klever-io/klv-bridge-eth-go/bridges/ethKC"
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
	t.Run("error on GetAndStoreActionIDForProposeSetStatusFromKC", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.GetAndStoreActionIDForProposeSetStatusFromKCCalled = func(ctx context.Context) (uint64, error) {
			return ethKC.InvalidActionID, expectedError
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("invalid actionID on GetAndStoreActionIDForProposeSetStatusFromKC", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.GetAndStoreActionIDForProposeSetStatusFromKCCalled = func(ctx context.Context) (uint64, error) {
			return ethKC.InvalidActionID, nil
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("error on WasActionSignedOnKC", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.WasActionSignedOnKCCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := signProposedSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("error on SignActionOnKC", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedSetStatus()
		bridgeStub.SignActionOnKCCalled = func(ctx context.Context) error {
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
			bridgeStub.WasActionSignedOnKCCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			wasCalled := false
			bridgeStub.SignActionOnKCCalled = func(ctx context.Context) error {
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
			bridgeStub.SignActionOnKCCalled = func(ctx context.Context) error {
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
	stub.GetAndStoreActionIDForProposeSetStatusFromKCCalled = func(ctx context.Context) (uint64, error) {
		return actionID, nil
	}
	stub.WasActionSignedOnKCCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	stub.SignActionOnKCCalled = func(ctx context.Context) error {
		return nil
	}
	return stub
}
