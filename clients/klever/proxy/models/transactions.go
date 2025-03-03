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

// SendTransactionData holds the data of a transaction sent to the network
type SendTransactionData struct {
	TxHash  string `json:"txHash"`
	TxCount int    `json:"txCount"`
}

type GenericResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  string      `json:"code"`
}

// SendTransactionResponse holds the response received from the network when broadcasting a transaction
type SendTransactionResponse struct {
	Data  *SendTransactionData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

// EstimateTransactionFeesResponse defines the structure of responses on EstimateTransactionFees API endpoint
type EstimateTransactionFeesResponse struct {
	Data  *transaction.FeesResponse `json:"fees"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

// KDAFungibleResponse holds the KDA (fungible) token data endpoint response
type KDAFungibleResponse struct {
	Data struct {
		TokenData *KDAFungibleTokenData `json:"tokenData"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

// KDAFungibleTokenData holds the KDA (fungible) token data definition
type KDAFungibleTokenData struct {
	TokenIdentifier string `json:"tokenIdentifier"`
	Balance         string `json:"balance"`
	Properties      string `json:"properties"`
}

// TransactionData represents the structure that maps and validates user input for publishing a new transaction
type TransactionData struct {
	*idata.Transaction
}

// GetTransactionResponseData follows the format of the data field of get transaction response
type GetTransactionResponseData struct {
	Transaction TransactionData `json:"transaction"`
}

// GetTransactionResponse defines a response from the node holding the transaction sent from the chain
type GetTransactionResponse struct {
	Data  GetTransactionResponseData `json:"data"`
	Error string                     `json:"error"`
	Code  string                     `json:"code"`
}

// TransactionStatus holds a transaction's status response from the network
type TransactionStatus struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
