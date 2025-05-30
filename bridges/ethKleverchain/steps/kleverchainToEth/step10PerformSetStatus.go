package kleverchaintoeth

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKleverchain/steps"
	"github.com/klever-io/klv-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type performSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *performSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	wasPerformed, err := step.bridge.WasActionPerformedOnKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the set status was proposed or not",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return GettingPendingBatchFromKleverchain
	}

	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "action ID performed",
			"action ID", step.bridge.GetStoredActionID())
		return GettingPendingBatchFromKleverchain
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
		return step.Identifier()
	}

	err = step.bridge.PerformActionOnKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error performing action ID",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return GettingPendingBatchFromKleverchain
	}

	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *performSetStatusStep) Identifier() core.StepIdentifier {
	return PerformingSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *performSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
