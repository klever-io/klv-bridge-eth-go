package kctoeth

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
	getBatchFromKc                               = "GetBatchFromKc"
	storeBatchFromKc                             = "StoreBatchFromKc"
	wasTransferPerformedOnEthereum               = "WasTransferPerformedOnEthereum"
	signTransferOnEthereum                       = "SignTransferOnEthereum"
	ProcessMaxQuorumRetriesOnEthereum            = "ProcessMaxQuorumRetriesOnEthereum"
	processQuorumReachedOnEthereum               = "ProcessQuorumReachedOnEthereum"
	performTransferOnEthereum                    = "PerformTransferOnEthereum"
	getBatchStatusesFromEthereum                 = "GetBatchStatusesFromEthereum"
	wasSetStatusProposedOnKc                     = "WasSetStatusProposedOnKc"
	proposeSetStatusOnKc                         = "ProposeSetStatusOnKc"
	getAndStoreActionIDForProposeSetStatusFromKc = "GetAndStoreActionIDForProposeSetStatusFromKc"
	wasActionSignedOnKc                          = "WasActionSignedOnKc"
	signActionOnKc                               = "SignActionOnKc"
	ProcessMaxQuorumRetriesOnKc                  = "ProcessMaxQuorumRetriesOnKc"
	processQuorumReachedOnKc                     = "ProcessQuorumReachedOnKc"
	wasActionPerformedOnKc                       = "WasActionPerformedOnKc"
	performActionOnKc                            = "PerformActionOnKc"
	resetRetriesCountOnEthereum                  = "ResetRetriesCountOnEthereum"
	resetRetriesCountOnKc                        = "ResetRetriesCountOnKc"
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
	processQuorumReachedOnKcHandler       func() bool
	myTurnHandler                         func() bool
	wasSetStatusProposedOnKcHandler       func() bool
	wasActionSignedOnKcHandler            func() bool
	wasActionPerformedOnKcHandler         func() bool
	maxRetriesReachedEthereumHandler      func() bool
	maxRetriesReachedKcHandler            func() bool
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
	stub.GetAndStoreActionIDForProposeSetStatusFromKcCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeSetStatusFromKc {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 2, errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredActionIDCalled = func() uint64 {
		return 2
	}
	stub.GetBatchFromKcCalled = func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
		if args.failingStep == getBatchFromKc {
			return &bridgeCore.TransferBatch{}, errHandler.storeAndReturnError(expectedErr)
		}
		return &bridgeCore.TransferBatch{}, errHandler.storeAndReturnError(nil)
	}
	stub.StoreBatchFromKcCalled = func(batch *bridgeCore.TransferBatch) error {
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
	stub.WasSetStatusProposedOnKcCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasSetStatusProposedOnKc {
			return false, errHandler.storeAndReturnError(expectedErr)
		}
		return args.wasSetStatusProposedOnKcHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeSetStatusOnKcCalled = func(ctx context.Context) error {
		if args.failingStep == proposeSetStatusOnKc {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnKcCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnKc {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSignedOnKcHandler(), errHandler.storeAndReturnError(nil)
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

		return args.processQuorumReachedOnKcHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnKcCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnKc {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionPerformedOnKcHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnKcCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnKc {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnKcCalled = func() bool {
		return args.maxRetriesReachedKcHandler()
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
		processQuorumReachedOnKcHandler:       trueHandler,
		wasActionSignedOnKcHandler:            trueHandler,
		wasActionPerformedOnKcHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler: falseHandler,
		maxRetriesReachedEthereumHandler:      falseHandler,
		maxRetriesReachedKcHandler:            falseHandler,
		wasSetStatusProposedOnKcHandler:       falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromKc)
	numSteps := 12
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnKc))
	assert.Equal(t, 2, executor.GetFunctionCounter(getBatchFromKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(storeBatchFromKc))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(getStoredBatch))
	assert.Equal(t, 1, executor.GetFunctionCounter(signTransferOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(myTurnAsLeader))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(waitForTransferConfirmation))
	assert.Equal(t, 1, executor.GetFunctionCounter(resolveNewDepositsStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(wasSetStatusProposedOnKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(performTransferOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(WaitAndReturnFinalBatchStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(proposeSetStatusOnKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(getAndStoreActionIDForProposeSetStatusFromKc))
	assert.Equal(t, 2, executor.GetFunctionCounter(wasActionPerformedOnKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnKc))

	assert.Equal(t, 1, executor.GetFunctionCounter(wasActionSignedOnKc))
	assert.Equal(t, 1, executor.GetFunctionCounter(getStoredActionID))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []bridgeCore.StepIdentifier{
		getBatchFromKc,
		wasTransferPerformedOnEthereum,
		signTransferOnEthereum,
		processQuorumReachedOnEthereum,
		performTransferOnEthereum,
		wasSetStatusProposedOnKc,
		proposeSetStatusOnKc,
		getAndStoreActionIDForProposeSetStatusFromKc,
		wasActionSignedOnKc,
		processQuorumReachedOnKc,
		wasActionPerformedOnKc,
		performActionOnKc,
		signActionOnKc,
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
		processQuorumReachedOnKcHandler:       trueHandler,
		wasActionSignedOnKcHandler:            trueHandler,
		wasActionPerformedOnKcHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler: falseHandler,
		maxRetriesReachedEthereumHandler:      falseHandler,
		maxRetriesReachedKcHandler:            falseHandler,
		wasSetStatusProposedOnKcHandler:       falseHandler,
	}

	if stepThatErrors == "SignActionOnKc" {
		args.wasActionSignedOnKcHandler = falseHandler
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromKc)

	maxNumSteps := 12
	for i := 0; i < maxNumSteps; i++ {
		err := sm.Execute(context.Background())
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == GettingPendingBatchFromKc {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
