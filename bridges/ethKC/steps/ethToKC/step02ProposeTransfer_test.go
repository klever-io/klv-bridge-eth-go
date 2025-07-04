package ethtokc

import (
	"context"
	"testing"

	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/stretchr/testify/assert"
)

func TestExecuteProposeTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return nil
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := bridgeCore.StepIdentifier(GettingPendingBatchFromEthereum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on WasTransferProposedOnKC", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return testBatch
		}
		bridgeStub.WasTransferProposedOnKCCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := bridgeCore.StepIdentifier(GettingPendingBatchFromEthereum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("not leader", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return testBatch
		}
		bridgeStub.WasTransferProposedOnKCCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return false
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on ProposeTransferOnKC", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return testBatch
		}
		bridgeStub.WasTransferProposedOnKCCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.ProposeTransferOnKCCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := bridgeCore.StepIdentifier(GettingPendingBatchFromEthereum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work - transfer already proposed", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return testBatch
		}
		bridgeStub.WasTransferProposedOnKCCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := bridgeCore.StepIdentifier(SigningProposedTransferOnKC)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return testBatch
		}
		bridgeStub.WasTransferProposedOnKCCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.ProposeTransferOnKCCalled = func(ctx context.Context) error {
			return nil
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}
		// Test IsInterfaceNil
		assert.NotNil(t, step.IsInterfaceNil())

		expectedStepIdentifier := bridgeCore.StepIdentifier(SigningProposedTransferOnKC)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}
