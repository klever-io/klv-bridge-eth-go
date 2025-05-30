package kleverchaintoeth

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKleverchain/steps"
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
		return GettingPendingBatchFromKleverchain
	}

	if step.bridge.ProcessMaxRetriesOnWasTransferProposedOnKleverchain() {
		step.bridge.PrintInfo(logger.LogDebug, "max number of retries reached, resetting counter")
		return GettingPendingBatchFromKleverchain
	}

	wasSetStatusProposed, err := step.bridge.WasSetStatusProposedOnKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the set status action was proposed or not on Kleverchain",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromKleverchain
	}

	if wasSetStatusProposed {
		return SigningProposedSetStatusOnKleverchain
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
		return step.Identifier()
	}

	err = step.bridge.ProposeSetStatusOnKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error proposing transfer on Kleverchain",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromKleverchain
	}

	return SigningProposedSetStatusOnKleverchain
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return ProposingSetStatusOnKleverchain
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
