package models

import "github.com/klever-io/klever-go/data/transaction"

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
