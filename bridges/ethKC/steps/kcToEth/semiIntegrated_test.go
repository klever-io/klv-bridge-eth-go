package kctoeth

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKC/steps"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	getBatchFromKC                               = "GetBatchFromKC"
	storeBatchFromKC                             = "StoreBatchFromKC"
	wasTransferPerformedOnEthereum               = "WasTransferPerformedOnEthereum"
	signTransferOnEthereum                       = "SignTransferOnEthereum"
	ProcessMaxQuorumRetriesOnEthereum            = "ProcessMaxQuorumRetriesOnEthereum"
	processQuorumReachedOnEthereum               = "ProcessQuorumReachedOnEthereum"
	performTransferOnEthereum                    = "PerformTransferOnEthereum"
	getBatchStatusesFromEthereum                 = "GetBatchStatusesFromEthereum"
	wasSetStatusProposedOnKC                     = "WasSetStatusProposedOnKC"
	proposeSetStatusOnKC                         = "ProposeSetStatusOnKC"
	getAndStoreActionIDForProposeSetStatusFromKC = "GetAndStoreActionIDForProposeSetStatusFromKC"
	wasActionSignedOnKC                          = "WasActionSignedOnKC"
	signActionOnKC                               = "SignActionOnKC"
	ProcessMaxQuorumRetriesOnKC                  = "ProcessMaxQuorumRetriesOnKC"
	processQuorumReachedOnKC                     = "ProcessQuorumReachedOnKC"
	wasActionPerformedOnKC                       = "WasActionPerformedOnKC"
	performActionOnKC                            = "PerformActionOnKC"
	resetRetriesCountOnEthereum                  = "ResetRetriesCountOnEthereum"
	resetRetriesCountOnKC                        = "ResetRetriesCountOnKC"
	getStoredBatch                               = "GetStoredBatch"
	myTurnAsLeader                               = "MyTurnAsLeader"
	waitForTransferConfirmation                  = "WaitForTransferConfirmation"
	WaitAndReturnFinalBatchStatuses              = "WaitAndReturnFinalBatchStatuses"
	resolveNewDepositsStatuses                   = "ResolveNewDepositsStatuses"
	getStoredActionID                            = "GetStoredActionID"
)

type argsBridgeStub struct {
	failingStep                           string
	wasTransferPerformedOnEthereumHandler func() bool
	processQuorumReachedOnEthereumHandler func() bool
	processQuorumReachedOnKCHandler       func() bool
	myTurnHandler                         func() bool
	wasSetStatusProposedOnKCHandler       func() bool
	wasActionSignedOnKCHandler            func() bool
	wasActionPerformedOnKCHandler         func() bool
	maxRetriesReachedEthereumHandler      func() bool
	maxRetriesReachedKCHandler            func() bool
}

var trueHandler = func() bool { return true }
var falseHandler = func() bool { return false }

type errorHandler struct {
	lastError error
}

func (eh *errorHandler) storeAndReturnError(err error) error {
	eh.lastError = err
	return err
}

func createStateMachine(t *testing.T, executor steps.Executor, initialStep bridgeCore.StepIdentifier) *stateMachine.StateMachineMock {
	stepsSlice, err := CreateSteps(executor)
	require.Nil(t, err)

	sm := stateMachine.NewStateMachineMock(stepsSlice, initialStep)
	err = sm.Initialize()
	require.Nil(t, err)

	return sm
}

func createMockBridge(args argsBridgeStub) (*bridgeTests.BridgeExecutorStub, *errorHandler) {
	errHandler := &errorHandler{}
	stub := bridgeTests.NewBridgeExecutorStub()
	expectedErr := errors.New("expected error")
	stub.MyTurnAsLeaderCalled = func() bool {
		return args.myTurnHandler()
	}
	stub.GetAndStoreActionIDForProposeSetStatusFromKCCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeSetStatusFromKC {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 2, errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredActionIDCalled = func() uint64 {
		return 2
	}
	stub.GetBatchFromKCCalled = func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
		if args.failingStep == getBatchFromKC {
			return &bridgeCore.TransferBatch{}, errHandler.storeAndReturnError(expectedErr)
		}
		return &bridgeCore.TransferBatch{}, errHandler.storeAndReturnError(nil)
	}
	stub.StoreBatchFromKCCalled = func(batch *bridgeCore.TransferBatch) error {
		return nil
	}
	stub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
		return &bridgeCore.TransferBatch{}
	}
	stub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasTransferPerformedOnEthereum {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasTransferPerformedOnEthereumHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.SignTransferOnEthereumCalled = func() error {
		if args.failingStep == signTransferOnEthereum {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessQuorumReachedOnEthereumCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == processQuorumReachedOnEthereum {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.processQuorumReachedOnEthereumHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformTransferOnEthereumCalled = func(ctx context.Context) error {
		if args.failingStep == performTransferOnEthereum {
			return errHandler.storeAndReturnError(expectedErr)
		}
		return errHandler.storeAndReturnError(nil)
	}
	stub.WaitForTransferConfirmationCalled = func(ctx context.Context) {
		stub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
			return true, errHandler.storeAndReturnError(nil)
		}
	}
	stub.WaitAndReturnFinalBatchStatusesCalled = func(ctx context.Context) []byte {
		if args.failingStep == getBatchStatusesFromEthereum {
			return nil
		}
		return []byte{0x3}
	}
	stub.GetBatchStatusesFromEthereumCalled = func(ctx context.Context) ([]byte, error) {
		if args.failingStep == getBatchStatusesFromEthereum {
			return nil, errHandler.storeAndReturnError(expectedErr)
		}
		return []byte{}, errHandler.storeAndReturnError(nil)
	}
	stub.ResolveNewDepositsStatusesCalled = func(numDeposits uint64) {

	}
	stub.WasSetStatusProposedOnKCCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasSetStatusProposedOnKC {
			return false, errHandler.storeAndReturnError(expectedErr)
		}
		return args.wasSetStatusProposedOnKCHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeSetStatusOnKCCalled = func(ctx context.Context) error {
		if args.failingStep == proposeSetStatusOnKC {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnKCCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnKC {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSignedOnKCHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.SignActionOnKCCalled = func(ctx context.Context) error {
		if args.failingStep == signActionOnKC {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessQuorumReachedOnKCCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == processQuorumReachedOnKC {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.processQuorumReachedOnKCHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnKCCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnKC {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionPerformedOnKCHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnKCCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnKC {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnKCCalled = func() bool {
		return args.maxRetriesReachedKCHandler()
	}
	stub.ProcessMaxQuorumRetriesOnEthereumCalled = func() bool {
		return args.maxRetriesReachedEthereumHandler()
	}

	return stub, errHandler
}

func TestHappyCaseWhenLeaderSetStatusAlreadySigned(t *testing.T) {
	t.Parallel()

	numCalled := 0
	args := argsBridgeStub{
		myTurnHandler:                         trueHandler,
		processQuorumReachedOnEthereumHandler: trueHandler,
		processQuorumReachedOnKCHandler:       trueHandler,
		wasActionSignedOnKCHandler:            trueHandler,
		wasActionPerformedOnKCHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler: falseHandler,
		maxRetriesReachedEthereumHandler:      falseHandler,
		maxRetriesReachedKCHandler:            falseHandler,
		wasSetStatusProposedOnKCHandler:       falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromKC)
	numSteps := 12
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnKC))
	assert.Equal(t, 2, executor.GetFunctionCounter(getBatchFromKC))
	assert.Equal(t, 1, executor.GetFunctionCounter(storeBatchFromKC))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(getStoredBatch))
	assert.Equal(t, 1, executor.GetFunctionCounter(signTransferOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(myTurnAsLeader))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnKC))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnKC))
	assert.Equal(t, 1, executor.GetFunctionCounter(waitForTransferConfirmation))
	assert.Equal(t, 1, executor.GetFunctionCounter(resolveNewDepositsStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(wasSetStatusProposedOnKC))
	assert.Equal(t, 1, executor.GetFunctionCounter(performTransferOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(WaitAndReturnFinalBatchStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(proposeSetStatusOnKC))
	assert.Equal(t, 1, executor.GetFunctionCounter(getAndStoreActionIDForProposeSetStatusFromKC))
	assert.Equal(t, 2, executor.GetFunctionCounter(wasActionPerformedOnKC))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnKC))

	assert.Equal(t, 1, executor.GetFunctionCounter(wasActionSignedOnKC))
	assert.Equal(t, 1, executor.GetFunctionCounter(getStoredActionID))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []bridgeCore.StepIdentifier{
		getBatchFromKC,
		wasTransferPerformedOnEthereum,
		signTransferOnEthereum,
		processQuorumReachedOnEthereum,
		performTransferOnEthereum,
		wasSetStatusProposedOnKC,
		proposeSetStatusOnKC,
		getAndStoreActionIDForProposeSetStatusFromKC,
		wasActionSignedOnKC,
		processQuorumReachedOnKC,
		wasActionPerformedOnKC,
		performActionOnKC,
		signActionOnKC,
	}

	for _, stepThatError := range stepsThatCanError {
		testErrorFlow(t, stepThatError)
	}
}

func testErrorFlow(t *testing.T, stepThatErrors bridgeCore.StepIdentifier) {
	t.Logf("\n\n\nnew test for stepThatError: %s", stepThatErrors)
	numCalled := 0
	args := argsBridgeStub{
		failingStep:                           string(stepThatErrors),
		myTurnHandler:                         trueHandler,
		processQuorumReachedOnEthereumHandler: trueHandler,
		processQuorumReachedOnKCHandler:       trueHandler,
		wasActionSignedOnKCHandler:            trueHandler,
		wasActionPerformedOnKCHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler: falseHandler,
		maxRetriesReachedEthereumHandler:      falseHandler,
		maxRetriesReachedKCHandler:            falseHandler,
		wasSetStatusProposedOnKCHandler:       falseHandler,
	}

	if stepThatErrors == "SignActionOnKC" {
		args.wasActionSignedOnKCHandler = falseHandler
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromKC)

	maxNumSteps := 12
	for i := 0; i < maxNumSteps; i++ {
		err := sm.Execute(context.Background())
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == GettingPendingBatchFromKC {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
