package endpointProviders

import "fmt"

const (
	networkConfig            = "network/config"
	nodeStatus               = "node/overview"
	account                  = "address/%s"
	estimateTransactionFees  = "transaction/estimate-fee"
	sendTransaction          = "transaction/broadcast"
	sendMultipleTransactions = "transaction/broadcast"
	transactionStatus        = "transaction/%s/status"
	transactionInfo          = "transaction/%s"
	kda                      = "address/%s/kda/%s"
)

type baseEndpointProvider struct{}

// GetNetworkConfig returns the network config endpoint
func (base *baseEndpointProvider) GetNetworkConfig() string {
	return networkConfig
}

// GetNodeStatus returns the network status endpoint
func (base *baseEndpointProvider) GetNodeStatus() string {
	return nodeStatus
}

// GetAccount returns the account endpoint
func (base *baseEndpointProvider) GetAccount(addressAsBech32 string) string {
	return fmt.Sprintf(account, addressAsBech32)
}

// GetKDATokenData returns the kda endpoint
func (base *baseEndpointProvider) GetKDATokenData(addressAsBech32 string, tokenIdentifier string) string {
	return fmt.Sprintf(kda, addressAsBech32, tokenIdentifier)
}

// GetCostTransaction returns the transaction cost endpoint
func (base *baseEndpointProvider) GetEstimateTransactionFees() string {
	return estimateTransactionFees
}

// GetSendTransaction returns the send transaction endpoint
func (base *baseEndpointProvider) GetSendTransaction() string {
	return sendTransaction
}

// GetSendMultipleTransactions returns the send multiple transactions endpoint
func (base *baseEndpointProvider) GetSendMultipleTransactions() string {
	return sendMultipleTransactions
}

// GetTransactionStatus returns the transaction status endpoint
func (base *baseEndpointProvider) GetTransactionStatus(hexHash string) string {
	return fmt.Sprintf(transactionStatus, hexHash)
}

// GetTransactionInfo returns the transaction info endpoint
func (base *baseEndpointProvider) GetTransactionInfo(hexHash string) string {
	return fmt.Sprintf(transactionInfo, hexHash)
}
