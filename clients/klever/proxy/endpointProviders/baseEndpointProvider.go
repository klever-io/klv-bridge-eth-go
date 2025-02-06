package endpointProviders

import "fmt"

const (
	networkConfig              = "network/config"
	networkEconomics           = "network/economics"
	ratingsConfig              = "network/ratings"
	enableEpochsConfig         = "network/enable-epochs"
	account                    = "address/%s"
	estimateTransactionFees    = "transaction/estimate-fee"
	sendTransaction            = "transaction/send"
	sendMultipleTransactions   = "transaction/send-multiple"
	transactionStatus          = "transaction/%s/status"
	processedTransactionStatus = "transaction/%s/process-status"
	transactionInfo            = "transaction/%s"
	vmValues                   = "vm-values/query"
	genesisNodesConfig         = "network/genesis-nodes"
	rawStartOfEpochMetaBlock   = "internal/raw/startofepoch/metablock/by-epoch/%d"
	rawStartOfEpochValidators  = "internal/json/startofepoch/validators/by-epoch/%d"
	esdt                       = "address/%s/esdt/%s"
	nft                        = "address/%s/nft/%s/nonce/%d"
)

type baseEndpointProvider struct{}

// GetNetworkConfig returns the network config endpoint
func (base *baseEndpointProvider) GetNetworkConfig() string {
	return networkConfig
}

// GetNetworkEconomics returns the network economics endpoint
func (base *baseEndpointProvider) GetNetworkEconomics() string {
	return networkEconomics
}

// GetEnableEpochsConfig returns the enable epochs config endpoint
func (base *baseEndpointProvider) GetEnableEpochsConfig() string {
	return enableEpochsConfig
}

// GetAccount returns the account endpoint
func (base *baseEndpointProvider) GetAccount(addressAsBech32 string) string {
	return fmt.Sprintf(account, addressAsBech32)
}

// GetESDTTokenData returns the esdt endpoint
func (base *baseEndpointProvider) GetESDTTokenData(addressAsBech32 string, tokenIdentifier string) string {
	return fmt.Sprintf(esdt, addressAsBech32, tokenIdentifier)
}

// GetNFTTokenData returns the NFT/SFT/MetaESDT endpoint
func (base *baseEndpointProvider) GetNFTTokenData(addressAsBech32 string, tokenIdentifier string, nonce uint64) string {
	return fmt.Sprintf(nft, addressAsBech32, tokenIdentifier, nonce)
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
