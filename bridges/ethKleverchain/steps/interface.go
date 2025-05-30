package steps

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// Executor defines a generic bridge interface able to handle both halves of the bridge
type Executor interface {
	PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeader() bool

	GetBatchFromKleverchain(ctx context.Context) (*bridgeCore.TransferBatch, error)
	StoreBatchFromKleverchain(batch *bridgeCore.TransferBatch) error
	GetStoredBatch() *bridgeCore.TransferBatch

	GetLastExecutedEthBatchIDFromKleverchain(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error

	GetAndStoreActionIDForProposeTransferOnKleverchain(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromKleverchain(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64

	WasTransferProposedOnKleverchain(ctx context.Context) (bool, error)
	ProposeTransferOnKleverchain(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnKleverchain() bool
	ResetRetriesOnWasTransferProposedOnKleverchain()

	WasSetStatusProposedOnKleverchain(ctx context.Context) (bool, error)
	ProposeSetStatusOnKleverchain(ctx context.Context) error

	WasActionSignedOnKleverchain(ctx context.Context) (bool, error)
	SignActionOnKleverchain(ctx context.Context) error

	ProcessQuorumReachedOnKleverchain(ctx context.Context) (bool, error)
	WasActionPerformedOnKleverchain(ctx context.Context) (bool, error)
	PerformActionOnKleverchain(ctx context.Context) error
	ResolveNewDepositsStatuses(numDeposits uint64)

	ProcessMaxQuorumRetriesOnKleverchain() bool
	ResetRetriesCountOnKleverchain()

	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum() error
	PerformTransferOnEthereum(ctx context.Context) error
	ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	WaitForTransferConfirmation(ctx context.Context)
	WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte
	GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error)

	ProcessMaxQuorumRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()
	ClearStoredP2PSignaturesForEthereum()

	CheckKleverchainClientAvailability(ctx context.Context) error
	CheckEthereumClientAvailability(ctx context.Context) error
	CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error

	IsInterfaceNil() bool
}
