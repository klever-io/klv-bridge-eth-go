package factory

import "github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"

// EndpointProvider is able to return endpoint routes strings
type EndpointProvider interface {
	GetNetworkConfig() string
	GetNetworkEconomics() string
	GetAccount(addressAsBech32 string) string
	GetEstimateTransactionFees() string
	GetSendTransaction() string
	GetSendMultipleTransactions() string
	GetTransactionStatus(hexHash string) string
	GetTransactionInfo(hexHash string) string
	GetVmValues() string
	GetNodeStatus(shardID uint32) string
	GetRestAPIEntityType() models.RestAPIEntityType
	GetProcessedTransactionStatus(hexHash string) string
	GetESDTTokenData(addressAsBech32 string, tokenIdentifier string) string
	IsInterfaceNil() bool
}
