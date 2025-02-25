package multiversx

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/builders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/errors"
	"github.com/klever-io/klv-bridge-eth-go/parsers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	getPendingTransactionsFunction = "getPendingTransactions"
	okCodeAfterExecution           = "ok"
	scProxyCallFunction            = "execute"
	minCheckValues                 = 1
	transactionNotFoundErrString   = "transaction not found"
	minGasToExecuteSCCalls         = 2010000 // the absolut minimum gas limit to do a SC call
	contractMaxGasLimit            = 249999999
)

// ArgsScCallExecutor represents the DTO struct for creating a new instance of type scCallExecutor
type ArgsScCallExecutor struct {
	ScProxyBech32Address            string
	Proxy                           proxy.Proxy
	Codec                           Codec
	Filter                          ScCallsExecuteFilter
	Log                             logger.Logger
	ExtraGasToExecute               uint64
	MaxGasLimitToUse                uint64
	GasLimitForOutOfGasTransactions uint64
	NonceTxHandler                  NonceTransactionsHandler
	PrivateKey                      crypto.PrivateKey
	SingleSigner                    crypto.SingleSigner
	TransactionChecks               config.TransactionChecksConfig
	CloseAppChan                    chan struct{}
}

type scCallExecutor struct {
	scProxyBech32Address            string
	proxy                           proxy.Proxy
	codec                           Codec
	filter                          ScCallsExecuteFilter
	log                             logger.Logger
	extraGasToExecute               uint64
	maxGasLimitToUse                uint64
	gasLimitForOutOfGasTransactions uint64
	nonceTxHandler                  NonceTransactionsHandler
	privateKey                      crypto.PrivateKey
	singleSigner                    crypto.SingleSigner
	senderAddress                   address.Address
	numSentTransactions             uint32
	checkTransactionResults         bool
	timeBetweenChecks               time.Duration
	executionTimeout                time.Duration
	closeAppOnError                 bool
	extraDelayOnError               time.Duration
	closeAppChan                    chan struct{}
}

// NewScCallExecutor creates a new instance of type scCallExecutor
func NewScCallExecutor(args ArgsScCallExecutor) (*scCallExecutor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	publicKey := args.PrivateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	senderAddress, err := address.NewAddressFromBytes(publicKeyBytes)
	if err != nil {
		return nil, err
	}

	return &scCallExecutor{
		scProxyBech32Address:            args.ScProxyBech32Address,
		proxy:                           args.Proxy,
		codec:                           args.Codec,
		filter:                          args.Filter,
		log:                             args.Log,
		extraGasToExecute:               args.ExtraGasToExecute,
		maxGasLimitToUse:                args.MaxGasLimitToUse,
		gasLimitForOutOfGasTransactions: args.GasLimitForOutOfGasTransactions,
		nonceTxHandler:                  args.NonceTxHandler,
		privateKey:                      args.PrivateKey,
		singleSigner:                    args.SingleSigner,
		senderAddress:                   senderAddress,
		checkTransactionResults:         args.TransactionChecks.CheckTransactionResults,
		timeBetweenChecks:               time.Second * time.Duration(args.TransactionChecks.TimeInSecondsBetweenChecks),
		executionTimeout:                time.Second * time.Duration(args.TransactionChecks.ExecutionTimeoutInSeconds),
		closeAppOnError:                 args.TransactionChecks.CloseAppOnError,
		extraDelayOnError:               time.Second * time.Duration(args.TransactionChecks.ExtraDelayInSecondsOnError),
		closeAppChan:                    args.CloseAppChan,
	}, nil
}

func checkArgs(args ArgsScCallExecutor) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}
	if check.IfNil(args.Codec) {
		return errNilCodec
	}
	if check.IfNil(args.Filter) {
		return errNilFilter
	}
	if check.IfNil(args.Log) {
		return errNilLogger
	}
	if check.IfNil(args.NonceTxHandler) {
		return errNilNonceTxHandler
	}
	if check.IfNil(args.PrivateKey) {
		return errNilPrivateKey
	}
	if check.IfNil(args.SingleSigner) {
		return errNilSingleSigner
	}
	if args.MaxGasLimitToUse < minGasToExecuteSCCalls {
		return fmt.Errorf("%w for MaxGasLimitToUse: provided: %d, absolute minimum required: %d", errGasLimitIsLessThanAbsoluteMinimum, args.MaxGasLimitToUse, minGasToExecuteSCCalls)
	}
	if args.GasLimitForOutOfGasTransactions < minGasToExecuteSCCalls {
		return fmt.Errorf("%w for GasLimitForOutOfGasTransactions: provided: %d, absolute minimum required: %d", errGasLimitIsLessThanAbsoluteMinimum, args.GasLimitForOutOfGasTransactions, minGasToExecuteSCCalls)
	}
	err := checkTransactionChecksConfig(args)
	if err != nil {
		return err
	}

	_, err = address.NewAddress(args.ScProxyBech32Address)

	return err
}

func checkTransactionChecksConfig(args ArgsScCallExecutor) error {
	if !args.TransactionChecks.CheckTransactionResults {
		args.Log.Warn("transaction checks are disabled! This can lead to funds being drained in case of a repetitive error")
		return nil
	}

	if args.TransactionChecks.TimeInSecondsBetweenChecks < minCheckValues {
		return fmt.Errorf("%w for TransactionChecks.TimeInSecondsBetweenChecks, minimum: %d, got: %d",
			errInvalidValue, minCheckValues, args.TransactionChecks.TimeInSecondsBetweenChecks)
	}
	if args.TransactionChecks.ExecutionTimeoutInSeconds < minCheckValues {
		return fmt.Errorf("%w for TransactionChecks.ExecutionTimeoutInSeconds, minimum: %d, got: %d",
			errInvalidValue, minCheckValues, args.TransactionChecks.ExecutionTimeoutInSeconds)
	}
	if args.CloseAppChan == nil && args.TransactionChecks.CloseAppOnError {
		return fmt.Errorf("%w while the TransactionChecks.CloseAppOnError is set to true", errNilCloseAppChannel)
	}

	return nil
}

// Execute will execute one step: get all pending operations, call the filter and send execution transactions
func (executor *scCallExecutor) Execute(ctx context.Context) error {
	pendingOperations, err := executor.getPendingOperations(ctx)
	if err != nil {
		return err
	}

	filteredPendingOperations := executor.filterOperations(pendingOperations)

	return executor.executeOperations(ctx, filteredPendingOperations)
}

func (executor *scCallExecutor) getPendingOperations(ctx context.Context) (map[uint64]parsers.ProxySCCompleteCallData, error) {
	request := &models.VmValueRequest{
		Address:  executor.scProxyBech32Address,
		FuncName: getPendingTransactionsFunction,
	}

	response, err := executor.proxy.ExecuteVMQuery(ctx, request)
	if err != nil {
		executor.log.Error("got error on VMQuery", "FuncName", request.FuncName,
			"Args", request.Args, "SC address", request.Address, "Caller", request.CallerAddr, "error", err)
		return nil, err
	}
	if response.Data.ReturnCode != okCodeAfterExecution {
		return nil, errors.NewQueryResponseError(
			response.Data.ReturnCode,
			response.Data.ReturnMessage,
			request.FuncName,
			request.Address,
			request.Args...,
		)
	}

	return executor.parseResponse(response)
}

func (executor *scCallExecutor) parseResponse(response *models.VmValuesResponseData) (map[uint64]parsers.ProxySCCompleteCallData, error) {
	numResponseLines := len(response.Data.ReturnData)
	if numResponseLines%2 != 0 {
		return nil, fmt.Errorf("%w: expected an even number, got %d", errInvalidNumberOfResponseLines, numResponseLines)
	}

	result := make(map[uint64]parsers.ProxySCCompleteCallData, numResponseLines/2)

	for i := 0; i < numResponseLines; i += 2 {
		pendingOperationID := big.NewInt(0).SetBytes(response.Data.ReturnData[i])
		callData, err := executor.codec.DecodeProxySCCompleteCallData(response.Data.ReturnData[i+1])
		if err != nil {
			return nil, fmt.Errorf("%w for ReturnData at index %d", err, i+1)
		}

		result[pendingOperationID.Uint64()] = callData
	}

	return result, nil
}

func (executor *scCallExecutor) filterOperations(pendingOperations map[uint64]parsers.ProxySCCompleteCallData) map[uint64]parsers.ProxySCCompleteCallData {
	result := make(map[uint64]parsers.ProxySCCompleteCallData)
	for id, callData := range pendingOperations {
		if executor.filter.ShouldExecute(callData) {
			result[id] = callData
		}
	}

	executor.log.Debug("scCallExecutor.filterOperations", "input pending ops", len(pendingOperations), "result pending ops", len(result))

	return result
}

func (executor *scCallExecutor) executeOperations(ctx context.Context, pendingOperations map[uint64]parsers.ProxySCCompleteCallData) error {
	networkConfig, err := executor.proxy.GetNetworkConfig(ctx)
	if err != nil {
		return fmt.Errorf("%w while fetching network configs", err)
	}

	for id, callData := range pendingOperations {
		workingCtx, cancel := context.WithTimeout(ctx, executor.executionTimeout)

		executor.log.Debug("scCallExecutor.executeOperations", "executing ID", id, "call data", callData,
			"maximum timeout", executor.executionTimeout)
		err = executor.executeOperation(workingCtx, id, callData, networkConfig)
		cancel()

		if err != nil {
			return fmt.Errorf("%w for call data: %s", err, callData)
		}
	}

	return nil
}

func (executor *scCallExecutor) executeOperation(
	ctx context.Context,
	id uint64,
	callData parsers.ProxySCCompleteCallData,
	networkConfig *models.NetworkConfig,
) error {
	txBuilder := builders.NewTxDataBuilder()
	txBuilder.Function(scProxyCallFunction).ArgInt64(int64(id))

	dataBytes, err := txBuilder.ToDataBytes()
	if err != nil {
		return err
	}

	receiverAddr, err := address.NewAddress(executor.scProxyBech32Address)
	if err != nil {
		return err
	}

	tx := transaction.NewBaseTransaction(executor.senderAddress.Bytes(), 0, nil, 0, 0)
	err = tx.SetChainID([]byte(networkConfig.ChainID))
	if err != nil {
		return err
	}

	contractRequest := transaction.SmartContract{
		Address: receiverAddr.Bytes(),
	}

	txArgs := transaction.TXArgs{
		Type:     uint32(transaction.SmartContract_SCInvoke),
		Sender:   executor.senderAddress.Bytes(),
		Contract: json.RawMessage(contractRequest.String()),
		Data:     [][]byte{dataBytes},
	}

	err = tx.AddTransaction(txArgs)
	if err != nil {
		return err
	}

	to := callData.To.Bech32()
	if tx.GasLimit > contractMaxGasLimit {
		// the contract will refund this transaction, so we will use less gas to preserve funds
		executor.log.Warn("setting a lower gas limit for this transaction because it will be refunded",
			"computed gas limit", tx.GasLimit,
			"max allowed", executor.maxGasLimitToUse,
			"data", dataBytes,
			"from", callData.From.Hex(),
			"to", to,
			"token", callData.Token,
			"amount", callData.Amount,
			"nonce", callData.Nonce,
		)
		tx.GasLimit = executor.gasLimitForOutOfGasTransactions
	}

	if tx.GasLimit > executor.maxGasLimitToUse {
		executor.log.Warn("can not execute transaction because the provided gas limit on the SC call exceeds "+
			"the maximum gas limit allowance for this executor, WILL SKIP the execution",
			"computed gas limit", tx.GasLimit,
			"max allowed", executor.maxGasLimitToUse,
			"data", dataBytes,
			"from", callData.From.Hex(),
			"to", to,
			"token", callData.Token,
			"amount", callData.Amount,
			"nonce", callData.Nonce,
		)

		return nil
	}

	err = executor.nonceTxHandler.ApplyNonceAndGasPrice(ctx, executor.senderAddress, tx)
	if err != nil {
		return err
	}

	err = executor.signTransactionWithPrivateKey(tx)
	if err != nil {
		return err
	}

	hash, err := executor.nonceTxHandler.SendTransaction(ctx, tx)
	if err != nil {
		return err
	}

	executor.log.Info("scCallExecutor.executeOperation: sent transaction from executor",
		"hash", hash,
		"tx ID", id,
		"call data", callData.String(),
		"extra gas", executor.extraGasToExecute,
		"sender", executor.senderAddress.Bech32(),
		"to", to)

	atomic.AddUint32(&executor.numSentTransactions, 1)

	return executor.handleResults(ctx, hash)
}

func (executor *scCallExecutor) handleResults(ctx context.Context, hash string) error {
	if !executor.checkTransactionResults {
		return nil
	}

	err := executor.checkResultsUntilDone(ctx, hash)
	executor.waitForExtraDelay(ctx, err)
	return err
}

// signTransactionWithPrivateKey signs a transaction with the client's private key
func (executor *scCallExecutor) signTransactionWithPrivateKey(tx *transaction.Transaction) error {
	tx.Signature = [][]byte{}
	bytes, err := json.Marshal(&tx)
	if err != nil {
		return err
	}

	signature, err := executor.singleSigner.Sign(executor.privateKey, bytes)
	if err != nil {
		return err
	}

	tx.AddSignature(signature)

	return nil
}

func (executor *scCallExecutor) checkResultsUntilDone(ctx context.Context, hash string) error {
	timer := time.NewTimer(executor.timeBetweenChecks)
	defer timer.Stop()

	for {
		timer.Reset(executor.timeBetweenChecks)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			err, shouldStop := executor.checkResults(ctx, hash)
			if shouldStop {
				executor.handleError(ctx, err)
				return err
			}
		}
	}
}

func (executor *scCallExecutor) checkResults(ctx context.Context, hash string) (error, bool) {
	txStatus, err := executor.proxy.ProcessTransactionStatus(ctx, hash)
	if err != nil {
		if err.Error() == transactionNotFoundErrString {
			return nil, false
		}

		return err, true
	}

	if txStatus == transaction.Transaction_SUCCESS {
		return nil, true
	}

	executor.logFullTransaction(ctx, hash)
	return fmt.Errorf("%w for tx hash %s", errTransactionFailed, hash), true
}

func (executor *scCallExecutor) handleError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if !executor.closeAppOnError {
		return
	}

	go func() {
		// wait here until we could write in the close app chan
		// ... or the context expired (application might close)
		select {
		case <-ctx.Done():
		case executor.closeAppChan <- struct{}{}:
		}
	}()
}

func (executor *scCallExecutor) logFullTransaction(ctx context.Context, hash string) {
	txData, err := executor.proxy.GetTransactionInfoWithResults(ctx, hash)
	if err != nil {
		executor.log.Error("error getting the transaction for display", "error", err)
		return
	}

	txDataString, err := json.MarshalIndent(txData, "", "  ")
	if err != nil {
		executor.log.Error("error preparing transaction for display", "error", err)
		return
	}

	executor.log.Error("transaction failed", "hash", hash, "full transaction details", string(txDataString))
}

func (executor *scCallExecutor) waitForExtraDelay(ctx context.Context, err error) {
	if err == nil {
		return
	}

	timer := time.NewTimer(executor.extraDelayOnError)
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// GetNumSentTransaction returns the total sent transactions
func (executor *scCallExecutor) GetNumSentTransaction() uint32 {
	return atomic.LoadUint32(&executor.numSentTransactions)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *scCallExecutor) IsInterfaceNil() bool {
	return executor == nil
}
