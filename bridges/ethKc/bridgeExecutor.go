package ethKc

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/clients/ethereum/contract"
	"github.com/klever-io/klv-bridge-eth-go/core"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// splits - represent the number of times we split the maximum interval
// we wait for the transfer confirmation on Ethereum
const splits = 10
const minRetries = 1

// ArgsBridgeExecutor is the arguments DTO struct used in both bridges
type ArgsBridgeExecutor struct {
	Log                        logger.Logger
	TopologyProvider           TopologyProvider
	KcClient                   KcClient
	EthereumClient             EthereumClient
	TimeForWaitOnEthereum      time.Duration
	StatusHandler              core.StatusHandler
	SignaturesHolder           SignaturesHolder
	BalanceValidator           BalanceValidator
	MaxQuorumRetriesOnEthereum uint64
	MaxQuorumRetriesOnKc       uint64
	MaxRetriesOnWasProposed    uint64
}

type bridgeExecutor struct {
	log                        logger.Logger
	topologyProvider           TopologyProvider
	kcClient                   KcClient
	ethereumClient             EthereumClient
	timeForWaitOnEthereum      time.Duration
	statusHandler              core.StatusHandler
	sigsHolder                 SignaturesHolder
	balanceValidator           BalanceValidator
	maxQuorumRetriesOnEthereum uint64
	maxQuorumRetriesOnKc       uint64
	maxRetriesOnWasProposed    uint64

	batch                   *bridgeCore.TransferBatch
	actionID                uint64
	msgHash                 common.Hash
	quorumRetriesOnEthereum uint64
	quorumRetriesOnKc       uint64
	retriesOnWasProposed    uint64
}

// NewBridgeExecutor creates a bridge executor, which can be used for both half-bridges
func NewBridgeExecutor(args ArgsBridgeExecutor) (*bridgeExecutor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	executor := createBridgeExecutor(args)
	return executor, nil
}

func checkArgs(args ArgsBridgeExecutor) error {
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if check.IfNil(args.KcClient) {
		return ErrNilKcClient
	}
	if check.IfNil(args.EthereumClient) {
		return ErrNilEthereumClient
	}
	if check.IfNil(args.TopologyProvider) {
		return ErrNilTopologyProvider
	}
	if check.IfNil(args.StatusHandler) {
		return ErrNilStatusHandler
	}
	if args.TimeForWaitOnEthereum < durationLimit {
		return ErrInvalidDuration
	}
	if check.IfNil(args.SignaturesHolder) {
		return ErrNilSignaturesHolder
	}
	if check.IfNil(args.BalanceValidator) {
		return ErrNilBalanceValidator
	}
	if args.MaxQuorumRetriesOnEthereum < minRetries {
		return fmt.Errorf("%w for args.MaxQuorumRetriesOnEthereum, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.MaxQuorumRetriesOnEthereum, minRetries)
	}
	if args.MaxQuorumRetriesOnKc < minRetries {
		return fmt.Errorf("%w for args.MaxQuorumRetriesOnKc, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.MaxQuorumRetriesOnKc, minRetries)
	}
	if args.MaxRetriesOnWasProposed < minRetries {
		return fmt.Errorf("%w for args.MaxRetriesOnWasProposed, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.MaxRetriesOnWasProposed, minRetries)
	}
	return nil
}

func createBridgeExecutor(args ArgsBridgeExecutor) *bridgeExecutor {
	return &bridgeExecutor{
		log:                        args.Log,
		kcClient:                   args.KcClient,
		ethereumClient:             args.EthereumClient,
		topologyProvider:           args.TopologyProvider,
		statusHandler:              args.StatusHandler,
		timeForWaitOnEthereum:      args.TimeForWaitOnEthereum,
		sigsHolder:                 args.SignaturesHolder,
		balanceValidator:           args.BalanceValidator,
		maxQuorumRetriesOnEthereum: args.MaxQuorumRetriesOnEthereum,
		maxQuorumRetriesOnKc:       args.MaxQuorumRetriesOnKc,
		maxRetriesOnWasProposed:    args.MaxRetriesOnWasProposed,
	}
}

// PrintInfo will print the provided data through the inner logger instance
func (executor *bridgeExecutor) PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{}) {
	executor.log.Log(logLevel, message, extras...)

	switch logLevel {
	case logger.LogWarning, logger.LogError:
		executor.setExecutionMessageInStatusHandler(logLevel, message, extras...)
	}
}

func (executor *bridgeExecutor) setExecutionMessageInStatusHandler(level logger.LogLevel, message string, extras ...interface{}) {
	msg := fmt.Sprintf("%s: %s", level, message)
	for i := 0; i < len(extras)-1; i += 2 {
		msg += fmt.Sprintf(" %s = %s", convertObjectToString(extras[i]), convertObjectToString(extras[i+1]))
	}

	executor.statusHandler.SetStringMetric(core.MetricLastError, msg)
}

// MyTurnAsLeader returns true if the current relayer node is the leader
func (executor *bridgeExecutor) MyTurnAsLeader() bool {
	return executor.topologyProvider.MyTurnAsLeader()
}

// GetBatchFromKc fetches the pending batch from Kc
func (executor *bridgeExecutor) GetBatchFromKc(ctx context.Context) (*bridgeCore.TransferBatch, error) {
	batch, err := executor.kcClient.GetPendingBatch(ctx)
	if err == nil {
		executor.statusHandler.SetIntMetric(core.MetricNumBatches, int(batch.ID)-1)
	}
	return batch, err
}

// StoreBatchFromKc saves the pending batch from Kc
func (executor *bridgeExecutor) StoreBatchFromKc(batch *bridgeCore.TransferBatch) error {
	if batch == nil {
		return ErrNilBatch
	}

	executor.batch = batch
	return nil
}

// GetStoredBatch returns the stored batch
func (executor *bridgeExecutor) GetStoredBatch() *bridgeCore.TransferBatch {
	return executor.batch
}

// GetLastExecutedEthBatchIDFromKc returns the last executed batch ID that is stored on the Klever Blockchain SC
func (executor *bridgeExecutor) GetLastExecutedEthBatchIDFromKc(ctx context.Context) (uint64, error) {
	batchID, err := executor.kcClient.GetLastExecutedEthBatchID(ctx)
	if err == nil {
		executor.statusHandler.SetIntMetric(core.MetricNumBatches, int(batchID))
	}
	return batchID, err
}

// VerifyLastDepositNonceExecutedOnEthereumBatch will check the deposit Nonces from the fetched batch from Ethereum client
func (executor *bridgeExecutor) VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	lastNonce, err := executor.kcClient.GetLastExecutedEthTxID(ctx)
	if err != nil {
		return err
	}

	return executor.verifyDepositNonces(lastNonce)
}

func (executor *bridgeExecutor) verifyDepositNonces(lastNonce uint64) error {
	startNonce := lastNonce + 1
	for _, dt := range executor.batch.Deposits {
		if dt.Nonce != startNonce {
			return fmt.Errorf("%w for deposit %s, expected: %d", ErrInvalidDepositNonce, dt.String(), startNonce)
		}

		startNonce++
	}

	return nil
}

// GetAndStoreActionIDForProposeTransferOnKc fetches the action ID for ProposeTransfer by using the stored batch. Stores the action ID and returns it
func (executor *bridgeExecutor) GetAndStoreActionIDForProposeTransferOnKc(ctx context.Context) (uint64, error) {
	if executor.batch == nil {
		return InvalidActionID, ErrNilBatch
	}

	actionID, err := executor.kcClient.GetActionIDForProposeTransfer(ctx, executor.batch)
	if err != nil {
		return InvalidActionID, err
	}

	executor.actionID = actionID

	return actionID, nil
}

// GetAndStoreActionIDForProposeSetStatusFromKc fetches the action ID for SetStatus by using the stored batch. Stores the action ID and returns it
func (executor *bridgeExecutor) GetAndStoreActionIDForProposeSetStatusFromKc(ctx context.Context) (uint64, error) {
	if executor.batch == nil {
		return InvalidActionID, ErrNilBatch
	}

	actionID, err := executor.kcClient.GetActionIDForSetStatusOnPendingTransfer(ctx, executor.batch)
	if err != nil {
		return InvalidActionID, err
	}

	executor.actionID = actionID

	return actionID, nil
}

// GetStoredActionID returns the stored action ID
func (executor *bridgeExecutor) GetStoredActionID() uint64 {
	return executor.actionID
}

// WasTransferProposedOnKc checks if the transfer was proposed on Kc
func (executor *bridgeExecutor) WasTransferProposedOnKc(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, ErrNilBatch
	}

	return executor.kcClient.WasProposedTransfer(ctx, executor.batch)
}

// ProposeTransferOnKc propose the transfer on Kc
func (executor *bridgeExecutor) ProposeTransferOnKc(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	hash, err := executor.kcClient.ProposeTransfer(ctx, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("proposed transfer", "hash", hash,
		"batch ID", executor.batch.ID, "action ID", executor.actionID)

	return nil
}

// ProcessMaxRetriesOnWasTransferProposedOnKc checks if the retries on Kc were reached and increments the counter
func (executor *bridgeExecutor) ProcessMaxRetriesOnWasTransferProposedOnKc() bool {
	if executor.retriesOnWasProposed < executor.maxRetriesOnWasProposed {
		executor.retriesOnWasProposed++
		return false
	}

	return true
}

// ResetRetriesOnWasTransferProposedOnKc resets the number of retries on was transfer proposed
func (executor *bridgeExecutor) ResetRetriesOnWasTransferProposedOnKc() {
	executor.retriesOnWasProposed = 0
}

// WasSetStatusProposedOnKc checks if set status was proposed on Kc
func (executor *bridgeExecutor) WasSetStatusProposedOnKc(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, ErrNilBatch
	}

	return executor.kcClient.WasProposedSetStatus(ctx, executor.batch)
}

// ProposeSetStatusOnKc propose set status on Kc
func (executor *bridgeExecutor) ProposeSetStatusOnKc(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	hash, err := executor.kcClient.ProposeSetStatus(ctx, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("proposed set status", "hash", hash,
		"batch ID", executor.batch.ID)

	return nil
}

// WasActionSignedOnKc returns true if the current relayer already signed the action
func (executor *bridgeExecutor) WasActionSignedOnKc(ctx context.Context) (bool, error) {
	return executor.kcClient.WasSigned(ctx, executor.actionID)
}

// SignActionOnKc calls the Kc client to generate and send the signature
func (executor *bridgeExecutor) SignActionOnKc(ctx context.Context) error {
	hash, err := executor.kcClient.Sign(ctx, executor.actionID)
	if err != nil {
		return err
	}

	executor.log.Info("signed proposed transfer", "hash", hash, "action ID", executor.actionID)

	return nil
}

// ProcessQuorumReachedOnKc returns true if the proposed transfer reached the set quorum
func (executor *bridgeExecutor) ProcessQuorumReachedOnKc(ctx context.Context) (bool, error) {
	return executor.kcClient.QuorumReached(ctx, executor.actionID)
}

// WaitForTransferConfirmation waits for the confirmation of a transfer
func (executor *bridgeExecutor) WaitForTransferConfirmation(ctx context.Context) {
	wasPerformed := false
	for i := 0; i < splits && !wasPerformed; i++ {
		if executor.waitWithContextSucceeded(ctx) {
			wasPerformed, _ = executor.WasTransferPerformedOnEthereum(ctx)
		}
	}
}

// WaitAndReturnFinalBatchStatuses waits for the statuses to be final
func (executor *bridgeExecutor) WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte {
	for i := 0; i < splits; i++ {
		if !executor.waitWithContextSucceeded(ctx) {
			return nil
		}

		statuses, err := executor.GetBatchStatusesFromEthereum(ctx)
		if err != nil {
			executor.log.Debug("got message while fetching batch statuses", "message", err)
			continue
		}
		if len(statuses) == 0 {
			executor.log.Debug("no status available")
			continue
		}

		executor.log.Debug("bridgeExecutor.WaitAndReturnFinalBatchStatuses", "statuses", statuses)
		return statuses
	}

	return nil
}

func (executor *bridgeExecutor) waitWithContextSucceeded(ctx context.Context) bool {
	timer := time.NewTimer(executor.timeForWaitOnEthereum / splits)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		executor.log.Debug("closing due to context expiration")
		return false
	case <-timer.C:
		return true
	}
}

// GetBatchStatusesFromEthereum gets statuses for the batch
func (executor *bridgeExecutor) GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error) {
	if executor.batch == nil {
		return nil, ErrNilBatch
	}

	statuses, err := executor.ethereumClient.GetTransactionsStatuses(ctx, executor.batch.ID)
	if err != nil {
		return nil, err
	}

	return statuses, nil
}

// WasActionPerformedOnKc returns true if the action was already performed
func (executor *bridgeExecutor) WasActionPerformedOnKc(ctx context.Context) (bool, error) {
	return executor.kcClient.WasExecuted(ctx, executor.actionID)
}

// PerformActionOnKc sends the perform-action transaction on the Klever Blockchain chain
func (executor *bridgeExecutor) PerformActionOnKc(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	hash, err := executor.kcClient.PerformAction(ctx, executor.actionID, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("sent perform action transaction", "hash", hash,
		"batch ID", executor.batch.ID, "action ID", executor.actionID)

	return nil
}

// ResolveNewDepositsStatuses resolves the new deposits statuses for batch
func (executor *bridgeExecutor) ResolveNewDepositsStatuses(numDeposits uint64) {
	executor.batch.ResolveNewDeposits(int(numDeposits))
}

// ProcessMaxQuorumRetriesOnKc checks if the retries on Klever Blockchain were reached and increments the counter
func (executor *bridgeExecutor) ProcessMaxQuorumRetriesOnKc() bool {
	if executor.quorumRetriesOnKc < executor.maxQuorumRetriesOnKc {
		executor.quorumRetriesOnKc++
		return false
	}

	return true
}

// ResetRetriesCountOnKc resets the number of retries on Kc
func (executor *bridgeExecutor) ResetRetriesCountOnKc() {
	executor.quorumRetriesOnKc = 0
}

// GetAndStoreBatchFromEthereum fetches and stores the batch from the ethereum client
func (executor *bridgeExecutor) GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error {
	batch, isFinal, err := executor.ethereumClient.GetBatch(ctx, nonce)
	if err != nil {
		return err
	}

	isBatchInvalid := batch.ID != nonce || len(batch.Deposits) == 0 || !isFinal
	if isBatchInvalid {
		return fmt.Errorf("%w, requested nonce: %d, fetched nonce: %d, num deposits: %d, isFinal: %v",
			ErrFinalBatchNotFound, nonce, batch.ID, len(batch.Deposits), isFinal)
	}

	batch, err = executor.addBatchSCMetadata(ctx, batch)
	if err != nil {
		return err
	}
	executor.batch = batch

	return nil
}

// addBatchSCMetadata fetches the logs containing sc calls metadata for the current batch
func (executor *bridgeExecutor) addBatchSCMetadata(ctx context.Context, transfers *bridgeCore.TransferBatch) (*bridgeCore.TransferBatch, error) {
	if transfers == nil {
		return nil, ErrNilBatch
	}

	events, err := executor.ethereumClient.GetBatchSCMetadata(ctx, transfers.ID, int64(transfers.BlockNumber))
	if err != nil {
		return nil, err
	}

	for i, t := range transfers.Deposits {
		transfers.Deposits[i] = executor.addMetadataToTransfer(t, events)
	}

	return transfers, nil
}

func (executor *bridgeExecutor) addMetadataToTransfer(transfer *bridgeCore.DepositTransfer, events []*contract.ERC20SafeERC20SCDeposit) *bridgeCore.DepositTransfer {
	for _, event := range events {
		if event.DepositNonce.Uint64() == transfer.Nonce {
			processData(transfer, event.CallData)
			return transfer
		}
	}

	transfer.Data = []byte{bridgeCore.MissingDataProtocolMarker}
	transfer.DisplayableData = ""

	return transfer
}

func processData(transfer *bridgeCore.DepositTransfer, buff []byte) {
	transfer.Data = buff
	dataLen := len(transfer.Data)
	if dataLen == 0 {
		transfer.Data = []byte{bridgeCore.MissingDataProtocolMarker}
		transfer.DisplayableData = ""
		return
	}
	// this check is optional, but brings an optimisation to reduce the gas used in case of a bad callData
	if dataLen == 1 && buff[0] == bridgeCore.MissingDataProtocolMarker {
		return
	}

	// we have a data field, add the marker & the correct length
	transfer.DisplayableData = hex.EncodeToString(transfer.Data)
	buff32 := make([]byte, bridgeCore.Uint32ArgBytes)
	binary.BigEndian.PutUint32(buff32, uint32(dataLen))

	prefix := append([]byte{bridgeCore.DataPresentProtocolMarker}, buff32...)

	transfer.Data = append(prefix, transfer.Data...)
}

// WasTransferPerformedOnEthereum returns true if the batch was performed on Ethereum
func (executor *bridgeExecutor) WasTransferPerformedOnEthereum(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, ErrNilBatch
	}

	return executor.ethereumClient.WasExecuted(ctx, executor.batch.ID)
}

// SignTransferOnEthereum generates the message hash for batch and broadcast the signature
func (executor *bridgeExecutor) SignTransferOnEthereum() error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	argLists := batchProcessor.ExtractListKlvToEth(executor.batch)
	hash, err := executor.ethereumClient.GenerateMessageHash(argLists, executor.batch.ID)
	if err != nil {
		return err
	}

	executor.log.Info("generated message hash on Ethereum", "hash", hash,
		"batch ID", executor.batch.ID)

	executor.msgHash = hash
	executor.ethereumClient.BroadcastSignatureForMessageHash(hash)
	return nil
}

// PerformTransferOnEthereum transfers a batch to Ethereum
func (executor *bridgeExecutor) PerformTransferOnEthereum(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	quorumSize, err := executor.ethereumClient.GetQuorumSize(ctx)
	if err != nil {
		return err
	}

	executor.log.Debug("fetched quorum size", "quorum", quorumSize.Int64())

	argLists := batchProcessor.ExtractListKlvToEth(executor.batch)

	executor.log.Info("executing transfer " + executor.batch.String())

	hash, err := executor.ethereumClient.ExecuteTransfer(ctx, executor.msgHash, argLists, executor.batch.ID, int(quorumSize.Int64()))
	if err != nil {
		return err
	}

	executor.log.Info("sent execute transfer", "hash", hash,
		"batch ID", executor.batch.ID)

	return nil
}

func (executor *bridgeExecutor) checkCumulatedTransfers(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	for i, ethToken := range ethTokens {
		err := executor.balanceValidator.CheckToken(ctx, ethToken, kdaTokens[i], amounts[i], direction)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckAvailableTokens checks the available balances
func (executor *bridgeExecutor) CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	ethTokens, kdaTokens, amounts = executor.getCumulatedTransfers(ethTokens, kdaTokens, amounts)

	return executor.checkCumulatedTransfers(ctx, ethTokens, kdaTokens, amounts, direction)
}

func (executor *bridgeExecutor) getCumulatedTransfers(ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int) ([]common.Address, [][]byte, []*big.Int) {
	cumulatedAmounts := make(map[common.Address]*big.Int)
	uniqueTokens := make([]common.Address, 0)
	uniqueConvertedTokens := make([][]byte, 0)

	for i, token := range ethTokens {
		existingValue, exists := cumulatedAmounts[token]
		if exists {
			existingValue.Add(existingValue, amounts[i])
			continue
		}

		cumulatedAmounts[token] = big.NewInt(0).Set(amounts[i]) // work on a new pointer
		uniqueTokens = append(uniqueTokens, token)
		uniqueConvertedTokens = append(uniqueConvertedTokens, kdaTokens[i])
	}

	finalAmounts := make([]*big.Int, len(uniqueTokens))
	for i, token := range uniqueTokens {
		finalAmounts[i] = cumulatedAmounts[token]
	}

	return uniqueTokens, uniqueConvertedTokens, finalAmounts
}

// ProcessQuorumReachedOnEthereum returns true if the proposed transfer reached the set quorum
func (executor *bridgeExecutor) ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error) {
	return executor.ethereumClient.IsQuorumReached(ctx, executor.msgHash)
}

// ProcessMaxQuorumRetriesOnEthereum checks if the retries on Ethereum were reached and increments the counter
func (executor *bridgeExecutor) ProcessMaxQuorumRetriesOnEthereum() bool {
	if executor.quorumRetriesOnEthereum < executor.maxQuorumRetriesOnEthereum {
		executor.quorumRetriesOnEthereum++
		return false
	}

	return true
}

// ResetRetriesCountOnEthereum resets the number of retries on Ethereum
func (executor *bridgeExecutor) ResetRetriesCountOnEthereum() {
	executor.quorumRetriesOnEthereum = 0
}

// ClearStoredP2PSignaturesForEthereum deletes all stored P2P signatures used for Ethereum client
func (executor *bridgeExecutor) ClearStoredP2PSignaturesForEthereum() {
	executor.sigsHolder.ClearStoredSignatures()
	executor.log.Info("cleared stored P2P signatures")
}

// CheckKcClientAvailability trigger a self availability check for the Klever Blockchain client
func (executor *bridgeExecutor) CheckKcClientAvailability(ctx context.Context) error {
	return executor.kcClient.CheckClientAvailability(ctx)
}

// CheckEthereumClientAvailability trigger a self availability check for the Ethereum client
func (executor *bridgeExecutor) CheckEthereumClientAvailability(ctx context.Context) error {
	return executor.ethereumClient.CheckClientAvailability(ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *bridgeExecutor) IsInterfaceNil() bool {
	return executor == nil
}
