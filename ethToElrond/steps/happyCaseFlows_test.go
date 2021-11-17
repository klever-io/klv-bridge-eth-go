package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	getPendingBatch                         = "GetPendingBatch"
	hasPendingBatch                         = "HasPendingBatch"
	isLeader                                = "IsLeader"
	proposeTransferOnDestination            = "ProposeTransferOnDestination"
	setStatusRejectedOnAllTransactions      = "SetStatusRejectedOnAllTransactions"
	waitStepToFinish                        = "WaitStepToFinish"
	wasProposeTransferExecutedOnDestination = "WasProposeTransferExecutedOnDestination"
	signProposeTransferOnDestination        = "SignProposeTransferOnDestination"
	isQuorumReachedForProposeTransfer       = "IsQuorumReachedForProposeTransfer"
	executeTransferOnDestination            = "ExecuteTransferOnDestination"
	wasTransferExecutedOnDestination        = "WasTransferExecutedOnDestination"
	cleanStoredSignatures                   = "CleanStoredSignatures"
	updateTransactionsStatusesIfNeeded      = "UpdateTransactionsStatusesIfNeeded"
	proposeSetStatusOnSource                = "ProposeSetStatusOnSource"
	wasProposeSetStatusExecutedOnSource     = "WasProposeSetStatusExecutedOnSource"
	signProposeSetStatusOnSource            = "SignProposeSetStatusOnSource"
	isQuorumReachedForProposeSetStatus      = "IsQuorumReachedForProposeSetStatus"
	executeSetStatusOnSource                = "ExecuteSetStatusOnSource"
	wasSetStatusExecutedOnSource            = "WasSetStatusExecutedOnSource"
)

var trueHandler = func() bool { return true }
var trueHandlerWithContext = func(_ context.Context) bool { return true }
var falseHandler = func() bool { return false }
var falseHandlerWithContext = func(_ context.Context) bool { return false }

func setAllDecisionHandlersToTrue(bem *stateMachine.BridgeExecutorMock) {
	bem.HasPendingBatchCalled = trueHandler
	bem.IsPendingBatchReadyCalled = func(ctx context.Context) (bool, error) {
		return true, nil
	}
	bem.IsLeaderCalled = trueHandler
	bem.WasProposeTransferExecutedOnDestinationCalled = trueHandlerWithContext
	bem.WasProposeSetStatusExecutedOnSourceCalled = trueHandlerWithContext
	bem.WasTransferExecutedOnDestinationCalled = trueHandlerWithContext
	bem.WasSetStatusExecutedOnSourceCalled = trueHandlerWithContext
	bem.IsQuorumReachedForProposeTransferCalled = trueHandlerWithContext
	bem.IsQuorumReachedForProposeSetStatusCalled = trueHandlerWithContext
}

func TestGetPendingEndlessLoop(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	bem.HasPendingBatchCalled = falseHandler

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 100
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	assert.Equal(t, numSteps, bem.GetFunctionCounter("GetPendingBatch"))
}

func TestFlowAsLeaderForTwoCompleteFlows(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 14
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.ExecutingSetStatus,
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.ExecutingSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
}

func TestFlowAsSignerForTwoCompleteFlows(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsLeaderCalled = falseHandler

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 14
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.ExecutingSetStatus,
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.ExecutingSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
}

func TestFlowAsLeaderForOneCompleteFlowWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 7
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.ExecutingSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter(getPendingBatch))
	assert.Equal(t, 1, bem.GetFunctionCounter(hasPendingBatch))
	assert.Equal(t, 4, bem.GetFunctionCounter(isLeader))
	assert.Equal(t, 1, bem.GetFunctionCounter(proposeTransferOnDestination))
	assert.Equal(t, 6, bem.GetFunctionCounter(waitStepToFinish))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasProposeTransferExecutedOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(signProposeTransferOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(isQuorumReachedForProposeTransfer))
	assert.Equal(t, 1, bem.GetFunctionCounter(executeTransferOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasTransferExecutedOnDestination))
	assert.Equal(t, 2, bem.GetFunctionCounter(cleanStoredSignatures))
	assert.Equal(t, 1, bem.GetFunctionCounter(updateTransactionsStatusesIfNeeded))
	assert.Equal(t, 1, bem.GetFunctionCounter(proposeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasProposeSetStatusExecutedOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(signProposeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(isQuorumReachedForProposeSetStatus))
	assert.Equal(t, 1, bem.GetFunctionCounter(executeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasSetStatusExecutedOnSource))
}

func TestFlowAsSignerForOneCompleteFlowWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsLeaderCalled = falseHandler

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 7
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.ExecutingSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter(getPendingBatch))
	assert.Equal(t, 1, bem.GetFunctionCounter(hasPendingBatch))
	assert.Equal(t, 4, bem.GetFunctionCounter(isLeader))
	assert.Equal(t, 0, bem.GetFunctionCounter(proposeTransferOnDestination))
	assert.Equal(t, 6, bem.GetFunctionCounter(waitStepToFinish))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasProposeTransferExecutedOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(signProposeTransferOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(isQuorumReachedForProposeTransfer))
	assert.Equal(t, 0, bem.GetFunctionCounter(executeTransferOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasTransferExecutedOnDestination))
	assert.Equal(t, 2, bem.GetFunctionCounter(cleanStoredSignatures))
	assert.Equal(t, 1, bem.GetFunctionCounter(updateTransactionsStatusesIfNeeded))
	assert.Equal(t, 0, bem.GetFunctionCounter(proposeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasProposeSetStatusExecutedOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(signProposeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(isQuorumReachedForProposeSetStatus))
	assert.Equal(t, 0, bem.GetFunctionCounter(executeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasSetStatusExecutedOnSource))
}
