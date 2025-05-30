package kleverchaintoeth

import (
	"context"

	ethKleverchain "github.com/klever-io/klv-bridge-eth-go/bridges/ethKleverchain"
	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKleverchain/steps"
	"github.com/klever-io/klv-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type signProposedSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil stored batch")
		return GettingPendingBatchFromKleverchain
	}

	actionID, err := step.bridge.GetAndStoreActionIDForProposeSetStatusFromKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error fetching action ID", "batch ID", storedBatch.ID, "error", err)
		return GettingPendingBatchFromKleverchain
	}
	if actionID == ethKleverchain.InvalidActionID {
		step.bridge.PrintInfo(logger.LogError, "contract error, got invalid action ID",
			"batch ID", storedBatch.ID, "error", err, "action ID", actionID)
		return GettingPendingBatchFromKleverchain
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched action ID", "action ID", actionID, "batch ID", storedBatch.ID)

	wasSigned, err := step.bridge.WasActionSignedOnKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the proposed transfer was signed or not",
			"batch ID", storedBatch.ID, "error", err)
		return GettingPendingBatchFromKleverchain
	}

	if wasSigned {
		return WaitingForQuorumOnSetStatus
	}

	err = step.bridge.SignActionOnKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error signing the proposed set status",
			"batch ID", storedBatch.ID, "error", err)
		return GettingPendingBatchFromKleverchain
	}

	return WaitingForQuorumOnSetStatus
}

// Identifier returns the step's identifier
func (step *signProposedSetStatusStep) Identifier() core.StepIdentifier {
	return SigningProposedSetStatusOnKleverchain
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *signProposedSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
