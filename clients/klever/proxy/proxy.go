package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/factory"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	sdkHttp "github.com/klever-io/klv-bridge-eth-go/core/http"
)

const (
	withResultsQueryParam = "?withResults=true"
	withTxsAndLogs        = "?withTxs=true&withLogs=true"

	// MinAllowedDeltaToFinal is the minimum value between nonces allowed when checking finality on a shard
	MinAllowedDeltaToFinal = 1
)

var (
	// MaximumBlocksDelta is the maximum allowed delta between the final block and the current block
	MaximumBlocksDelta uint64 = 500
)

// ArgsProxy is the DTO used in the kc proxy constructor
type ArgsProxy struct {
	ProxyURL               string
	Client                 sdkHttp.Client
	SameScState            bool
	ShouldBeSynced         bool
	FinalityCheck          bool
	AllowedDeltaToFinal    int
	CacheExpirationTime    time.Duration
	EntityType             models.RestAPIEntityType
	FilterQueryBlockCacher BlockDataCache
}

// proxy implements basic functions for interacting with a kc Proxy
type proxy struct {
	*baseProxy
	sameScState            bool
	shouldBeSynced         bool
	finalityCheck          bool
	allowedDeltaToFinal    int
	filterQueryBlockCacher BlockDataCache
}

// NewProxy initializes and returns a proxy object
func NewProxy(args ArgsProxy) (*proxy, error) {
	err := checkArgsProxy(args)
	if err != nil {
		return nil, err
	}

	endpointProvider, err := factory.CreateEndpointProvider(args.EntityType)
	if err != nil {
		return nil, err
	}

	clientWrapper := sdkHttp.NewHttpClientWrapper(args.Client, args.ProxyURL)
	baseArgs := argsBaseProxy{
		httpClientWrapper: clientWrapper,
		expirationTime:    args.CacheExpirationTime,
		endpointProvider:  endpointProvider,
	}

	baseProxyInstance, err := newBaseProxy(baseArgs)
	if err != nil {
		return nil, err
	}

	cacher := args.FilterQueryBlockCacher
	if cacher == nil {
		cacher = &DisabledBlockDataCache{}
	}

	ep := &proxy{
		baseProxy:              baseProxyInstance,
		sameScState:            args.SameScState,
		shouldBeSynced:         args.ShouldBeSynced,
		finalityCheck:          args.FinalityCheck,
		allowedDeltaToFinal:    args.AllowedDeltaToFinal,
		filterQueryBlockCacher: cacher,
	}

	return ep, nil
}

func checkArgsProxy(args ArgsProxy) error {
	if args.FinalityCheck {
		if args.AllowedDeltaToFinal < MinAllowedDeltaToFinal {
			return fmt.Errorf("%w, provided: %d, minimum: %d",
				ErrInvalidAllowedDeltaToFinal, args.AllowedDeltaToFinal, MinAllowedDeltaToFinal)
		}
	}

	return nil
}

// ExecuteVMQuery retrieves data from existing SC trie through the use of a VM
func (ep *proxy) ExecuteVMQuery(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
	// checks if the address is valid before sending the request
	if _, err := address.NewAddress(vmRequest.Address); err != nil {
		return nil, err
	}

	jsonVMRequestWithOptionalParams := models.VmValueRequestWithOptionalParameters{
		VmValueRequest: vmRequest,
		SameScState:    ep.sameScState,
		ShouldBeSynced: ep.shouldBeSynced,
	}

	jsonVMRequest, err := json.Marshal(jsonVMRequestWithOptionalParams)
	if err != nil {
		return nil, err
	}

	buff, code, err := ep.PostHTTP(ctx, ep.endpointProvider.GetVmQuery(), jsonVMRequest)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	endpointProviderType := ep.endpointProvider.GetRestAPIEntityType()
	switch endpointProviderType {
	case models.Proxy:
		return ep.parseVmQueryProxy(buff)
	case models.ObserverNode:
		return ep.parseVmQueryNode(buff)
	}

	return &models.VmValuesResponseData{}, ErrInvalidEndpointProvider
}

func (ep *proxy) parseVmQueryNode(buff []byte) (*models.VmValuesResponseData, error) {
	response := &models.NodeResponseVmValue{}
	err := json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return &response.Data, nil
}

func (ep *proxy) parseVmQueryProxy(buff []byte) (*models.VmValuesResponseData, error) {
	response := &models.ProxyResponseVmValue{}
	err := json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return &models.VmValuesResponseData{
		Data: &response.Data,
	}, nil
}

// GetAccount retrieves an account info from the network (nonce, balance)
func (ep *proxy) GetAccount(ctx context.Context, address address.Address) (*models.Account, error) {
	if check.IfNil(address) {
		return nil, ErrNilAddress
	}
	if !address.IsValid() {
		return nil, ErrInvalidAddress
	}

	endpointProviderType := ep.endpointProvider.GetRestAPIEntityType()
	switch endpointProviderType {
	case models.Proxy:
		return ep.getAccountProxy(ctx, address)
	case models.ObserverNode:
		return ep.getAccountNode(ctx, address)
	}

	return &models.Account{}, ErrInvalidEndpointProvider
}

func (ep *proxy) getAccountNode(ctx context.Context, address address.Address) (*models.Account, error) {
	endpoint := ep.endpointProvider.GetAccount(address.Bech32())

	buff, code, err := ep.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.AccountNodeResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return &response.Data.AccountData, nil
}

func (ep *proxy) getAccountProxy(ctx context.Context, address address.Address) (*models.Account, error) {
	endpoint := ep.endpointProvider.GetAccount(address.Bech32())

	buff, code, err := ep.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.AccountApiResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return &models.Account{
		Address:  response.Data.AccountData.Address,
		Balance:  uint64(response.Data.AccountData.Balance),
		Nonce:    response.Data.AccountData.Nonce,
		RootHash: response.Data.AccountData.RootHash,
	}, nil
}

// SendTransaction broadcasts a transaction to the network and returns the txhash if successful
func (ep *proxy) SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error) {
	broadcast := models.BroadcastTransactionData{
		Tx: tx,
	}

	jsonTx, err := json.Marshal(broadcast)
	if err != nil {
		return "", err
	}

	buff, code, err := ep.PostHTTP(ctx, ep.endpointProvider.GetSendTransaction(), jsonTx)
	if err != nil || code != http.StatusOK {
		return "", createHTTPStatusError(code, err)
	}

	response := &models.SendTransactionResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return "", err
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}

	return response.Data.TxHash, nil
}

// SendTransactions broadcasts the provided transactions to the network and returns the txhashes if successful
func (ep *proxy) SendTransactions(ctx context.Context, txs []*transaction.Transaction) ([]string, error) {
	broadcast := models.BroadcastBulkTransactionData{
		Txs: txs,
	}

	jsonTx, err := json.Marshal(broadcast)
	if err != nil {
		return nil, err
	}

	buff, code, err := ep.PostHTTP(ctx, ep.endpointProvider.GetSendMultipleTransactions(), jsonTx)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.SendBulkTransactionsResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return response.Data.Hashes, nil
}

// GetTransactionStatus retrieves a transaction's status from the network
func (ep *proxy) GetTransactionStatus(ctx context.Context, hash string) (string, error) {
	endpoint := ep.endpointProvider.GetTransactionStatus(hash)
	buff, code, err := ep.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return "", createHTTPStatusError(code, err)
	}

	response := &models.TransactionStatus{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return "", err
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}

	return response.Data.Status, nil
}

// GetTransactionInfo retrieves a transaction's details from the network
func (ep *proxy) GetTransactionInfo(ctx context.Context, hash string) (*models.TransactionData, error) {
	return ep.getTransactionInfo(ctx, hash, false)
}

// GetTransactionInfoWithResults retrieves a transaction's details from the network with events
func (ep *proxy) GetTransactionInfoWithResults(ctx context.Context, hash string) (*models.TransactionData, error) {
	return ep.getTransactionInfo(ctx, hash, true)
}

func (ep *proxy) getTransactionInfo(ctx context.Context, hash string, withResults bool) (*models.TransactionData, error) {
	endpoint := ep.endpointProvider.GetTransactionInfo(hash)
	if withResults {
		endpoint += withResultsQueryParam
	}

	buff, code, err := ep.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.GetTransactionResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return &response.Data.Transaction, nil
}

// RequestTransactionCost retrieves how many gas a transaction will consume
func (ep *proxy) EstimateTransactionFees(ctx context.Context, tx *transaction.Transaction) (*transaction.FeesResponse, error) {
	jsonTx, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	buff, code, err := ep.PostHTTP(ctx, ep.endpointProvider.GetEstimateTransactionFees(), jsonTx)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.EstimateTransactionFeesResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return response.Data, nil
}

// GetKDATokenData returns the address' fungible token data
func (ep *proxy) GetKDATokenData(
	ctx context.Context,
	address address.Address,
	tokenIdentifier string,
) (*models.KDAFungibleTokenData, error) {
	if check.IfNil(address) {
		return nil, ErrNilAddress
	}
	if !address.IsValid() {
		return nil, ErrInvalidAddress
	}

	endpoint := ep.endpointProvider.GetKDATokenData(address.Bech32(), tokenIdentifier)
	buff, code, err := ep.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.KDAFungibleResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return response.Data.TokenData, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ep *proxy) IsInterfaceNil() bool {
	return ep == nil
}
