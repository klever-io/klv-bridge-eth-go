package models

import (
	idata "github.com/klever-io/klever-go/indexer/data"
)

// Account defines the data structure for an account
type Account struct {
	*idata.AccountInfo
	Assets map[string]*idata.AccountKDA `json:"assets"`
}

// ResponseAccount follows the format of the data field of an account response
type ResponseAccount struct {
	AccountData Account `json:"account"`
}

// AccountApiResponse defines a wrapped account that the node respond with
type AccountApiResponse struct {
	Data  ResponseAccount `json:"data"`
	Error string          `json:"error"`
	Code  string          `json:"code"`
}
