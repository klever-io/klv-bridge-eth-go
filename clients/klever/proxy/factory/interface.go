package factory

import "github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"

// EndpointProvider is able to return endpoint routes strings
type EndpointProvider interface {
	GetNetworkConfig() string
	GetAccount(addressAsBech32 string) string
	GetEstimateTransactionFees() string
	GetSendTransaction() string
	GetSendMultipleTransactions() string
	GetTransactionStatus(hexHash string) string
	GetTransactionInfo(hexHash string) string
	GetVmValues() string
	GetNodeStatus() string
	GetRestAPIEntityType() models.RestAPIEntityType
	GetProcessedTransactionStatus(hexHash string) string
	GetKDATokenData(addressAsBech32 string, tokenIdentifier string) string
	IsInterfaceNil() bool
}
