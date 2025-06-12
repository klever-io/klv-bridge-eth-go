package ethtokc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKc/steps"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	getAndStoreBatchFromEthereum                  = "GetAndStoreBatchFromEthereum"
	getLastExecutedEthBatchIDFromKc               = "GetLastExecutedEthBatchIDFromKc"
	verifyLastDepositNonceExecutedOnEthereumBatch = "VerifyLastDepositNonceExecutedOnEthereumBatch"
	wasTransferProposedOnKc                       = "WasTransferProposedOnKc"
	wasActionSignedOnKc                           = "WasActionSignedOnKc"
	signActionOnKc                                = "SignActionOnKc"
	getAndStoreActionIDForProposeTransferOnKc     = "GetAndStoreActionIDForProposeTransferOnKc"
	ProcessMaxQuorumRetriesOnKc                   = "ProcessMaxQuorumRetriesOnKc"
	resetRetriesCountOnKc                         = "ResetRetriesCountOnKc"
	processQuorumReachedOnKc                      = "ProcessQuorumReachedOnKc"
	wasActionPerformedOnKc                        = "WasActionPerformedOnKc"
	proposeTransferOnKc                           = "ProposeTransferOnKc"
	performActionOnKc                             = "PerformActionOnKc"
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
	stub.GetAndStoreActionIDForProposeTransferOnKcCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeTransferOnKc {
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
	stub.GetLastExecutedEthBatchIDFromKcCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getLastExecutedEthBatchIDFromKc {
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
	stub.WasTransferProposedOnKcCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasTransferProposedOnKc {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasTransferProposedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeTransferOnKcCalled = func(ctx context.Context) error {
		if args.failingStep == proposeTransferOnKc {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnKcCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnKc {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSigned(), errHandler.storeAndReturnError(nil)
	}
	stub.SignActionOnKcCalled = func(ctx context.Context) error {
		if args.failingStep == signActionOnKc {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessQuorumReachedOnKcCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == processQuorumReachedOnKc {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.isQuorumReachedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnKcCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnKc {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionIDPerformedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnKcCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnKc {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnKcCalled = func() bool {
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

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnKc))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnKc))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnKc))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnKc))
	assert.Equal(t, 0, executor.GetFunctionCounter(performActionOnKc))

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

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnKc))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnKc))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnKc))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnKc))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnKc))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []bridgeCore.StepIdentifier{
		getAndStoreActionIDForProposeTransferOnKc,
		getAndStoreBatchFromEthereum,
		getLastExecutedEthBatchIDFromKc,
		verifyLastDepositNonceExecutedOnEthereumBatch,
		wasTransferProposedOnKc,
		proposeTransferOnKc,
		wasTransferProposedOnKc,
		signActionOnKc,
		processQuorumReachedOnKc,
		wasActionPerformedOnKc,
		performActionOnKc,
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
