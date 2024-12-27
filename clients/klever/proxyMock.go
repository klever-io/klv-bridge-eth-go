package klever

import (
	"context"

	"github.com/klever-io/klv-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-sdk-go/data"
)

// TODO: remove CreateMockProxyKLV when real proxy is implemented
// CreateMockProxyKLV returns a interactors.ProxyStub, to be used as a klever client proxy and
// be used to run the bridge while real proxy implementation is in development
func CreateMockProxyKLV(returningBytes [][]byte) *interactors.ProxyStub {
	return &interactors.ProxyStub{
		ExecuteVMQueryCalled:   executeVmQueryMock,
		GetNetworkStatusCalled: getNetworkStatusMock,
		GetNetworkConfigCalled: getNetworkConfigMock,
	}
}

func executeVmQueryMock(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	returningBytes := [][]byte{}
	return &data.VmValuesResponseData{
		Data: &vm.VMOutputApi{
			ReturnCode: okCodeAfterExecution,
			ReturnData: returningBytes,
		},
	}, nil
}

func getNetworkStatusMock(ctx context.Context, shardID uint32) (*data.NetworkStatus, error) {
	expectedNonce := uint64(0)

	return &data.NetworkStatus{
		Nonce: expectedNonce,
	}, nil
}

func getNetworkConfigMock(ctx context.Context) (*data.NetworkConfig, error) {
	chainID := "chain ID"
	minGasPrice := uint64(12234)
	minTxVersion := uint32(122)

	return &data.NetworkConfig{
		ChainID:               chainID,
		MinGasPrice:           minGasPrice,
		MinTransactionVersion: minTxVersion,
	}, nil
}
