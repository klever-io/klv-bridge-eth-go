package kctoeth

import (
	"context"
	"errors"

	"github.com/klever-io/klv-bridge-eth-go/bridges/ethKC/steps"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type resolveSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *resolveSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	step.bridge.ClearStoredP2PSignaturesForEthereum()
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch stored")
		return GettingPendingBatchFromKC
	}

	batch, err := step.GetBatchFromKC(ctx)
	isEmptyBatch := batch == nil || (err != nil && errors.Is(err, clients.ErrNoPendingBatchAvailable))
	if isEmptyBatch {
		step.bridge.PrintInfo(logger.LogDebug, "nil/empty batch fetched")
		return GettingPendingBatchFromKC
	}
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while fetching batch", "error", err)
		return GettingPendingBatchFromKC
	}

	statuses := step.bridge.WaitAndReturnFinalBatchStatuses(ctx)
	if len(statuses) == 0 {
		return GettingPendingBatchFromKC
	}

	storedBatch.Statuses = statuses

	step.bridge.ResolveNewDepositsStatuses(uint64(len(batch.Statuses)))

	return ProposingSetStatusOnKC
}

// Identifier returns the step's identifier
func (step *resolveSetStatusStep) Identifier() core.StepIdentifier {
	return ResolvingSetStatusOnKC
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *resolveSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}

// GetBatchFromKC fetches the batch from the Klever Blockchain
func (step *resolveSetStatusStep) GetBatchFromKC(ctx context.Context) (*core.TransferBatch, error) {
	return step.bridge.GetBatchFromKC(ctx)
}
