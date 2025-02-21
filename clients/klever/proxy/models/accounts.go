package models

import (
	idata "github.com/klever-io/klever-go/indexer/data"
)

// Account defines the data structure for an account
type Account struct {
	*idata.AccountInfo
	Assets map[string]*idata.AccountKDA `json:"assets"`
}
