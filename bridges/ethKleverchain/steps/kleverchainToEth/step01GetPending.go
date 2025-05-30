package kleverchaintoeth

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKleverchain/steps"
	"github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type getPendingStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) core.StepIdentifier {
	err := step.bridge.CheckKleverchainClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "Kleverchain client unavailable", "message", err)
	}
	err = step.bridge.CheckEthereumClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "Ethereum client unavailable", "message", err)
	}
	step.bridge.ResetRetriesCountOnEthereum()
	step.resetCountersOnKleverchain()

	batch, err := step.bridge.GetBatchFromKleverchain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "cannot fetch Kleverchain batch", "message", err)
		return step.Identifier()
	}
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no new batch found on Kleverchain")
		return step.Identifier()
	}

	err = step.bridge.StoreBatchFromKleverchain(batch)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error storing Kleverchain batch", "error", err)
		return step.Identifier()
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched new batch from Kleverchain "+batch.String())

	wasPerformed, err := step.bridge.WasTransferPerformedOnEthereum(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if transfer was performed or not", "error", err)
		return step.Identifier()
	}
	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "transfer performed")
		return ResolvingSetStatusOnKleverchain
	}

	argLists := batchProcessor.ExtractListKlvToEth(batch)
	err = step.bridge.CheckAvailableTokens(ctx, argLists.EthTokens, argLists.KdaTokenBytes, argLists.Amounts, argLists.Direction)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error checking available tokens", "error", err, "batch", batch.String())
		return step.Identifier()
	}

	return SigningProposedTransferOnEthereum
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return GettingPendingBatchFromKleverchain
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}

func (step *getPendingStep) resetCountersOnKleverchain() {
	step.bridge.ResetRetriesCountOnKleverchain()
	step.bridge.ResetRetriesOnWasTransferProposedOnKleverchain()
}
