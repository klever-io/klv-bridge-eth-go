package kleverchaintoeth

import (
	"context"
	"testing"

	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

var initialStep = bridgeCore.StepIdentifier(GettingPendingBatchFromKleverchain)

func TestExecute_SignProposedTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedTransfer()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
			return nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("nil batch on SignTransferOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedTransfer()
		bridgeStub.SignTransferOnEthereumCalled = func() error {
			return expectedError
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedTransfer()

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStepIdentifier := bridgeCore.StepIdentifier(WaitingForQuorumOnTransfer)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}

func createStubExecutorSignProposedTransfer() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
		return testBatch
	}
	stub.SignTransferOnEthereumCalled = func() error {
		return nil
	}
	return stub
}
