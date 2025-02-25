package endpointProviders

import "fmt"

const (
	networkConfig              = "network/config"
	account                    = "address/%s"
	estimateTransactionFees    = "transaction/estimate-fee"
	sendTransaction            = "transaction/send"
	sendMultipleTransactions   = "transaction/send-multiple"
	transactionStatus          = "transaction/%s/status"
	processedTransactionStatus = "transaction/%s/process-status"
	transactionInfo            = "transaction/%s"
	vmValues                   = "vm-values/query"
	kda                        = "address/%s/kda/%s"
)

type baseEndpointProvider struct{}

// GetNetworkConfig returns the network config endpoint
func (base *baseEndpointProvider) GetNetworkConfig() string {
	return networkConfig
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

// GetProcessedTransactionStatus returns the transaction status endpoint
func (base *baseEndpointProvider) GetProcessedTransactionStatus(hexHash string) string {
	return fmt.Sprintf(processedTransactionStatus, hexHash)
}

// GetTransactionInfo returns the transaction info endpoint
func (base *baseEndpointProvider) GetTransactionInfo(hexHash string) string {
	return fmt.Sprintf(transactionInfo, hexHash)
}

// GetVmValues returns the VM values endpoint
func (base *baseEndpointProvider) GetVmValues() string {
	return vmValues
}
