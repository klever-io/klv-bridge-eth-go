package models

import (
	"github.com/klever-io/klever-go/data/vm"
)

// VmValuesResponseData follows the format of the data field in an API response for a VM values query
type VmValuesResponseData struct {
	Data *vm.VMOutputApi `json:"data"`
}

// NodeResponseVmValue defines a wrapper over string containing returned data in hex format from node
type NodeResponseVmValue struct {
	Data  VmValuesResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

// ProxyResponseVmValue defines a wrapper over string containing returned data in hex format from proxy
type ProxyResponseVmValue struct {
	Data  vm.VMOutputApi `json:"data"`
	Error string         `json:"error"`
	Code  string         `json:"code"`
}

// TODO: check if needed since there is one similar in klever-io/klever-go/network/api/vm/vmValuesGroup.go and also
// check if the struct needs to be separeted with VmValueRequestWithOptionalParameters
// VmValueRequest defines the request struct for values available in a VM
type VmValueRequest struct {
	Address    string           `json:"scAddress"`
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
