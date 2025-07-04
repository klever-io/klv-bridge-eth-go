package klever

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	factoryHasher "github.com/klever-io/klever-go/crypto/hashing/factory"
	"github.com/klever-io/klever-go/tools/marshal/factory"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/builders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/interactors/nonceHandlerV2"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
	"github.com/klever-io/klv-bridge-eth-go/config"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	proposeTransferFuncName         = "proposeMultiTransferKdaBatch"
	proposeSetStatusFuncName        = "proposeKdaSafeSetCurrentTransactionBatchStatus"
	signFuncName                    = "sign"
	performActionFuncName           = "performAction"
	minClientAvailabilityAllowDelta = 1

	kleverDataGetterLogId = "KleverEth-KleverDataGetter"
)

// ClientArgs represents the argument for the NewClient constructor function
type ClientArgs struct {
	GasMapConfig                 config.KleverGasMapConfig
	Proxy                        proxy.Proxy
	Log                          logger.Logger
	RelayerPrivateKey            crypto.PrivateKey
	MultisigContractAddress      address.Address
	SafeContractAddress          address.Address
	IntervalToResendTxsInSeconds uint64
	TokensMapper                 TokensMapper
	RoleProvider                 roleProvider
	StatusHandler                bridgeCore.StatusHandler
	ClientAvailabilityAllowDelta uint64
}

// client represents the Klever Blockchain Client implementation
type client struct {
	*klvClientDataGetter
	txHandler                    txHandler
	tokensMapper                 TokensMapper
	relayerPublicKey             crypto.PublicKey
	relayerAddress               address.Address
	multisigContractAddress      address.Address
	safeContractAddress          address.Address
	log                          logger.Logger
	gasMapConfig                 config.KleverGasMapConfig
	addressPublicKeyConverter    bridgeCore.AddressConverter
	statusHandler                bridgeCore.StatusHandler
	clientAvailabilityAllowDelta uint64

	lastNonce                uint64
	retriesAvailabilityCheck uint64
	mut                      sync.RWMutex
}

// NewClient returns a new Klever Blockchain Client instance
func NewClient(args ClientArgs) (*client, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	argNonceHandler := nonceHandlerV2.ArgsNonceTransactionsHandlerV2{
		Proxy:            args.Proxy,
		IntervalToResend: time.Second * time.Duration(args.IntervalToResendTxsInSeconds),
	}
	nonceTxsHandler, err := nonceHandlerV2.NewNonceTransactionHandlerV2(argNonceHandler)
	if err != nil {
		return nil, err
	}

	publicKey := args.RelayerPrivateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	relayerAddress, err := address.NewAddressFromBytes(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("%w to get relayer address from bytes", err)
	}

	argsKLVClientDataGetter := ArgsKLVClientDataGetter{
		MultisigContractAddress: args.MultisigContractAddress,
		SafeContractAddress:     args.SafeContractAddress,
		RelayerAddress:          relayerAddress,
		Proxy:                   args.Proxy,
		Log:                     bridgeCore.NewLoggerWithIdentifier(logger.GetOrCreate(kleverDataGetterLogId), kleverDataGetterLogId),
	}
	getter, err := NewKLVClientDataGetter(argsKLVClientDataGetter)
	if err != nil {
		return nil, err
	}

	addressConverter, err := converters.NewAddressConverter()
	if err != nil {
		return nil, clients.ErrNilAddressConverter
	}

	bech23MultisigAddress := args.MultisigContractAddress.Bech32()

	bech23SafeAddress := args.SafeContractAddress.Bech32()

	internalMarshalizer, err := factory.NewMarshalizer(factory.ProtoMarshalizer)
	if err != nil {
		return nil, err
	}

	hasher, err := factoryHasher.NewHasher("blake2b")
	if err != nil {
		return nil, err
	}

	c := &client{
		txHandler: &transactionHandler{
			proxy:                   args.Proxy,
			relayerAddress:          relayerAddress,
			multisigAddressAsBech32: bech23MultisigAddress,
			nonceTxHandler:          nonceTxsHandler,
			relayerPrivateKey:       args.RelayerPrivateKey,
			singleSigner:            &singlesig.Ed25519Signer{},
			roleProvider:            args.RoleProvider,
			internalMarshalizer:     internalMarshalizer,
			hasher:                  hasher,
		},
		klvClientDataGetter:          getter,
		relayerPublicKey:             publicKey,
		relayerAddress:               relayerAddress,
		multisigContractAddress:      args.MultisigContractAddress,
		safeContractAddress:          args.SafeContractAddress,
		log:                          args.Log,
		gasMapConfig:                 args.GasMapConfig,
		addressPublicKeyConverter:    addressConverter,
		tokensMapper:                 args.TokensMapper,
		statusHandler:                args.StatusHandler,
		clientAvailabilityAllowDelta: args.ClientAvailabilityAllowDelta,
	}

	bech32RelayerAddress := relayerAddress.Bech32()
	c.log.Info("NewKleverBlockchainClient",
		"relayer address", bech32RelayerAddress,
		"multisig contract address", bech23MultisigAddress,
		"safe contract address", bech23SafeAddress)

	return c, nil
}

func checkArgs(args ClientArgs) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}
	if check.IfNil(args.RelayerPrivateKey) {
		return clients.ErrNilPrivateKey
	}
	if check.IfNil(args.MultisigContractAddress) {
		return fmt.Errorf("%w for the MultisigContractAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.SafeContractAddress) {
		return fmt.Errorf("%w for the SafeContractAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}
	if check.IfNil(args.TokensMapper) {
		return clients.ErrNilTokensMapper
	}
	if check.IfNil(args.RoleProvider) {
		return errNilRoleProvider
	}
	if check.IfNil(args.StatusHandler) {
		return clients.ErrNilStatusHandler
	}
	if args.ClientAvailabilityAllowDelta < minClientAvailabilityAllowDelta {
		return fmt.Errorf("%w for args.ClientAvailabilityAllowDelta, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.ClientAvailabilityAllowDelta, minClientAvailabilityAllowDelta)
	}
	err := checkGasMapValues(args.GasMapConfig)
	if err != nil {
		return err
	}
	return nil
}

func checkGasMapValues(gasMap config.KleverGasMapConfig) error {
	gasMapValue := reflect.ValueOf(gasMap)
	typeOfGasMapValue := gasMapValue.Type()

	for i := 0; i < gasMapValue.NumField(); i++ {
		fieldVal := gasMapValue.Field(i).Uint()
		if fieldVal == 0 {
			return fmt.Errorf("%w for field %s", errInvalidGasValue, typeOfGasMapValue.Field(i).Name)
		}
	}

	return nil
}

// GetPendingBatch returns the pending batch
func (c *client) GetPendingBatch(ctx context.Context) (*bridgeCore.TransferBatch, error) {
	c.log.Info("getting pending batch...")
	responseData, err := c.GetCurrentBatchAsDataBytes(ctx)
	if err != nil {
		return nil, err
	}

	if emptyResponse(responseData) {
		return nil, clients.ErrNoPendingBatchAvailable
	}

	return c.createPendingBatchFromResponse(ctx, responseData)
}

// GetBatch returns the batch (if existing)
func (c *client) GetBatch(ctx context.Context, batchID uint64) (*bridgeCore.TransferBatch, error) {
	c.log.Debug("getting batch", "ID", batchID)
	responseData, err := c.GetBatchAsDataBytes(ctx, batchID)
	if err != nil {
		return nil, err
	}

	if emptyResponse(responseData) {
		return nil, clients.ErrNoBatchAvailable
	}

	return c.createPendingBatchFromResponse(ctx, responseData)
}

func emptyResponse(response [][]byte) bool {
	return len(response) == 0 || (len(response) == 1 && len(response[0]) == 0)
}

func (c *client) createPendingBatchFromResponse(ctx context.Context, responseData [][]byte) (*bridgeCore.TransferBatch, error) {
	numFieldsForTransaction := 6
	dataLen := len(responseData)
	haveCorrectNumberOfArgs := (dataLen-1)%numFieldsForTransaction == 0 && dataLen > 1
	if !haveCorrectNumberOfArgs {
		return nil, fmt.Errorf("%w, got %d argument(s)", errInvalidNumberOfArguments, dataLen)
	}

	batchID, err := parseUInt64FromByteSlice(responseData[0])
	if err != nil {
		return nil, fmt.Errorf("%w while parsing batch ID", err)
	}

	batch := &bridgeCore.TransferBatch{
		ID: batchID,
	}

	cachedTokens := make(map[string][]byte)
	transferIndex := 0
	for i := 1; i < dataLen; i += numFieldsForTransaction {
		// blockNonce is the i-th element, let's ignore it for now
		depositNonce, errParse := parseUInt64FromByteSlice(responseData[i+1])
		if errParse != nil {
			return nil, fmt.Errorf("%w while parsing the deposit nonce, transfer index %d", errParse, transferIndex)
		}

		amount := big.NewInt(0).SetBytes(responseData[i+5])
		deposit := &bridgeCore.DepositTransfer{
			Nonce:            depositNonce,
			FromBytes:        responseData[i+2],
			DisplayableFrom:  c.addressPublicKeyConverter.ToBech32StringSilent(responseData[i+2]),
			ToBytes:          responseData[i+3],
			DisplayableTo:    c.addressPublicKeyConverter.ToHexStringWithPrefix(responseData[i+3]),
			SourceTokenBytes: responseData[i+4],
			DisplayableToken: string(responseData[i+4]),
			Amount:           amount,
		}

		storedConvertedTokenBytes, exists := cachedTokens[deposit.DisplayableToken]
		if !exists {
			deposit.DestinationTokenBytes, err = c.tokensMapper.ConvertToken(ctx, deposit.SourceTokenBytes)
			if err != nil {
				return nil, fmt.Errorf("%w while converting token bytes, transfer index %d", err, transferIndex)
			}
			cachedTokens[deposit.DisplayableToken] = deposit.DestinationTokenBytes
		} else {
			deposit.DestinationTokenBytes = storedConvertedTokenBytes
		}

		batch.Deposits = append(batch.Deposits, deposit)
		transferIndex++
	}

	batch.Statuses = make([]byte, len(batch.Deposits))

	c.log.Debug("created batch " + batch.String())

	return batch, nil
}

func (c *client) createCommonTxDataBuilder(funcName string, id int64) builders.TxDataBuilder {
	return builders.NewTxDataBuilder().Function(funcName).ArgInt64(id)
}

// ProposeSetStatus will trigger the proposal of the KDA safe set current transaction batch status operation
func (c *client) ProposeSetStatus(ctx context.Context, batch *bridgeCore.TransferBatch) (string, error) {
	if batch == nil {
		return "", clients.ErrNilBatch
	}

	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(proposeSetStatusFuncName, int64(batch.ID))
	for _, stat := range batch.Statuses {
		txBuilder.ArgBytes([]byte{stat})
	}

	gasLimit := c.gasMapConfig.ProposeStatusBase + uint64(len(batch.Deposits))*c.gasMapConfig.ProposeStatusForEach
	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, gasLimit)
	if err == nil {
		c.log.Info("proposed set statuses "+batch.String(), "transaction hash", hash)
	}

	return hash, err
}

// ProposeTransfer will trigger the propose transfer operation
func (c *client) ProposeTransfer(ctx context.Context, batch *bridgeCore.TransferBatch) (string, error) {
	if batch == nil {
		return "", clients.ErrNilBatch
	}

	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(proposeTransferFuncName, int64(batch.ID))

	for _, dt := range batch.Deposits {
		txBuilder.ArgBytes(dt.FromBytes).
			ArgBytes(dt.ToBytes).
			ArgBytes(dt.DestinationTokenBytes).
			ArgBigInt(dt.Amount).
			ArgInt64(int64(dt.Nonce)).
			ArgBytes(dt.Data)
	}

	gasLimit := c.gasMapConfig.ProposeTransferBase + uint64(len(batch.Deposits))*c.gasMapConfig.ProposeTransferForEach
	extraGasForScCalls := c.computeExtraGasForSCCallsBasic(batch, false)
	gasLimit += extraGasForScCalls
	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, gasLimit)
	if err == nil {
		c.log.Info("proposed transfer "+batch.String(), "transaction hash", hash)
	}

	return hash, err
}

// Sign will trigger the execution of a sign operation
func (c *client) Sign(ctx context.Context, actionID uint64) (string, error) {
	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(signFuncName, int64(actionID))

	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, c.gasMapConfig.Sign)
	if err == nil {
		c.log.Info("signed", "action ID", actionID, "transaction hash", hash)
	}

	return hash, err
}

// PerformAction will trigger the execution of the provided action ID
func (c *client) PerformAction(ctx context.Context, actionID uint64, batch *bridgeCore.TransferBatch) (string, error) {
	if batch == nil {
		return "", clients.ErrNilBatch
	}

	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(performActionFuncName, int64(actionID))

	gasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Statuses))*c.gasMapConfig.PerformActionForEach
	gasLimit += c.computeExtraGasForSCCallsBasic(batch, true)
	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, gasLimit)

	if err == nil {
		c.log.Info("performed action", "actionID", actionID, "transaction hash", hash)
	}

	return hash, err
}

func (c *client) computeExtraGasForSCCallsBasic(batch *bridgeCore.TransferBatch, performAction bool) uint64 {
	gasLimit := uint64(0)
	for _, deposit := range batch.Deposits {
		if bytes.Equal(deposit.Data, []byte{bridgeCore.MissingDataProtocolMarker}) {
			continue
		}

		computedLen := 1                     // extra argument separator (@)
		computedLen += len(deposit.Data) * 2 // the data is hexed, so, double the size

		gasLimit += uint64(computedLen) * c.gasMapConfig.ScCallPerByte
		if performAction {
			gasLimit += c.gasMapConfig.ScCallPerformForEach
		}
	}

	return gasLimit
}

func (c *client) checkIsPaused(ctx context.Context) error {
	isPaused, err := c.IsPaused(ctx)
	if err != nil {
		return fmt.Errorf("%w in client.ExecuteTransfer", err)
	}
	if isPaused {
		return fmt.Errorf("%w in client.ExecuteTransfer", clients.ErrMultisigContractPaused)
	}

	return nil
}

// IsMintBurnToken returns true if the provided token is whitelisted for mint/burn operations
func (c *client) IsMintBurnToken(ctx context.Context, token []byte) (bool, error) {
	return c.isMintBurnToken(ctx, token)
}

// IsNativeToken returns true if the provided token is native
func (c *client) IsNativeToken(ctx context.Context, token []byte) (bool, error) {
	return c.isNativeToken(ctx, token)
}

// TotalBalances returns the total stored tokens
func (c *client) TotalBalances(ctx context.Context, token []byte) (*big.Int, error) {
	return c.getTotalBalances(ctx, token)
}

// MintBalances returns the minted tokens
func (c *client) MintBalances(ctx context.Context, token []byte) (*big.Int, error) {
	return c.getMintBalances(ctx, token)
}

// BurnBalances returns the burned tokens
func (c *client) BurnBalances(ctx context.Context, token []byte) (*big.Int, error) {
	return c.getBurnBalances(ctx, token)
}

// CheckRequiredBalance will check the required balance for the provided token
func (c *client) CheckRequiredBalance(ctx context.Context, token []byte, value *big.Int) error {
	isMintBurn, err := c.IsMintBurnToken(ctx, token)
	if err != nil {
		return err
	}

	if isMintBurn {
		return nil
	}
	safeAddress := c.safeContractAddress.Bech32()

	kda, err := c.proxy.GetKDATokenData(ctx, c.safeContractAddress, string(token))
	if err != nil {
		return fmt.Errorf("%w for address %s for KDA token %s", err, safeAddress, string(token))
	}

	existingBalance, ok := big.NewInt(0).SetString(kda.Balance, 10)
	if !ok {
		return fmt.Errorf("%w for KDA token %s and address %s", errInvalidBalance, string(token), safeAddress)
	}

	if value.Cmp(existingBalance) > 0 {
		return fmt.Errorf("%w, existing: %s, required: %s for ERC20 token %s and address %s",
			errInsufficientKDABalance, existingBalance.String(), value.String(), string(token), safeAddress)
	}

	c.log.Debug("checked ERC20 balance",
		"KDA token", string(token),
		"address", safeAddress,
		"existing balance", existingBalance.String(),
		"needed", value.String())

	return nil
}

// CheckClientAvailability will check the client availability and will set the metric accordingly
func (c *client) CheckClientAvailability(ctx context.Context) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	currentNonce, err := c.GetCurrentNonce(ctx)
	if err != nil {
		c.setStatusForAvailabilityCheck(bridgeCore.Unavailable, err.Error(), currentNonce)

		return err
	}

	if currentNonce != c.lastNonce {
		c.retriesAvailabilityCheck = 0
		c.lastNonce = currentNonce
	}

	// if we reached this point we will need to increment the retries counter
	defer c.incrementRetriesAvailabilityCheck()

	if c.retriesAvailabilityCheck > c.clientAvailabilityAllowDelta {
		message := fmt.Sprintf("nonce %d fetched for %d times in a row", currentNonce, c.retriesAvailabilityCheck)
		c.setStatusForAvailabilityCheck(bridgeCore.Unavailable, message, currentNonce)

		return nil
	}

	c.setStatusForAvailabilityCheck(bridgeCore.Available, "", currentNonce)

	return nil
}

func (c *client) incrementRetriesAvailabilityCheck() {
	c.retriesAvailabilityCheck++
}

func (c *client) setStatusForAvailabilityCheck(status bridgeCore.ClientStatus, message string, nonce uint64) {
	c.statusHandler.SetStringMetric(bridgeCore.MetricKCClientStatus, status.String())
	c.statusHandler.SetStringMetric(bridgeCore.MetricLastKCClientError, message)
	c.statusHandler.SetIntMetric(bridgeCore.MetricLastBlockNonce, int(nonce))
}

// Close will close any started go routines. It returns nil.
func (c *client) Close() error {
	return c.txHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}
