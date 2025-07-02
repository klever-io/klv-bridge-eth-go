package ethtokc

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKC/steps"
	"github.com/klever-io/klv-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type proposeTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no batch found")
		return GettingPendingBatchFromEthereum
	}

	wasTransferProposed, err := step.bridge.WasTransferProposedOnKC(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the batch was proposed or not on KC",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	if wasTransferProposed {
		return SigningProposedTransferOnKC
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
		return step.Identifier()
	}

	err = step.bridge.ProposeTransferOnKC(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error proposing transfer on KC",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	return SigningProposedTransferOnKC
}

// Identifier returns the step's identifier
func (step *proposeTransferStep) Identifier() core.StepIdentifier {
	return ProposingTransferOnKC
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
