package klever

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/builders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/errors"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	okCodeAfterExecution                                      = "Ok"
	internalError                                             = "internal error"
	getCurrentTxBatchFuncName                                 = "getCurrentTxBatch"
	getBatchFuncName                                          = "getBatch"
	wasTransferActionProposedFuncName                         = "wasTransferActionProposed"
	wasActionExecutedFuncName                                 = "wasActionExecuted"
	getActionIdForTransferBatchFuncName                       = "getActionIdForTransferBatch"
	wasSetCurrentTransactionBatchStatusActionProposedFuncName = "wasSetCurrentTransactionBatchStatusActionProposed"
	getStatusesAfterExecutionFuncName                         = "getStatusesAfterExecution"
	getActionIdForSetCurrentTransactionBatchStatusFuncName    = "getActionIdForSetCurrentTransactionBatchStatus"
	getTokenIdForErc20AddressFuncName                         = "getTokenIdForErc20Address"
	getErc20AddressForTokenIdFuncName                         = "getErc20AddressForTokenId"
	quorumReachedFuncName                                     = "quorumReached"
	getLastExecutedEthBatchIdFuncName                         = "getLastExecutedEthBatchId"
	getLastExecutedEthTxId                                    = "getLastExecutedEthTxId"
	signedFuncName                                            = "signed"
	getAllStakedRelayersFuncName                              = "getAllStakedRelayers"
	isPausedFuncName                                          = "isPaused"
	isMintBurnTokenFuncName                                   = "isMintBurnToken"
	isNativeTokenFuncName                                     = "isNativeToken"
	getTotalBalances                                          = "getTotalBalances"
	getMintBalances                                           = "getMintBalances"
	getBurnBalances                                           = "getBurnBalances"
	getAllKnownTokens                                         = "getAllKnownTokens"
	getLastBatchId                                            = "getLastBatchId"
)

// ArgsklvClientDataGetter is the arguments DTO used in the NewklvClientDataGetter constructor
type ArgsKLVClientDataGetter struct {
	MultisigContractAddress address.Address
	SafeContractAddress     address.Address
	RelayerAddress          address.Address
	Proxy                   proxy.Proxy
	Log                     logger.Logger
}

type klvClientDataGetter struct {
	multisigContractAddress       address.Address
	safeContractAddress           address.Address
	bech32MultisigContractAddress string
	relayerAddress                address.Address
	proxy                         proxy.Proxy
	log                           logger.Logger
}

// NewklvClientDataGetter creates a new instance of the dataGetter type
func NewKLVClientDataGetter(args ArgsKLVClientDataGetter) (*klvClientDataGetter, error) {
	if check.IfNil(args.Log) {
		return nil, errNilLogger
	}
	if check.IfNil(args.Proxy) {
		return nil, errNilProxy
	}
	if check.IfNil(args.RelayerAddress) {
		return nil, fmt.Errorf("%w for the RelayerAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.MultisigContractAddress) {
		return nil, fmt.Errorf("%w for the MultisigContractAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.SafeContractAddress) {
		return nil, fmt.Errorf("%w for the SafeContractAddress argument", errNilAddressHandler)
	}
	bech32Address := args.MultisigContractAddress.Bech32()

	return &klvClientDataGetter{
		multisigContractAddress:       args.MultisigContractAddress,
		safeContractAddress:           args.SafeContractAddress,
		bech32MultisigContractAddress: bech32Address,
		relayerAddress:                args.RelayerAddress,
		proxy:                         args.Proxy,
		log:                           args.Log,
	}, nil
}

// ExecuteQueryReturningBytes will try to execute the provided query and return the result as slice of byte slices
func (dataGetter *klvClientDataGetter) ExecuteQueryReturningBytes(ctx context.Context, request *models.VmValueRequest) ([][]byte, error) {
	if request == nil {
		return nil, errNilRequest
	}

	response, err := dataGetter.proxy.ExecuteVMQuery(ctx, request)
	if err != nil {
		dataGetter.log.Error("got error on VMQuery", "FuncName", request.FuncName,
			"Args", request.Args, "SC address", request.Address, "Caller", request.CallerAddr, "error", err)
		return nil, err
	}
	dataGetter.log.Debug("executed VMQuery", "FuncName", request.FuncName,
		"Args", request.Args, "SC address", request.Address, "Caller", request.CallerAddr,
		"response.ReturnCode", response.Data.ReturnCode,
		"response.ReturnData", fmt.Sprintf("%+v", response.Data.ReturnData))
	if response.Data.ReturnCode != okCodeAfterExecution {
		return nil, errors.NewQueryResponseError(
			response.Data.ReturnCode,
			response.Data.ReturnMessage,
			request.FuncName,
			request.Address,
			request.Args...,
		)
	}
	return response.Data.ReturnData, nil
}

// GetCurrentNonce will get from the shard containing the multisig contract the latest block's nonce
func (dataGetter *klvClientDataGetter) GetCurrentNonce(ctx context.Context) (uint64, error) {
	nodeStatus, err := dataGetter.proxy.GetNetworkStatus(ctx)
	if err != nil {
		return 0, err
	}
	if nodeStatus == nil {
		return 0, errNilNodeStatusResponse
	}

	return nodeStatus.Nonce, nil
}

// ExecuteQueryReturningBool will try to execute the provided query and return the result as bool
func (dataGetter *klvClientDataGetter) ExecuteQueryReturningBool(ctx context.Context, request *models.VmValueRequest) (bool, error) {
	response, err := dataGetter.ExecuteQueryReturningBytes(ctx, request)
	if err != nil {
		return false, err
	}

	if len(response) == 0 {
		return false, nil
	}

	return dataGetter.parseBool(response[0], request.FuncName, request.Address, request.Args...)
}

func (dataGetter *klvClientDataGetter) parseBool(buff []byte, funcName string, address string, args ...string) (bool, error) {
	if len(buff) == 0 {
		return false, nil
	}

	result, err := strconv.ParseBool(fmt.Sprintf("%d", buff[0]))
	if err != nil {
		return false, errors.NewQueryResponseError(
			internalError,
			fmt.Sprintf("error converting the received bytes to bool, %s", err.Error()),
			funcName,
			address,
			args...,
		)
	}

	return result, nil
}

// ExecuteQueryReturningUint64 will try to execute the provided query and return the result as uint64
func (dataGetter *klvClientDataGetter) ExecuteQueryReturningUint64(ctx context.Context, request *models.VmValueRequest) (uint64, error) {
	response, err := dataGetter.ExecuteQueryReturningBytes(ctx, request)
	if err != nil {
		return 0, err
	}

	if len(response) == 0 {
		return 0, nil
	}
	if len(response[0]) == 0 {
		return 0, nil
	}

	num, err := parseUInt64FromByteSlice(response[0])
	if err != nil {
		return 0, errors.NewQueryResponseError(
			internalError,
			err.Error(),
			request.FuncName,
			request.Address,
			request.Args...,
		)
	}

	return num, nil
}

// ExecuteQueryReturningBigInt will try to execute the provided query and return the result as big.Int
func (dataGetter *klvClientDataGetter) ExecuteQueryReturningBigInt(ctx context.Context, request *models.VmValueRequest) (*big.Int, error) {
	response, err := dataGetter.ExecuteQueryReturningBytes(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return big.NewInt(0), nil
	}
	if len(response[0]) == 0 {
		return big.NewInt(0), nil
	}

	num := big.NewInt(0).SetBytes(response[0])
	return num, nil
}

func parseUInt64FromByteSlice(bytes []byte) (uint64, error) {
	num := big.NewInt(0).SetBytes(bytes)
	if !num.IsUint64() {
		return 0, errNotUint64Bytes
	}

	return num.Uint64(), nil
}

func (dataGetter *klvClientDataGetter) executeQueryFromBuilder(ctx context.Context, builder builders.VMQueryBuilder) ([][]byte, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return nil, err
	}

	return dataGetter.ExecuteQueryReturningBytes(ctx, vmValuesRequest)
}

func (dataGetter *klvClientDataGetter) executeQueryUint64FromBuilder(ctx context.Context, builder builders.VMQueryBuilder) (uint64, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return 0, err
	}

	return dataGetter.ExecuteQueryReturningUint64(ctx, vmValuesRequest)
}

func (dataGetter *klvClientDataGetter) executeQueryBigIntFromBuilder(ctx context.Context, builder builders.VMQueryBuilder) (*big.Int, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return nil, err
	}

	return dataGetter.ExecuteQueryReturningBigInt(ctx, vmValuesRequest)
}

func (dataGetter *klvClientDataGetter) executeQueryBoolFromBuilder(ctx context.Context, builder builders.VMQueryBuilder) (bool, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return false, err
	}

	return dataGetter.ExecuteQueryReturningBool(ctx, vmValuesRequest)
}

func (dataGetter *klvClientDataGetter) createMultisigDefaultVmQueryBuilder() builders.VMQueryBuilder {
	return builders.NewVMQueryBuilder().Address(dataGetter.multisigContractAddress).CallerAddress(dataGetter.relayerAddress)
}

func (dataGetter *klvClientDataGetter) createSafeDefaultVmQueryBuilder() builders.VMQueryBuilder {
	return builders.NewVMQueryBuilder().Address(dataGetter.safeContractAddress).CallerAddress(dataGetter.relayerAddress)
}

// GetCurrentBatchAsDataBytes will assemble a builder and query the proxy for the current pending batch
func (dataGetter *klvClientDataGetter) GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getCurrentTxBatchFuncName)

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// GetBatchAsDataBytes will assemble a builder and query the proxy for the batch info
func (dataGetter *klvClientDataGetter) GetBatchAsDataBytes(ctx context.Context, batchID uint64) ([][]byte, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getBatchFuncName)
	builder.ArgInt64(int64(batchID))

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// GetTokenIdForErc20Address will assemble a builder and query the proxy for a token id given a specific erc20 address
func (dataGetter *klvClientDataGetter) GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getTokenIdForErc20AddressFuncName)
	builder.ArgBytes(erc20Address)

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// GetERC20AddressForTokenId will assemble a builder and query the proxy for an erc20 address given a specific token id
func (dataGetter *klvClientDataGetter) GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getErc20AddressForTokenIdFuncName)
	builder.ArgBytes(tokenId)
	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// WasProposedTransfer returns true if the transfer action proposed was triggered
func (dataGetter *klvClientDataGetter) WasProposedTransfer(ctx context.Context, batch *bridgeCore.TransferBatch) (bool, error) {
	if batch == nil {
		return false, clients.ErrNilBatch
	}

	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(wasTransferActionProposedFuncName).ArgInt64(int64(batch.ID))
	dataGetter.addBatchInfo(builder, batch)

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// WasExecuted returns true if the provided actionID was executed or not
func (dataGetter *klvClientDataGetter) WasExecuted(ctx context.Context, actionID uint64) (bool, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(wasActionExecutedFuncName).ArgInt64(int64(actionID))

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetActionIDForProposeTransfer returns the action ID for the proposed transfer operation
func (dataGetter *klvClientDataGetter) GetActionIDForProposeTransfer(ctx context.Context, batch *bridgeCore.TransferBatch) (uint64, error) {
	if batch == nil {
		return 0, clients.ErrNilBatch
	}

	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getActionIdForTransferBatchFuncName).ArgInt64(int64(batch.ID))
	dataGetter.addBatchInfo(builder, batch)

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// WasProposedSetStatus returns true if the proposed set status was triggered
func (dataGetter *klvClientDataGetter) WasProposedSetStatus(ctx context.Context, batch *bridgeCore.TransferBatch) (bool, error) {
	if batch == nil {
		return false, clients.ErrNilBatch
	}

	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(wasSetCurrentTransactionBatchStatusActionProposedFuncName).ArgInt64(int64(batch.ID))
	for _, stat := range batch.Statuses {
		builder.ArgBytes([]byte{stat})
	}

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetTransactionsStatuses will return the transactions statuses from the batch ID
func (dataGetter *klvClientDataGetter) GetTransactionsStatuses(ctx context.Context, batchID uint64) ([]byte, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getStatusesAfterExecutionFuncName).ArgInt64(int64(batchID))

	values, err := dataGetter.executeQueryFromBuilder(ctx, builder)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%w for batch ID %v", errNoStatusForBatchID, batchID)
	}

	isFinished, err := dataGetter.parseBool(values[0], getStatusesAfterExecutionFuncName, dataGetter.bech32MultisigContractAddress)
	if err != nil {
		return nil, err
	}
	if !isFinished {
		return nil, fmt.Errorf("%w for batch ID %v", errBatchNotFinished, batchID)
	}

	results := make([]byte, len(values)-1)
	for i := 1; i < len(values); i++ {
		results[i-1], err = getStatusFromBuff(values[i])
		if err != nil {
			return nil, fmt.Errorf("%w for result index %d", err, i-1)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%w status is finished, no results are given", errMalformedBatchResponse)
	}

	return results, nil
}

// GetActionIDForSetStatusOnPendingTransfer returns the action ID for setting the status on the pending transfer batch
func (dataGetter *klvClientDataGetter) GetActionIDForSetStatusOnPendingTransfer(ctx context.Context, batch *bridgeCore.TransferBatch) (uint64, error) {
	if batch == nil {
		return 0, clients.ErrNilBatch
	}

	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getActionIdForSetCurrentTransactionBatchStatusFuncName).ArgInt64(int64(batch.ID))
	for _, stat := range batch.Statuses {
		builder.ArgBytes([]byte{stat})
	}

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// QuorumReached returns true if the provided action ID reached the set quorum
func (dataGetter *klvClientDataGetter) QuorumReached(ctx context.Context, actionID uint64) (bool, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(quorumReachedFuncName).ArgInt64(int64(actionID))

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetLastExecutedEthBatchID returns the last executed Ethereum batch ID
func (dataGetter *klvClientDataGetter) GetLastExecutedEthBatchID(ctx context.Context) (uint64, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder().Function(getLastExecutedEthBatchIdFuncName)

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// GetLastExecutedEthTxID returns the last executed Ethereum deposit ID
func (dataGetter *klvClientDataGetter) GetLastExecutedEthTxID(ctx context.Context) (uint64, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder().Function(getLastExecutedEthTxId)

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// WasSigned returns true if the action was already signed by the current relayer
func (dataGetter *klvClientDataGetter) WasSigned(ctx context.Context, actionID uint64) (bool, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(signedFuncName).ArgAddress(dataGetter.relayerAddress).ArgInt64(int64(actionID))

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetAllStakedRelayers returns all staked relayers defined in Klever Blockchain SC
func (dataGetter *klvClientDataGetter) GetAllStakedRelayers(ctx context.Context) ([][]byte, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(getAllStakedRelayersFuncName)

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// IsPaused returns true if the multisig contract is paused
func (dataGetter *klvClientDataGetter) IsPaused(ctx context.Context) (bool, error) {
	builder := dataGetter.createMultisigDefaultVmQueryBuilder()
	builder.Function(isPausedFuncName)

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

func (dataGetter *klvClientDataGetter) isMintBurnToken(ctx context.Context, token []byte) (bool, error) {
	builder := dataGetter.createSafeDefaultVmQueryBuilder()
	builder.Function(isMintBurnTokenFuncName).ArgBytes(token)

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// isNativeToken returns true if the token is native
func (dataGetter *klvClientDataGetter) isNativeToken(ctx context.Context, token []byte) (bool, error) {
	builder := dataGetter.createSafeDefaultVmQueryBuilder()
	builder.Function(isNativeTokenFuncName).ArgBytes(token)

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

func (dataGetter *klvClientDataGetter) getTotalBalances(ctx context.Context, token []byte) (*big.Int, error) {
	builder := dataGetter.createSafeDefaultVmQueryBuilder()
	builder.Function(getTotalBalances).ArgBytes(token)

	return dataGetter.executeQueryBigIntFromBuilder(ctx, builder)
}

func (dataGetter *klvClientDataGetter) getMintBalances(ctx context.Context, token []byte) (*big.Int, error) {
	builder := dataGetter.createSafeDefaultVmQueryBuilder()
	builder.Function(getMintBalances).ArgBytes(token)

	return dataGetter.executeQueryBigIntFromBuilder(ctx, builder)
}

func (dataGetter *klvClientDataGetter) getBurnBalances(ctx context.Context, token []byte) (*big.Int, error) {
	builder := dataGetter.createSafeDefaultVmQueryBuilder()
	builder.Function(getBurnBalances).ArgBytes(token)

	return dataGetter.executeQueryBigIntFromBuilder(ctx, builder)
}

func (dataGetter *klvClientDataGetter) addBatchInfo(builder builders.VMQueryBuilder, batch *bridgeCore.TransferBatch) {
	for _, dt := range batch.Deposits {
		builder.ArgBytes(dt.FromBytes).
			ArgBytes(dt.ToBytes).
			ArgBytes(dt.DestinationTokenBytes).
			ArgBigInt(dt.Amount).
			ArgInt64(int64(dt.Nonce)).
			ArgBytes(dt.Data)
	}
}

func getStatusFromBuff(buff []byte) (byte, error) {
	if len(buff) == 0 {
		return 0, errMalformedBatchResponse
	}

	return buff[len(buff)-1], nil
}

// GetAllKnownTokens returns all registered tokens
func (dataGetter *klvClientDataGetter) GetAllKnownTokens(ctx context.Context) ([][]byte, error) {
	builder := dataGetter.createSafeDefaultVmQueryBuilder()
	builder.Function(getAllKnownTokens)

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// GetLastKCBatchID returns the highest batch ID the safe contract reached. This might be a WIP batch that is not executable yet
func (dataGetter *klvClientDataGetter) GetLastKCBatchID(ctx context.Context) (uint64, error) {
	builder := dataGetter.createSafeDefaultVmQueryBuilder()
	builder.Function(getLastBatchId)

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// IsInterfaceNil returns true if there is no value under the interface
func (dataGetter *klvClientDataGetter) IsInterfaceNil() bool {
	return dataGetter == nil
}
