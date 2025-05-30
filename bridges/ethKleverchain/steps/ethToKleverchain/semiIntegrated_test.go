package ethtokleverchain

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKleverchain/steps"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	getAndStoreBatchFromEthereum                       = "GetAndStoreBatchFromEthereum"
	getLastExecutedEthBatchIDFromKleverchain           = "GetLastExecutedEthBatchIDFromKleverchain"
	verifyLastDepositNonceExecutedOnEthereumBatch      = "VerifyLastDepositNonceExecutedOnEthereumBatch"
	wasTransferProposedOnKleverchain                   = "WasTransferProposedOnKleverchain"
	wasActionSignedOnKleverchain                       = "WasActionSignedOnKleverchain"
	signActionOnKleverchain                            = "SignActionOnKleverchain"
	getAndStoreActionIDForProposeTransferOnKleverchain = "GetAndStoreActionIDForProposeTransferOnKleverchain"
	ProcessMaxQuorumRetriesOnKleverchain               = "ProcessMaxQuorumRetriesOnKleverchain"
	resetRetriesCountOnKleverchain                     = "ResetRetriesCountOnKleverchain"
	processQuorumReachedOnKleverchain                  = "ProcessQuorumReachedOnKleverchain"
	wasActionPerformedOnKleverchain                    = "WasActionPerformedOnKleverchain"
	proposeTransferOnKleverchain                       = "ProposeTransferOnKleverchain"
	performActionOnKleverchain                         = "PerformActionOnKleverchain"
)

var trueHandler = func() bool { return true }
var falseHandler = func() bool { return false }

type errorHandler struct {
	lastError error
}

func (eh *errorHandler) storeAndReturnError(err error) error {
	eh.lastError = err
	return err
}

type argsBridgeStub struct {
	failingStep                      string
	myTurnHandler                    func() bool
	wasTransferProposedHandler       func() bool
	wasProposedTransferSignedHandler func() bool
	wasActionSigned                  func() bool
	isQuorumReachedHandler           func() bool
	wasActionIDPerformedHandler      func() bool
	maxRetriesReachedHandler         func() bool
	validateBatchHandler             func() bool
}

func createMockBridge(args argsBridgeStub) (*bridgeTests.BridgeExecutorStub, *errorHandler) {
	errHandler := &errorHandler{}
	stub := bridgeTests.NewBridgeExecutorStub()
	expectedErr := errors.New("expected error")
	stub.MyTurnAsLeaderCalled = func() bool {
		return args.myTurnHandler()
	}
	stub.GetAndStoreActionIDForProposeTransferOnKleverchainCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeTransferOnKleverchain {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 2, errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredActionIDCalled = func() uint64 {
		return 2
	}
	stub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
		if args.failingStep == getAndStoreBatchFromEthereum {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
		return &bridgeCore.TransferBatch{}
	}
	stub.GetLastExecutedEthBatchIDFromKleverchainCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getLastExecutedEthBatchIDFromKleverchain {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 3, errHandler.storeAndReturnError(nil)
	}
	stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled = func(ctx context.Context) error {
		if args.failingStep == verifyLastDepositNonceExecutedOnEthereumBatch {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasTransferProposedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasTransferProposedOnKleverchain {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasTransferProposedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeTransferOnKleverchainCalled = func(ctx context.Context) error {
		if args.failingStep == proposeTransferOnKleverchain {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnKleverchain {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSigned(), errHandler.storeAndReturnError(nil)
	}
	stub.SignActionOnKleverchainCalled = func(ctx context.Context) error {
		if args.failingStep == signActionOnKleverchain {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessQuorumReachedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == processQuorumReachedOnKleverchain {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.isQuorumReachedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnKleverchain {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionIDPerformedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnKleverchainCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnKleverchain {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnKleverchainCalled = func() bool {
		return args.maxRetriesReachedHandler()
	}

	return stub, errHandler
}

func createStateMachine(t *testing.T, executor steps.Executor, initialStep bridgeCore.StepIdentifier) *stateMachine.StateMachineMock {
	stepsSlice, err := CreateSteps(executor)
	require.Nil(t, err)

	sm := stateMachine.NewStateMachineMock(stepsSlice, initialStep)
	err = sm.Initialize()
	require.Nil(t, err)

	return sm
}

func TestHappyCaseWhenLeader(t *testing.T) {
	t.Parallel()

	args := argsBridgeStub{
		myTurnHandler:                    trueHandler,
		isQuorumReachedHandler:           trueHandler,
		wasActionIDPerformedHandler:      trueHandler,
		validateBatchHandler:             trueHandler,
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
		wasActionSigned:                  falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromEthereum)
	numSteps := 20
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnKleverchain))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnKleverchain))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnKleverchain))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnKleverchain))
	assert.Equal(t, 0, executor.GetFunctionCounter(performActionOnKleverchain))

	assert.Nil(t, eh.lastError)
}

func TestHappyCaseWhenLeaderAndActionIdNotPerformed(t *testing.T) {
	t.Parallel()

	numCalled := 0
	args := argsBridgeStub{
		myTurnHandler:          trueHandler,
		isQuorumReachedHandler: trueHandler,
		validateBatchHandler:   trueHandler,
		wasActionIDPerformedHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
		wasActionSigned:                  falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromEthereum)
	numSteps := 20
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnKleverchain))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnKleverchain))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnKleverchain))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnKleverchain))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnKleverchain))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []bridgeCore.StepIdentifier{
		getAndStoreActionIDForProposeTransferOnKleverchain,
		getAndStoreBatchFromEthereum,
		getLastExecutedEthBatchIDFromKleverchain,
		verifyLastDepositNonceExecutedOnEthereumBatch,
		wasTransferProposedOnKleverchain,
		proposeTransferOnKleverchain,
		wasTransferProposedOnKleverchain,
		signActionOnKleverchain,
		processQuorumReachedOnKleverchain,
		wasActionPerformedOnKleverchain,
		performActionOnKleverchain,
	}

	for _, stepThatError := range stepsThatCanError {
		testErrorFlow(t, stepThatError)
	}
}

func testErrorFlow(t *testing.T, stepThatErrors bridgeCore.StepIdentifier) {
	numCalled := 0
	args := argsBridgeStub{
		failingStep:            string(stepThatErrors),
		myTurnHandler:          trueHandler,
		isQuorumReachedHandler: trueHandler,
		validateBatchHandler:   trueHandler,
		wasActionIDPerformedHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
		wasActionSigned:                  falseHandler,
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromEthereum)

	maxNumSteps := 10
	for i := 0; i < maxNumSteps; i++ {
		err := sm.Execute(context.Background())
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == GettingPendingBatchFromEthereum {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
