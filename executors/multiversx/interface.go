package multiversx

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/parsers"
	"github.com/multiversx/mx-sdk-go/data"
)

// Proxy defines the behavior of a proxy able to serve MultiversX blockchain requests
type Proxy interface {
	GetNetworkConfig(ctx context.Context) (*models.NetworkConfig, error)
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	SendTransactions(ctx context.Context, txs []*transaction.Transaction) ([]string, error)
	ExecuteVMQuery(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error)
	GetAccount(ctx context.Context, address address.Address) (*models.Account, error)
	GetNetworkStatus(ctx context.Context, shardID uint32) (*data.NetworkStatus, error)
	GetShardOfAddress(ctx context.Context, bech32Address string) (uint32, error)
	GetESDTTokenData(ctx context.Context, address address.Address, tokenIdentifier string) (*data.ESDTFungibleTokenData, error)
	GetTransactionInfoWithResults(ctx context.Context, hash string) (*data.TransactionInfo, error)
	ProcessTransactionStatus(ctx context.Context, hexTxHash string) (transaction.Transaction_TXResult, error)
	IsInterfaceNil() bool
}

// NonceTransactionsHandler represents the interface able to handle the current nonce and the transactions resend mechanism
type NonceTransactionsHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, address address.Address, tx *transaction.Transaction) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	Close() error
	IsInterfaceNil() bool
}

// ScCallsExecuteFilter defines the operations supported by a filter that allows selective executions of batches
type ScCallsExecuteFilter interface {
	ShouldExecute(callData parsers.ProxySCCompleteCallData) bool
	IsInterfaceNil() bool
}

// Codec defines the operations implemented by a MultiversX codec
type Codec interface {
	DecodeProxySCCompleteCallData(buff []byte) (parsers.ProxySCCompleteCallData, error)
	ExtractGasLimitFromRawCallData(buff []byte) (uint64, error)
	IsInterfaceNil() bool
}
