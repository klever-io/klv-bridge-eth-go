package kctoeth

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKc/steps"
	"github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type getPendingStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) core.StepIdentifier {
	err := step.bridge.CheckKcClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "Klever Blockchain client unavailable", "message", err)
	}
	err = step.bridge.CheckEthereumClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "Ethereum client unavailable", "message", err)
	}
	step.bridge.ResetRetriesCountOnEthereum()
	step.resetCountersOnKc()

	batch, err := step.bridge.GetBatchFromKc(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "cannot fetch Klever Blockchain batch", "message", err)
		return step.Identifier()
	}
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no new batch found on Kc")
		return step.Identifier()
	}

	err = step.bridge.StoreBatchFromKc(batch)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error storing Klever Blockchain batch", "error", err)
		return step.Identifier()
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched new batch from Klever Blockchain "+batch.String())

	wasPerformed, err := step.bridge.WasTransferPerformedOnEthereum(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if transfer was performed or not", "error", err)
		return step.Identifier()
	}
	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "transfer performed")
		return ResolvingSetStatusOnKc
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
	return GettingPendingBatchFromKc
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}

func (step *getPendingStep) resetCountersOnKc() {
	step.bridge.ResetRetriesCountOnKc()
	step.bridge.ResetRetriesOnWasTransferProposedOnKc()
}
