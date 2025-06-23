package kctoeth

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKC/steps"
	"github.com/klever-io/klv-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type proposeSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch stored")
		return GettingPendingBatchFromKC
	}

	if step.bridge.ProcessMaxRetriesOnWasTransferProposedOnKC() {
		step.bridge.PrintInfo(logger.LogDebug, "max number of retries reached, resetting counter")
		return GettingPendingBatchFromKC
	}

	wasSetStatusProposed, err := step.bridge.WasSetStatusProposedOnKC(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the set status action was proposed or not on KC",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromKC
	}

	if wasSetStatusProposed {
		return SigningProposedSetStatusOnKC
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
		return step.Identifier()
	}

	err = step.bridge.ProposeSetStatusOnKC(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error proposing transfer on KC",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromKC
	}

	return SigningProposedSetStatusOnKC
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return ProposingSetStatusOnKC
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
