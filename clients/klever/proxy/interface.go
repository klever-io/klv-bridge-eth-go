package proxy

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

// Proxy defines the behavior of a proxy able to serve MultiversX blockchain requests
type Proxy interface {
	GetNetworkConfig(ctx context.Context) (*models.NetworkConfig, error)
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	SendTransactions(ctx context.Context, txs []*transaction.Transaction) ([]string, error)
	ExecuteVMQuery(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error)
	GetAccount(ctx context.Context, address address.Address) (*models.Account, error)
	GetNetworkStatus(ctx context.Context) (*models.NodeOverview, error)
	GetKDATokenData(ctx context.Context, address address.Address, tokenIdentifier string) (*models.KDAFungibleTokenData, error)
	GetTransactionInfoWithResults(ctx context.Context, hash string) (*models.TransactionData, error)
	ProcessTransactionStatus(ctx context.Context, hexTxHash string) (transaction.Transaction_TXResult, error)
	EstimateTransactionFees(ctx context.Context, txs *transaction.Transaction) (*transaction.FeesResponse, error)
	IsInterfaceNil() bool
}

type httpClientWrapper interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
	PostHTTP(ctx context.Context, endpoint string, data []byte) ([]byte, int, error)
	IsInterfaceNil() bool
}

// BlockDataCache defines the methods required for a basic cache.
type BlockDataCache interface {
	Get(key []byte) (value interface{}, ok bool)
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	IsInterfaceNil() bool
}
