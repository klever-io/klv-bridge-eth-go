package models

import (
	idata "github.com/klever-io/klever-go/indexer/data"
)

// Account defines the data structure for an account received from the node
type Account struct {
	Address  string `json:"address"`
	RootHash string `json:"rootHash"`
	Balance  uint64 `json:"balance"`
	Nonce    uint64 `json:"nonce"`
}

// ResponseNodeAccount follows the format of the data field of an account response
type ResponseNodeAccount struct {
	AccountData Account `json:"account"`
}

// AccountNodeResponse defines a wrapped account that the node respond with
type AccountNodeResponse struct {
	Data  ResponseNodeAccount `json:"data"`
	Error string              `json:"error"`
	Code  string              `json:"code"`
}

// ProxyAccountData defines the data structure for an account received from proxy
type ProxyAccountData struct {
	*idata.AccountInfo
	Assets map[string]*idata.AccountKDA `json:"assets"`
}

// ResponseProxyAccount follows the format of the data field of an account response
type ResponseProxyAccount struct {
	AccountData ProxyAccountData `json:"account"`
}

// AccountApiResponse defines a wrapped account that the proxy respond with
type AccountApiResponse struct {
	Data  ResponseProxyAccount `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}
