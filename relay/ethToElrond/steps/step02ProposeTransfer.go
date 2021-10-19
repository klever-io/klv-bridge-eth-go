package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type proposeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeTransferStep) Execute(ctx context.Context) (relay.StepIdentifier, error) {
	if step.bridge.IsLeader() {
		err := step.bridge.ProposeTransferOnDestination(ctx)
		if err != nil {
			step.bridge.PrintDebugInfo("bridge.ProposeTransfer", "error", err)
			step.bridge.SetStatusRejectedOnAllTransactions()

			return ethToElrond.ProposingSetStatus, nil
		}
	}

	step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if !step.bridge.WasProposeTransferExecutedOnDestination() {
		// remain in this step
		return step.Identifier(), nil
	}

	step.bridge.SignProposeTransferOnDestination(ctx)

	return ethToElrond.WaitingSignaturesForProposeTransfer, nil
}

// Identifier returns the step's identifier
func (step *proposeTransferStep) Identifier() relay.StepIdentifier {
	return ethToElrond.ProposingTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
