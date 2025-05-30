package kleverchaintoeth

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
	getBatchFromKleverchain                               = "GetBatchFromKleverchain"
	storeBatchFromKleverchain                             = "StoreBatchFromKleverchain"
	wasTransferPerformedOnEthereum                        = "WasTransferPerformedOnEthereum"
	signTransferOnEthereum                                = "SignTransferOnEthereum"
	ProcessMaxQuorumRetriesOnEthereum                     = "ProcessMaxQuorumRetriesOnEthereum"
	processQuorumReachedOnEthereum                        = "ProcessQuorumReachedOnEthereum"
	performTransferOnEthereum                             = "PerformTransferOnEthereum"
	getBatchStatusesFromEthereum                          = "GetBatchStatusesFromEthereum"
	wasSetStatusProposedOnKleverchain                     = "WasSetStatusProposedOnKleverchain"
	proposeSetStatusOnKleverchain                         = "ProposeSetStatusOnKleverchain"
	getAndStoreActionIDForProposeSetStatusFromKleverchain = "GetAndStoreActionIDForProposeSetStatusFromKleverchain"
	wasActionSignedOnKleverchain                          = "WasActionSignedOnKleverchain"
	signActionOnKleverchain                               = "SignActionOnKleverchain"
	ProcessMaxQuorumRetriesOnKleverchain                  = "ProcessMaxQuorumRetriesOnKleverchain"
	processQuorumReachedOnKleverchain                     = "ProcessQuorumReachedOnKleverchain"
	wasActionPerformedOnKleverchain                       = "WasActionPerformedOnKleverchain"
	performActionOnKleverchain                            = "PerformActionOnKleverchain"
	resetRetriesCountOnEthereum                           = "ResetRetriesCountOnEthereum"
	resetRetriesCountOnKleverchain                        = "ResetRetriesCountOnKleverchain"
	getStoredBatch                                        = "GetStoredBatch"
	myTurnAsLeader                                        = "MyTurnAsLeader"
	waitForTransferConfirmation                           = "WaitForTransferConfirmation"
	WaitAndReturnFinalBatchStatuses                       = "WaitAndReturnFinalBatchStatuses"
	resolveNewDepositsStatuses                            = "ResolveNewDepositsStatuses"
	getStoredActionID                                     = "GetStoredActionID"
)

type argsBridgeStub struct {
	failingStep                              string
	wasTransferPerformedOnEthereumHandler    func() bool
	processQuorumReachedOnEthereumHandler    func() bool
	processQuorumReachedOnKleverchainHandler func() bool
	myTurnHandler                            func() bool
	wasSetStatusProposedOnKleverchainHandler func() bool
	wasActionSignedOnKleverchainHandler      func() bool
	wasActionPerformedOnKleverchainHandler   func() bool
	maxRetriesReachedEthereumHandler         func() bool
	maxRetriesReachedKleverchainHandler      func() bool
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
	stub.GetAndStoreActionIDForProposeSetStatusFromKleverchainCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeSetStatusFromKleverchain {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 2, errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredActionIDCalled = func() uint64 {
		return 2
	}
	stub.GetBatchFromKleverchainCalled = func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
		if args.failingStep == getBatchFromKleverchain {
			return &bridgeCore.TransferBatch{}, errHandler.storeAndReturnError(expectedErr)
		}
		return &bridgeCore.TransferBatch{}, errHandler.storeAndReturnError(nil)
	}
	stub.StoreBatchFromKleverchainCalled = func(batch *bridgeCore.TransferBatch) error {
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
	stub.WasSetStatusProposedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasSetStatusProposedOnKleverchain {
			return false, errHandler.storeAndReturnError(expectedErr)
		}
		return args.wasSetStatusProposedOnKleverchainHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeSetStatusOnKleverchainCalled = func(ctx context.Context) error {
		if args.failingStep == proposeSetStatusOnKleverchain {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnKleverchain {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSignedOnKleverchainHandler(), errHandler.storeAndReturnError(nil)
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

		return args.processQuorumReachedOnKleverchainHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnKleverchainCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnKleverchain {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionPerformedOnKleverchainHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnKleverchainCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnKleverchain {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnKleverchainCalled = func() bool {
		return args.maxRetriesReachedKleverchainHandler()
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
		myTurnHandler:                            trueHandler,
		processQuorumReachedOnEthereumHandler:    trueHandler,
		processQuorumReachedOnKleverchainHandler: trueHandler,
		wasActionSignedOnKleverchainHandler:      trueHandler,
		wasActionPerformedOnKleverchainHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler:    falseHandler,
		maxRetriesReachedEthereumHandler:         falseHandler,
		maxRetriesReachedKleverchainHandler:      falseHandler,
		wasSetStatusProposedOnKleverchainHandler: falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromKleverchain)
	numSteps := 12
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnKleverchain))
	assert.Equal(t, 2, executor.GetFunctionCounter(getBatchFromKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(storeBatchFromKleverchain))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(getStoredBatch))
	assert.Equal(t, 1, executor.GetFunctionCounter(signTransferOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(myTurnAsLeader))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(waitForTransferConfirmation))
	assert.Equal(t, 1, executor.GetFunctionCounter(resolveNewDepositsStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(wasSetStatusProposedOnKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(performTransferOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(WaitAndReturnFinalBatchStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(proposeSetStatusOnKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(getAndStoreActionIDForProposeSetStatusFromKleverchain))
	assert.Equal(t, 2, executor.GetFunctionCounter(wasActionPerformedOnKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnKleverchain))

	assert.Equal(t, 1, executor.GetFunctionCounter(wasActionSignedOnKleverchain))
	assert.Equal(t, 1, executor.GetFunctionCounter(getStoredActionID))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []bridgeCore.StepIdentifier{
		getBatchFromKleverchain,
		wasTransferPerformedOnEthereum,
		signTransferOnEthereum,
		processQuorumReachedOnEthereum,
		performTransferOnEthereum,
		wasSetStatusProposedOnKleverchain,
		proposeSetStatusOnKleverchain,
		getAndStoreActionIDForProposeSetStatusFromKleverchain,
		wasActionSignedOnKleverchain,
		processQuorumReachedOnKleverchain,
		wasActionPerformedOnKleverchain,
		performActionOnKleverchain,
		signActionOnKleverchain,
	}

	for _, stepThatError := range stepsThatCanError {
		testErrorFlow(t, stepThatError)
	}
}

func testErrorFlow(t *testing.T, stepThatErrors bridgeCore.StepIdentifier) {
	t.Logf("\n\n\nnew test for stepThatError: %s", stepThatErrors)
	numCalled := 0
	args := argsBridgeStub{
		failingStep:                              string(stepThatErrors),
		myTurnHandler:                            trueHandler,
		processQuorumReachedOnEthereumHandler:    trueHandler,
		processQuorumReachedOnKleverchainHandler: trueHandler,
		wasActionSignedOnKleverchainHandler:      trueHandler,
		wasActionPerformedOnKleverchainHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler:    falseHandler,
		maxRetriesReachedEthereumHandler:         falseHandler,
		maxRetriesReachedKleverchainHandler:      falseHandler,
		wasSetStatusProposedOnKleverchainHandler: falseHandler,
	}

	if stepThatErrors == "SignActionOnKleverchain" {
		args.wasActionSignedOnKleverchainHandler = falseHandler
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromKleverchain)

	maxNumSteps := 12
	for i := 0; i < maxNumSteps; i++ {
		err := sm.Execute(context.Background())
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == GettingPendingBatchFromKleverchain {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
