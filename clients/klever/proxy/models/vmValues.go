package models

import (
	"github.com/klever-io/klever-go/data/vm"
)

// VmValuesResponseData follows the format of the data field in an API response for a VM values query
type VmValuesResponseData struct {
	Data *vm.VMOutputApi `json:"data"`
}

// ResponseVmValue defines a wrapper over string containing returned data in hex format
type ResponseVmValue struct {
	Data  VmValuesResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

// VmValueRequest represents the structure on which user input for generating a new transaction will validate against
type VmValueRequest struct {
	ScAddress  string           `json:"scAddress"`
	FuncName   string           `json:"funcName"`
	CallerAddr string           `json:"caller"`
	CallValue  map[string]int64 `json:"value"`
	Args       []string         `json:"args"`
}

// VmValueRequestWithOptionalParameters defines the request struct for values available in a VM
type VmValueRequestWithOptionalParameters struct {
	*VmValueRequest
	SameScState    bool `json:"sameScState"`
	ShouldBeSynced bool `json:"shouldBeSynced"`
}
