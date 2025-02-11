package models

import (
	"github.com/klever-io/klever-go/data/transaction"
	idata "github.com/klever-io/klever-go/indexer/data"
)

// TxHashes represents a colection of hashs of each transaction returned by a SendBulkTransactions
type TxHashes []string

// SendBulkTransactionsResponse defines the structure of responses on SendBulkTransactions API endpoint
type SendBulkTransactionsResponse struct {
	Data  TxHashes `json:"txHashes"`
	Error string   `json:"error"`
	Code  string   `json:"code"`
}

// EstimateTransactionFeesResponse defines the structure of responses on EstimateTransactionFees API endpoint
type EstimateTransactionFeesResponse struct {
	Data  *transaction.FeesResponse `json:"txHashes"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

// Transaction represents the structure that maps and validates user input for publishing a new transaction
type Transaction struct {
	*idata.Transaction
}

// GetTransactionResponseData follows the format of the data field of get transaction response
type GetTransactionResponseData struct {
	Transaction Transaction `json:"transaction"`
}

// GetTransactionResponse defines a response from the node holding the transaction sent from the chain
type GetTransactionResponse struct {
	Data  GetTransactionResponseData `json:"data"`
	Error string                     `json:"error"`
	Code  string                     `json:"code"`
}
