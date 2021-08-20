package elrond

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

const (
	SignCost              = 35_000_000
	ProposeTransferCost   = 35_000_000
	ProposeTransferTxCost = 15_000_000
	ProposeStatusCost     = 50_000_000
	PerformActionCost     = 60_000_000
	PerformActionTxCost   = 20_000_000
	GetNextTxBatchCost    = 250_000_000
	nonceUpdateInterval   = time.Minute
)

const (
	NoRights = iota
	CanPropose
	CanProposeAndSign
)

type QueryResponseErr struct {
	code    string
	message string
}

func (e QueryResponseErr) Error() string {
	return fmt.Sprintf("Got response code %q and message %q", e.code, e.message)
}

type elrondProxy interface {
	GetNetworkConfig() (*data.NetworkConfig, error)
	SendTransaction(*data.Transaction) (string, error)
	GetTransactionInfoWithResults(hash string) (*data.TransactionInfo, error)
	ExecuteVMQuery(vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
}

type Client struct {
	proxy         elrondProxy
	bridgeAddress string
	privateKey    []byte
	address       core.AddressHandler
	nonce         uint64
	log           logger.Logger
}

func NewClient(config bridge.Config) (*Client, string, error) {
	log := logger.GetOrCreate("ElrondClient")

	proxy := blockchain.NewElrondProxy(config.NetworkAddress, nil)
	wallet := interactors.NewWallet()

	privateKey, err := wallet.LoadPrivateKeyFromPemFile(config.PrivateKey)
	if err != nil {
		return nil, "", err
	}

	address, err := wallet.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, "", err
	}

	log.Info("Elrond: NewClient", "address", address.AddressAsBech32String())

	client := &Client{
		proxy:         proxy,
		bridgeAddress: config.BridgeAddress,
		privateKey:    privateKey,
		address:       address,
		log: log,
	}

	go func() {
		for {
			account, err := proxy.GetAccount(address)
			if err == nil {
				client.nonce = account.Nonce
			}
			time.Sleep(nonceUpdateInterval)
		}
	}()

	return client, addressString, nil
}

func (c *Client) GetPending(context.Context) *bridge.Batch {
	c.log.Info("Elrond: Getting pending batch")
	responseData, err := c.getCurrentBatch()
	if err != nil {
		c.log.Error(fmt.Sprintf("Error querying current batch: %q", err.Error()))
		return nil
	}

	if emptyResponse(responseData) {
		_, err := c.getNextPendingBatch()
		if err != nil {
			c.log.Error(fmt.Sprintf("Error retrieving next pending batch %q", err.Error()))
			return nil
		}
	}

	responseData, err = c.getCurrentBatch()
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	if emptyResponse(responseData) {
		return nil
	}

	addrPkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
	var transactions []*bridge.DepositTransaction
	for i := 1; i < len(responseData); i += 6 {
		amount := new(big.Int).SetBytes(responseData[i+5])
		blockNonce, err := strconv.ParseInt(hex.EncodeToString(responseData[i]), 16, 64)
		if err != nil {
			c.log.Error(err.Error())
			return nil
		}
		depositNonce, err := strconv.ParseInt(hex.EncodeToString(responseData[i+1]), 16, 64)
		if err != nil {
			c.log.Error(err.Error())
			return nil
		}

		tx := &bridge.DepositTransaction{
			To:           fmt.Sprintf("0x%s", hex.EncodeToString(responseData[i+3])),
			From:         addrPkConv.Encode(responseData[i+2]),
			TokenAddress: fmt.Sprintf("0x%s", hex.EncodeToString(responseData[i+4])),
			Amount:       amount,
			DepositNonce: bridge.NewNonce(depositNonce),
			BlockNonce:   bridge.NewNonce(blockNonce),
			Status:       0,
			Error:        nil,
		}
		transactions = append(transactions, tx)
	}

	batchId, err := strconv.ParseInt(hex.EncodeToString(responseData[0]), 16, 64)
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	return &bridge.Batch{
		Id:           bridge.NewBatchId(batchId),
		Transactions: transactions,
	}
}

func (c *Client) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	builder := newBuilder().
		Func("proposeEsdtSafeSetCurrentTransactionBatchStatus").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		builder = builder.Int(big.NewInt(int64(tx.Status)))
	}

	hash, err := c.sendTransaction(builder, ProposeStatusCost)
	if err != nil {
		c.log.Error(err.Error())
	}
	c.log.Info(fmt.Sprintf("Elrond: Proposed status update with hash %s", hash))
}

func (c *Client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	builder := newBuilder().
		Func("proposeMultiTransferEsdtBatch").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		builder = builder.
			Address(tx.To).
			HexString(c.GetTokenId(tx.TokenAddress[2:])).
			BigInt(tx.Amount)
	}

	hash, err := c.sendTransaction(builder, uint64(ProposeTransferCost+len(batch.Transactions)*ProposeTransferTxCost))

	if err == nil {
		c.log.Info(fmt.Sprintf("Elrond: Proposed transfer for batch %v with hash %s", batch.Id, hash))
	} else {
		c.log.Error(fmt.Sprintf("Elrond: Propose transfer errored with: %q", err.Error()))
	}

	return hash, err
}

func (c *Client) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("wasTransferActionProposed").
		BatchId(batch.Id).
		WithTx(batch, c.GetTokenId).
		Build()

	return c.executeBoolQuery(valueRequest)
}

func (c *Client) GetActionIdForProposeTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("getActionIdForTransferBatch").
		BatchId(batch.Id).
		WithTx(batch, c.GetTokenId).
		Build()

	response, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return bridge.NewActionId(0)
	}

	actionId := bridge.NewActionId(int64(response))

	c.log.Info(fmt.Sprintf("Elrond: got actionId %v for batchId %v", actionId, batch.Id))

	return actionId
}

func (c *Client) WasProposedSetStatus(_ context.Context, batch *bridge.Batch) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("wasSetCurrentTransactionBatchStatusActionProposed").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		valueRequest = valueRequest.BigInt(big.NewInt(int64(tx.Status)))
	}

	return c.executeBoolQuery(valueRequest.Build())
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("getActionIdForSetCurrentTransactionBatchStatus").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		valueRequest = valueRequest.BigInt(big.NewInt(int64(tx.Status)))
	}

	response, err := c.executeUintQuery(valueRequest.Build())
	if err != nil {
		c.log.Error(err.Error())
		return bridge.NewActionId(0)
	}

	return bridge.NewActionId(int64(response))
}

func (c *Client) WasExecuted(_ context.Context, actionId bridge.ActionId, _ bridge.BatchId) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("wasActionExecuted").
		ActionId(actionId).
		Build()

	result := c.executeBoolQuery(valueRequest)

	if result {
		c.log.Info(fmt.Sprintf("Elrond: ActionId %v was executed", actionId))
	}

	return result
}

func (c *Client) Sign(_ context.Context, actionId bridge.ActionId) (string, error) {
	builder := newBuilder().
		Func("sign").
		ActionId(actionId)

	hash, err := c.sendTransaction(builder, SignCost)

	if err == nil {
		c.log.Info(fmt.Sprintf("Elrond: Singed with hash %q", hash))
	} else {
		c.log.Error(fmt.Sprintf("Elrond: Sign failed with %q;", err.Error()))
	}

	return hash, err
}

func (c *Client) Execute(_ context.Context, actionId bridge.ActionId, batch *bridge.Batch) (string, error) {
	builder := newBuilder().
		Func("performAction").
		ActionId(actionId)

	hash, err := c.sendTransaction(builder, uint64(PerformActionCost+len(batch.Transactions)*PerformActionTxCost))

	if err == nil {
		c.log.Info(fmt.Sprintf("Elrond: Executed actionId %v with hash %s", actionId, hash))
	} else {
		c.log.Error(fmt.Sprintf("Elrond: Executed failed with %q;", err.Error()))
	}

	return hash, err
}

func (c *Client) SignersCount(_ context.Context, actionId bridge.ActionId) uint {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("getActionSignerCount").
		ActionId(actionId).
		Build()

	count, _ := c.executeUintQuery(valueRequest)
	return uint(count)
}

// Mapper

func (c *Client) GetTokenId(address string) string {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("getTokenIdForErc20Address").
		HexString(address).
		Build()

	tokenId, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

	return tokenId
}

func (c *Client) GetErc20Address(tokenId string) string {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("getErc20AddressForTokenId").
		HexString(tokenId).
		Build()

	address, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

	return address
}

// RoleProvider

func (c *Client) IsWhitelisted(address string) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("userRole").
		HexString(address).
		Build()

	role, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	return role == CanProposeAndSign
}

// Helpers

func (c *Client) executeQuery(valueRequest *data.VmValueRequest) ([][]byte, error) {
	response, err := c.proxy.ExecuteVMQuery(valueRequest)
	if err != nil {
		return nil, err
	}

	if response.Data.ReturnCode != "ok" {
		return nil, QueryResponseErr{response.Data.ReturnCode, response.Data.ReturnMessage}
	}

	return response.Data.ReturnData, nil
}

func (c *Client) executeBoolQuery(valueRequest *data.VmValueRequest) bool {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	if len(responseData[0]) == 0 {
		return false
	}

	result, err := strconv.ParseBool(fmt.Sprintf("%d", responseData[0][0]))
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	return result
}

func (c *Client) executeUintQuery(valueRequest *data.VmValueRequest) (uint64, error) {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		return 0, err
	}

	if len(responseData[0]) == 0 {
		return 0, err
	}

	result, err := strconv.ParseUint(hex.EncodeToString(responseData[0]), 16, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *Client) executeStringQuery(valueRequest *data.VmValueRequest) (string, error) {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		return "", err
	}

	if len(responseData[0]) == 0 {
		return "", err
	}

	return fmt.Sprintf("%x", responseData[0]), nil
}

func (c *Client) signTransaction(builder *txDataBuilder, cost uint64) (*data.Transaction, error) {
	networkConfig, err := c.proxy.GetNetworkConfig()
	if err != nil {
		return nil, err
	}

	nonce := c.nonce
	if err != nil {
		return nil, err
	}

	tx := &data.Transaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: cost,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    nonce,
		Data:     builder.ToBytes(),
		SndAddr:  c.address.AddressAsBech32String(),
		RcvAddr:  c.bridgeAddress,
		Value:    "0",
	}

	err = c.signTransactionWithPrivateKey(tx, c.privateKey)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// signTransactionWithPrivateKey signs a transaction with the provided private key
// TODO use the transaction interactor for signing and sending transactions
func (c *Client) signTransactionWithPrivateKey(tx *data.Transaction, privateKey []byte) error {
	tx.Signature = ""
	txSingleSigner := &singlesig.Ed25519Signer{}
	suite := ed25519.NewEd25519()
	keyGen := signing.NewKeyGenerator(suite)
	txSignPrivKey, err := keyGen.PrivateKeyFromByteArray(privateKey)
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(&tx)
	if err != nil {
		return err
	}
	signature, err := txSingleSigner.Sign(txSignPrivKey, bytes)
	if err != nil {
		return err
	}
	tx.Signature = hex.EncodeToString(signature)

	return nil
}

func (c *Client) sendTransaction(builder *txDataBuilder, cost uint64) (string, error) {
	tx, err := c.signTransaction(builder, cost)
	if err != nil {
		return "", err
	}

	hash, err := c.proxy.SendTransaction(tx)
	if err == nil {
		c.nonce++
	}

	return hash, err
}

func (c *Client) getCurrentBatch() ([][]byte, error) {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String()).
		Func("getCurrentTxBatch").
		Build()

	return c.executeQuery(valueRequest)
}

func (c *Client) getNextPendingBatch() (string, error) {
	builder := newBuilder().
		Func("getNextTransactionBatch")

	return c.sendTransaction(builder, GetNextTxBatchCost)
}

func emptyResponse(response [][]byte) bool {
	return len(response) == 0 || (len(response) == 1 && len(response[0]) == 0)
}

// Builders

type valueRequestBuilder struct {
	address    string
	funcName   string
	callerAddr string
	args       []string
}

func newValueBuilder(address, callerAddr string) *valueRequestBuilder {
	return &valueRequestBuilder{
		address:    address,
		callerAddr: callerAddr,
		args:       []string{},
	}
}

func (builder *valueRequestBuilder) Build() *data.VmValueRequest {
	return &data.VmValueRequest{
		Address:    builder.address,
		FuncName:   builder.funcName,
		CallerAddr: builder.callerAddr,
		Args:       builder.args,
	}
}

func (builder *valueRequestBuilder) Func(functionName string) *valueRequestBuilder {
	builder.funcName = functionName

	return builder
}

func (builder *valueRequestBuilder) Nonce(nonce bridge.Nonce) *valueRequestBuilder {
	return builder.BigInt(nonce)
}

func (builder *valueRequestBuilder) BatchId(batchId bridge.BatchId) *valueRequestBuilder {
	return builder.BigInt(batchId)
}

func (builder *valueRequestBuilder) ActionId(actionId bridge.ActionId) *valueRequestBuilder {
	return builder.BigInt(actionId)
}

func (builder *valueRequestBuilder) BigInt(value *big.Int) *valueRequestBuilder {
	builder.args = append(builder.args, intToHex(value))

	return builder
}

func (builder *valueRequestBuilder) HexString(value string) *valueRequestBuilder {
	builder.args = append(builder.args, value)

	return builder
}

func (builder *valueRequestBuilder) Address(value string) *valueRequestBuilder {
	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
	buff, _ := pkConv.Decode(value)
	builder.args = append(builder.args, hex.EncodeToString(buff))

	return builder
}

func (builder *valueRequestBuilder) WithTx(batch *bridge.Batch, mapper func(string) string) *valueRequestBuilder {
	for _, tx := range batch.Transactions {
		builder = builder.
			Address(tx.To).
			HexString(mapper(tx.TokenAddress[2:])).
			BigInt(tx.Amount)
	}

	return builder
}

type txDataBuilder struct {
	function  string
	elements  []string
	separator string
}

func newBuilder() *txDataBuilder {
	return &txDataBuilder{
		function:  "",
		elements:  make([]string, 0),
		separator: "@",
	}
}

func (builder *txDataBuilder) Func(function string) *txDataBuilder {
	builder.function = function

	return builder
}

func (builder *txDataBuilder) ActionId(value bridge.ActionId) *txDataBuilder {
	return builder.Int(value)
}

func (builder *txDataBuilder) BatchId(value bridge.BatchId) *txDataBuilder {
	return builder.Int(value)
}

func (builder *txDataBuilder) Nonce(nonce bridge.Nonce) *txDataBuilder {
	return builder.Int(nonce)
}

func (builder *txDataBuilder) Int(value *big.Int) *txDataBuilder {
	builder.elements = append(builder.elements, intToHex(value))

	return builder
}

func (builder *txDataBuilder) BigInt(value *big.Int) *txDataBuilder {
	builder.elements = append(builder.elements, hex.EncodeToString(value.Bytes()))

	return builder
}

func (builder *txDataBuilder) Address(value string) *txDataBuilder {
	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
	buff, _ := pkConv.Decode(value)
	builder.elements = append(builder.elements, hex.EncodeToString(buff))

	return builder
}

func (builder *txDataBuilder) HexString(value string) *txDataBuilder {
	builder.elements = append(builder.elements, value)

	return builder
}

func (builder *txDataBuilder) ToString() string {
	result := builder.function
	for _, element := range builder.elements {
		result = result + builder.separator + element
	}

	return result
}

func (builder *txDataBuilder) ToBytes() []byte {
	return []byte(builder.ToString())
}

func intToHex(value *big.Int) string {
	return hex.EncodeToString(value.Bytes())
}
