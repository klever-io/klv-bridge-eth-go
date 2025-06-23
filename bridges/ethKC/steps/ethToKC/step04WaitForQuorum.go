package ethtokc

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKC/steps"
	"github.com/klever-io/klv-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type waitForQuorumStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumStep) Execute(ctx context.Context) core.StepIdentifier {
	if step.bridge.ProcessMaxQuorumRetriesOnKC() {
		step.bridge.PrintInfo(logger.LogDebug, "max number of retries reached, resetting counter")
		return GettingPendingBatchFromEthereum
	}

	isQuorumReached, err := step.bridge.ProcessQuorumReachedOnKC(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while checking the quorum", "error", err)
		return GettingPendingBatchFromEthereum
	}

	step.bridge.PrintInfo(logger.LogDebug, "quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier()
	}

	return PerformingActionID
}

// Identifier returns the step's identifier
func (step *waitForQuorumStep) Identifier() core.StepIdentifier {
	return WaitingForQuorum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumStep) IsInterfaceNil() bool {
	return step == nil
}
